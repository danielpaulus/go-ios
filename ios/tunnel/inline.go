package tunnel

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/Masterminds/semver"
	"github.com/danielpaulus/go-ios/ios"
	"github.com/quic-go/quic-go"
	log "github.com/sirupsen/logrus"
)

const (
	// InlineTunnelTimeout is the default timeout for creating an inline tunnel
	InlineTunnelTimeout = 30 * time.Second
)

// InlineTunnel represents a per-command tunnel that manages its own lifecycle.
// It creates a tunnel on-demand, executes the command, and tears down the tunnel automatically.
type InlineTunnel struct {
	Tunnel
	LocalPort int
}

// Close tears down the tunnel and releases all resources
func (t *InlineTunnel) Close() error {
	return t.Tunnel.Close()
}

// CreateInlineTunnel creates an authenticated tunnel to the device.
// It does not require sudo or a running tunnel agent.
// The pairRecordPath is used for iOS 17.0-17.3 (manual pairing); iOS 17.4+ uses lockdown.
func CreateInlineTunnel(ctx context.Context, udid string, pairRecordPath string) (*InlineTunnel, error) {
	device, err := ios.GetDevice(udid)
	if err != nil {
		return nil, fmt.Errorf("CreateInlineTunnel: failed to get device: %w", err)
	}

	version, err := ios.GetProductVersion(device)
	if err != nil {
		return nil, fmt.Errorf("CreateInlineTunnel: failed to get device version: %w", err)
	}
	if version.LessThan(ios.IOS17()) {
		return nil, fmt.Errorf("CreateInlineTunnel: inline tunnel requires iOS 17+, device has iOS %s", version.String())
	}

	// iOS 17.4+ - use lockdown-based userspace tunnel (no pair records needed)
	semver174 := semver.MustParse("17.4.0")
	if version.GreaterThan(semver174) || version.Equal(semver174) {
		return createInlineTunnelLockdown(ctx, device)
	}

	// iOS 17.0 - 17.3 - use manual pairing via mDNS
	if pairRecordPath == "" {
		pairRecordPath = "."
	}
	pm, err := NewPairRecordManager(pairRecordPath)
	if err != nil {
		return nil, fmt.Errorf("CreateInlineTunnel: failed to create pair record manager: %w", err)
	}
	return createInlineTunnelManualPairing(ctx, device, pm)
}

// createInlineTunnelLockdown creates an inline tunnel for iOS 17.4+ using lockdown
// It reuses ConnectUserSpaceTunnelLockdown and wraps the result in an InlineTunnel
func createInlineTunnelLockdown(ctx context.Context, device ios.DeviceEntry) (*InlineTunnel, error) {
	log.Info("Creating inline tunnel via lockdown for iOS 17.4+")

	// Use port 0 to let the OS assign a random available port
	tunnel, err := ConnectUserSpaceTunnelLockdown(device, 0)
	if err != nil {
		return nil, fmt.Errorf("createInlineTunnelLockdown: %w", err)
	}

	return &InlineTunnel{
		Tunnel:    tunnel,
		LocalPort: tunnel.UserspaceTUNPort,
	}, nil
}

// createInlineTunnelManualPairing creates an inline tunnel for iOS 17.0-17.3 using manual pairing
func createInlineTunnelManualPairing(ctx context.Context, device ios.DeviceEntry, pm PairRecordManager) (*InlineTunnel, error) {
	log.Info("Creating inline tunnel via mDNS + manual pairing for iOS 17.0-17.3")

	// Reuse the common pairing logic
	addr, tunnelListener, ts, err := manualPairAndCreateListener(ctx, device, pm)
	if err != nil {
		return nil, fmt.Errorf("createInlineTunnelManualPairing: %w", err)
	}
	defer ts.Close()

	// Connect to QUIC tunnel using userspace networking
	return connectInlineTunnel(ctx, tunnelListener, addr, device)
}

