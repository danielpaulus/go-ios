package tunnel

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/golog"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/link/sniffer"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/waiter"
)

// ioResourceCloser is a type for closing function.
type ioResourceCloser func()

// createIoCloser returns a ioResourceCloser for closing both writer and together
func createIoCloser(rw1, rw2 io.ReadWriteCloser) ioResourceCloser {

	// Using sync.Once is essential to close writer and reader just once
	var once sync.Once
	return func() {
		once.Do(func() {
			rw1.Close()
			rw2.Close()
		})
	}
}

// UserSpaceTUNInterface uses gVisor's netstack to create a userspace virtual network interface.
// You can use it to connect local tcp connections to remote adresses on the network.
// Set it up with the Init method and provide a io.ReadWriter to a IP/TUN compatible device.
// If EnableSniffer, raw TCP packets will be dumped to the console.
type UserSpaceTUNInterface struct {
	nicID tcpip.NICID
	//If EnableSniffer, raw TCP packets will be dumped to the console.
	EnableSniffer bool
	networkStack  *stack.Stack
	// endpoint is the link endpoint driving the device data plane; its Wait
	// returns once both packet-pump loops have stopped (i.e. the data plane died).
	endpoint *Endpoint
}

func (iface *UserSpaceTUNInterface) TunnelRWCThroughInterface(localPort uint16, remoteAddr net.IP, remotePort uint16, rw io.ReadWriteCloser) error {
	defer rw.Close()
	remote := tcpip.FullAddress{
		NIC:  iface.nicID,
		Addr: tcpip.AddrFromSlice(remoteAddr.To16()),
		Port: remotePort,
	}

	// Create TCP endpoint.
	var wq waiter.Queue
	ep, err := iface.networkStack.NewEndpoint(tcp.ProtocolNumber, ipv6.ProtocolNumber, &wq)
	if err != nil {
		return fmt.Errorf("TunnelRWCThroughInterface: NewEndpoint failed: %+v", err)
	}

	ep.SocketOptions().SetKeepAlive(true)
	// Set keep alive idle value more aggresive than the gVisor's 2 hours. NAT and Firewalls can drop the idle connections more aggresive.
	p := tcpip.KeepaliveIdleOption(30 * time.Second)
	ep.SetSockOpt(&p)

	o := tcpip.KeepaliveIntervalOption(1 * time.Second)
	ep.SetSockOpt(&o)

	// Bind if a port is specified.
	if localPort != 0 {
		if err := ep.Bind(tcpip.FullAddress{Port: localPort}); err != nil {
			return fmt.Errorf("TunnelRWCThroughInterface: Bind failed: %+v", err)
		}
	}
	// Issue connect request and wait for it to complete.
	waitEntry, notifyCh := waiter.NewChannelEntry(waiter.WritableEvents)
	wq.EventRegister(&waitEntry)
	err = ep.Connect(remote)
	if _, ok := err.(*tcpip.ErrConnectStarted); ok {
		<-notifyCh
		err = ep.LastError()
	}
	wq.EventUnregister(&waitEntry)
	if err != nil {
		return fmt.Errorf("TunnelRWCThroughInterface: Connect to remote failed: %+v", err)
	}

	golog.Info("Connected to remote", "module", logModule, "remoteAddr", remoteAddr, "remotePort", remotePort)
	remoteConn := gonet.NewTCPConn(&wq, ep)
	defer remoteConn.Close()
	perr := proxyConns(rw, remoteConn)
	if perr != nil {
		return fmt.Errorf("TunnelRWCThroughInterface: proxyConns failed: %+v", perr)
	}
	return nil
}

func proxyConns(rw1 io.ReadWriteCloser, rw2 io.ReadWriteCloser) error {

	// Use buffered channel for non-blocking send recieve. We use the same single channel 2 times for 2 ioCopyWithErr.
	errCh := make(chan error, 2)

	// Create a IO closing functions to unblock stuck io.Copy() call
	ioCloser := createIoCloser(rw1, rw2)

	// Send same error channel and the io close function
	go ioCopyWithErr(rw1, rw2, errCh, ioCloser)
	go ioCopyWithErr(rw2, rw1, errCh, ioCloser)

	// Read from error channel. As the channel is a FIFO queue first in first out, each <-errCh will read one message and remove it from the channel.
	// Order of messages are not important.
	err1 := <-errCh
	err2 := <-errCh

	return errors.Join(err1, err2)
}

func ioCopyWithErr(w io.Writer, r io.Reader, errCh chan error, ioCloser ioResourceCloser) {
	_, err := io.Copy(w, r)
	errCh <- err

	// Close the writer and reader to notify the second io.Copy() if one part of the connection closed.
	// This is also necessary to avoid resource leaking.
	ioCloser()
}

// Init initializes the virtual network interface.
// The connToTUNIface needs to be connection that understands IP packets to a remote TUN device or sth.
// provide mtu, ip address as a string and the prefix length of the interface.
func (iface *UserSpaceTUNInterface) Init(mtu uint32, connToTUNIface io.ReadWriteCloser, ipAddrString string, prefixLength int) error {
	parsedIP := net.ParseIP(ipAddrString)
	if parsedIP == nil {
		return fmt.Errorf("Init: invalid tunnel IP address %q", ipAddrString)
	}
	addr := tcpip.AddrFromSlice(parsedIP.To16())
	addrWithPrefix := addr.WithPrefix()
	addrWithPrefix.PrefixLen = prefixLength

	//Create a new stack, ipv6 is enough for ios devices
	iface.networkStack = stack.New(stack.Options{
		NetworkProtocols:   []stack.NetworkProtocolFactory{ipv6.NewProtocol},
		TransportProtocols: []stack.TransportProtocolFactory{tcp.NewProtocol},
	})

	// connToTUNIface needs to be connection that understands IP packets,
	// so we can use it to link it against a virtual network interface
	rwcEP, err := RWCEndpointNew(connToTUNIface, mtu, 0)
	if err != nil {
		return fmt.Errorf("initVirtualInterface: RWCEndpointNew failed: %+v", err)
	}
	iface.endpoint = rwcEP
	var linkEP stack.LinkEndpoint = rwcEP

	nicID := tcpip.NICID(iface.networkStack.UniqueID())
	iface.nicID = nicID
	if iface.EnableSniffer {
		linkEP = sniffer.New(linkEP)
	}
	if err := iface.networkStack.CreateNIC(nicID, linkEP); err != nil {
		return fmt.Errorf("initVirtualInterface: CreateNIC failed: %+v", err)
	}

	protocolAddr := tcpip.ProtocolAddress{
		Protocol:          ipv6.ProtocolNumber,
		AddressWithPrefix: addrWithPrefix,
	}
	if err := iface.networkStack.AddProtocolAddress(iface.nicID, protocolAddr, stack.AddressProperties{}); err != nil {
		return fmt.Errorf("initVirtualInterface: AddProtocolAddress(%d, %v, {}): %+v", nicID, protocolAddr, err)
	}

	// Add default route.
	iface.networkStack.SetRouteTable([]tcpip.Route{
		{
			Destination: header.IPv6EmptySubnet,
			NIC:         nicID,
		},
	})
	return nil
}

func ConnectUserSpaceTunnelLockdown(device ios.DeviceEntry, ifacePort int) (Tunnel, error) {
	conn, err := ios.ConnectToService(device, coreDeviceProxy)
	if err != nil {
		return Tunnel{}, err
	}
	return connectToUserspaceTunnelLockdown(context.TODO(), device, conn, ifacePort)
}