// connectInlineTunnel establishes the QUIC tunnel and sets up userspace networking
func connectInlineTunnel(ctx context.Context, info tunnelListener, addr string, device ios.DeviceEntry) (*InlineTunnel, error) {
	quicConn, tunnelInfo, err := connectQUICTunnel(ctx, info, addr)
	if err != nil {
		return nil, fmt.Errorf("connectInlineTunnel: %w", err)
	}

	// Setup userspace network interface using QUIC datagrams
	const prefixLength = 64
	iface := &UserSpaceTUNInterface{}

	// Create a wrapper that converts QUIC datagrams to/from the network stack
	quicWrapper := &quicDatagramWrapper{conn: quicConn, ctx: ctx}
	err = iface.Init(uint32(tunnelInfo.ClientParameters.Mtu), quicWrapper, tunnelInfo.ClientParameters.Address, prefixLength)
	if err != nil {
		quicConn.CloseWithError(0, "")
		return nil, fmt.Errorf("connectInlineTunnel: failed to setup tunnel interface: %w", err)
	}

	// Start local TCP listener for service connections
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		iface.networkStack.Close()
		quicConn.CloseWithError(0, "")
		return nil, fmt.Errorf("connectInlineTunnel: failed to create listener: %w", err)
	}

	localPort := listener.Addr().(*net.TCPAddr).Port
	slog.Info("Inline tunnel listening", "port", localPort)

	// Start accepting connections
	go listenToConns(*iface, listener)

	// Create closer that cleans up all resources
	closeFunc := func() error {
		var errs []error
		errs = append(errs, listener.Close())
		errs = append(errs, quicConn.CloseWithError(0, ""))
		iface.networkStack.Close()
		return errors.Join(errs...)
	}

	return &InlineTunnel{
		Tunnel: Tunnel{
			Address:          tunnelInfo.ServerAddress,
			RsdPort:          int(tunnelInfo.ServerRSDPort),
			Udid:             device.Properties.SerialNumber,
			UserspaceTUN:     true,
			UserspaceTUNPort: localPort,
			closer:           closeFunc,
		},
		LocalPort: localPort,
	}, nil
}

// quicDatagramWrapper wraps a QUIC connection to implement io.ReadWriteCloser
// It converts between QUIC datagrams and the byte stream expected by the network stack
type quicDatagramWrapper struct {
	conn   quic.Connection
	ctx    context.Context
	buffer []byte
	offset int
}

func (q *quicDatagramWrapper) Read(p []byte) (int, error) {
	// If we have buffered data, return it first
	if q.offset < len(q.buffer) {
		n := copy(p, q.buffer[q.offset:])
		q.offset += n
		if q.offset >= len(q.buffer) {
			q.buffer = nil
			q.offset = 0
		}
		return n, nil
	}

	// Read next datagram
	data, err := q.conn.ReceiveDatagram(q.ctx)
	if err != nil {
		return 0, err
	}

	n := copy(p, data)
	if n < len(data) {
		// Buffer remaining data
		q.buffer = data[n:]
		q.offset = 0
	}
	return n, nil
}

func (q *quicDatagramWrapper) Write(p []byte) (int, error) {
	// QUIC datagrams have a size limit, we need to handle large packets
	// The MTU should be configured properly so this shouldn't be an issue
	err := q.conn.SendDatagram(p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (q *quicDatagramWrapper) Close() error {
	return q.conn.CloseWithError(0, "")
}

// GetRsdPortProvider returns an RsdPortProvider for the inline tunnel
func (t *InlineTunnel) GetRsdPortProvider() (ios.RsdPortProvider, error) {
	// Connect to RSD on the tunnel to get service ports
	rsdService, err := ios.NewWithAddrPortDevice(t.Address, t.RsdPort, ios.DeviceEntry{
		UserspaceTUN:     true,
		UserspaceTUNHost: "127.0.0.1",
		UserspaceTUNPort: t.LocalPort,
	})
	if err != nil {
		return nil, fmt.Errorf("GetRsdPortProvider: failed to connect to RSD: %w", err)
	}
	defer rsdService.Close()

	rsdProvider, err := rsdService.Handshake()
	if err != nil {
		return nil, fmt.Errorf("GetRsdPortProvider: failed RSD handshake: %w", err)
	}
	return rsdProvider, nil
}

// ApplyToDevice updates the device entry with the inline tunnel information
func (t *InlineTunnel) ApplyToDevice(device *ios.DeviceEntry) error {
	rsdProvider, err := t.GetRsdPortProvider()
	if err != nil {
		return err
	}
	device.Address = t.Address
	device.Rsd = rsdProvider
	device.UserspaceTUN = true
	device.UserspaceTUNHost = "127.0.0.1"
	device.UserspaceTUNPort = t.LocalPort
	return nil
}