func connectToUserspaceTunnelLockdown(ctx context.Context, device ios.DeviceEntry, connToDevice io.ReadWriteCloser, ifacePort int) (Tunnel, error) {
	golog.Info("connect to lockdown tunnel endpoint on device", "module", logModule, "udid", device.Properties.SerialNumber, "ifacePort", ifacePort)
	tunnelInfo, err := exchangeCoreTunnelParameters(connToDevice)
	if err != nil {
		return Tunnel{}, fmt.Errorf("could not exchange tunnel parameters. %w", err)
	}
	golog.Info("userspace tunnel negotiated", "module", logModule, "udid", device.Properties.SerialNumber, "grantedMtu", tunnelInfo.ClientParameters.Mtu)
	const prefixLength = 64
	// The lockdown tunnel is a raw TCP byte stream carrying bare IPv6 packets, so
	// wrap it to preserve packet boundaries: the gVisor link endpoint assumes one
	// Read == one packet, and without reframing TCP coalescing corrupts and drops
	// packets (see framing.go; the kernel path reframes in forwardTCPToInterface).
	framedConn := newFramedIPv6Conn(connToDevice)
	iface := UserSpaceTUNInterface{}
	err = iface.Init(uint32(tunnelInfo.ClientParameters.Mtu), framedConn, tunnelInfo.ClientParameters.Address, prefixLength)
	if err != nil {
		return Tunnel{}, fmt.Errorf("could not setup tunnel interface. %w", err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", ifacePort))
	if err != nil {
		return Tunnel{}, fmt.Errorf("could not setup listener. %w", err)
	}

	go listenToConns(iface, listener)

	var closeOnce sync.Once
	var explicitClose atomic.Bool
	statsDone := make(chan struct{})
	doClose := func() error {
		var err error
		closeOnce.Do(func() {
			close(statsDone)
			iface.networkStack.Close()
			err = errors.Join(connToDevice.Close(), listener.Close())
		})
		return err
	}

	// Optional data-plane telemetry: set GO_IOS_TUNNEL_STATS to log, every 2s,
	// throughput plus where each pump loop spends wall-time (device read / gVisor
	// inject / waiting for gVisor / device write) and the gVisor TCP stats
	// (retransmits, timeouts, failed connections). Used to locate the concurrency
	// bottleneck without guessing.
	if os.Getenv("GO_IOS_TUNNEL_STATS") != "" {
		go logUserspaceTunnelStats(iface, device.Properties.SerialNumber, statsDone)
	}

	// If the data plane dies on its own (e.g. a fatal read error in the inbound
	// dispatch loop, which also stops the outbound loop), tear the tunnel down so
	// client connections fail fast instead of hanging forever against a half-dead
	// tunnel whose listener still accepts. Auto-recreation is the TunnelManager's
	// job once the tunnel is gone.
	go func() {
		iface.endpoint.Wait()
		if explicitClose.Load() {
			return
		}
		golog.Error("userspace tunnel data plane stopped unexpectedly, tearing down tunnel", "module", logModule, "udid", device.Properties.SerialNumber)
		_ = doClose()
	}()

	closeFunc := func() error {
		explicitClose.Store(true)
		return doClose()
	}
	return Tunnel{
		Address: tunnelInfo.ServerAddress,
		RsdPort: int(tunnelInfo.ServerRSDPort),
		Udid:    device.Properties.SerialNumber,
		closer:  closeFunc,
	}, nil
}

// logUserspaceTunnelStats periodically logs data-plane throughput, per-stage
// wall-time, and gVisor TCP counters until statsDone is closed.
func logUserspaceTunnelStats(iface UserSpaceTUNInterface, udid string, statsDone <-chan struct{}) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	var prev EndpointStats
	ms := func(nanos uint64) uint64 { return nanos / 1_000_000 }
	for {
		select {
		case <-statsDone:
			return
		case <-ticker.C:
			s := iface.endpoint.Stats()
			t := iface.networkStack.Stats().TCP
			golog.Info("userspace tunnel stats", "module", logModule, "udid", udid,
				"cumPktsIn", s.PktsIn, "cumBytesIn", s.BytesIn,
				"pktsIn", s.PktsIn-prev.PktsIn,
				"pktsOut", s.PktsOut-prev.PktsOut,
				"bytesIn", s.BytesIn-prev.BytesIn,
				"bytesOut", s.BytesOut-prev.BytesOut,
				"drops", s.Drops,
				"readMs", ms(s.ReadNanos-prev.ReadNanos),
				"injectMs", ms(s.InjectNanos-prev.InjectNanos),
				"waitMs", ms(s.WaitNanos-prev.WaitNanos),
				"writeMs", ms(s.WriteNanos-prev.WriteNanos),
				"tcpSegSent", t.SegmentsSent.Value(),
				"tcpSegRcvd", t.ValidSegmentsReceived.Value(),
				"tcpRetransmit", t.Retransmits.Value(),
				"tcpTimeouts", t.Timeouts.Value(),
				"tcpFastRetransmit", t.FastRetransmit.Value(),
				"tcpFailedConn", t.FailedConnectionAttempts.Value(),
				"tcpEstablished", t.CurrentEstablished.Value(),
			)
			prev = s
		}
	}
}

func listenToConns(iface UserSpaceTUNInterface, listener net.Listener) error {
	defer func() {
		golog.Info("Stopped listening for connections", "module", logModule)
	}()

	for {
		client, err := listener.Accept()
		if err != nil {
			// Accept fails permanently only when the listener is closed (tunnel
			// teardown). Per-connection problems are handled in the goroutine, so
			// they must never tear down the accept loop.
			return err
		}
		go handleUserspaceConn(iface, client)
	}
}

// handleUserspaceConn reads the fixed 20-byte preamble a client sends to select
// the device endpoint (16-byte remote IPv6 address + 4-byte little-endian port),
// then proxies the connection through the userspace interface. The preamble is
// read here, off the accept loop, with io.ReadFull so that (a) a slow or stalled
// client cannot block other connections from being accepted, and (b) a TCP
// segment boundary inside the preamble cannot misroute the connection.
func handleUserspaceConn(iface UserSpaceTUNInterface, client net.Conn) {
	golog.Info("Received connection request", "module", logModule, "from", client.RemoteAddr(), "to", client.LocalAddr())

	preamble := make([]byte, 20)
	if _, err := io.ReadFull(client, preamble); err != nil {
		golog.Error("failed to read userspace tunnel connection preamble", "module", logModule, "error", err)
		client.Close()
		return
	}
	remoteAddr := net.IP(preamble[:16])
	port := binary.LittleEndian.Uint32(preamble[16:20])
	golog.Info("Received connection request to device", "module", logModule, "ip", remoteAddr, "port", port)

	if err := iface.TunnelRWCThroughInterface(0, remoteAddr, uint16(port), client); err != nil {
		golog.Error("userspace tunnel connection failed", "module", logModule, "ip", remoteAddr, "port", port, "error", err)
	}
}
