package tunnel

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os/exec"
	"runtime"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/http"

	"github.com/quic-go/quic-go"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/songgao/water"
)

// Tunnel describes the parameters of an established tunnel to the device
type Tunnel struct {
	// Address is the IPv6 address of the device over the tunnel
	Address string `json:"address"`
	// RsdPort is the port on which remote service discover is reachable
	RsdPort int `json:"rsdPort"`
	// Udid is the id of the device for this tunnel
	Udid string `json:"udid"`
	// Userspace TUN device is used, connect to the local tcp port at Default
	UserspaceTUN     bool `json:"userspaceTun"`
	UserspaceTUNPort int  `json:"userspaceTunPort"`
	closer           func() error
}

// Close closes the connection to the device and removes the virtual network interface from the host
func (t Tunnel) Close() error {
	return t.closer()
}

// ManualPairAndConnectToTunnel tries to verify an existing pairing, and if this fails it triggers a new manual pairing process.
// After a successful pairing a tunnel for this device gets started and the tunnel information is returned
func ManualPairAndConnectToTunnel(ctx context.Context, device ios.DeviceEntry, p PairRecordManager) (Tunnel, error) {
	log.Info("ManualPairAndConnectToTunnel: starting manual pairing and tunnel connection, dont forget to stop remoted first with 'sudo pkill -SIGSTOP remoted' and run this with sudo.")
	addr, err := ios.FindDeviceInterfaceAddress(ctx, device)
	if err != nil {
		return Tunnel{}, fmt.Errorf("ManualPairAndConnectToTunnel: failed to find device ethernet interface: %w", err)
	}

	port, err := getUntrustedTunnelServicePort(addr, device)
	if err != nil {
		return Tunnel{}, fmt.Errorf("ManualPairAndConnectToTunnel: could not find port for '%s'", untrustedTunnelServiceName)
	}
	conn, err := ios.ConnectTUNDevice(addr, port, device)
	if err != nil {
		return Tunnel{}, fmt.Errorf("ManualPairAndConnectToTunnel: failed to connect to TUN device: %w", err)
	}
	h, err := http.NewHttpConnection(conn)
	if err != nil {
		return Tunnel{}, fmt.Errorf("ManualPairAndConnectToTunnel: failed to create HTTP2 connection: %w", err)
	}

	xpcConn, err := ios.CreateXpcConnection(h)
	if err != nil {
		return Tunnel{}, fmt.Errorf("ManualPairAndConnectToTunnel: failed to create RemoteXPC connection: %w", err)
	}
	ts := newTunnelServiceWithXpc(xpcConn, h, p)

	err = ts.ManualPair()
	if err != nil {
		return Tunnel{}, fmt.Errorf("ManualPairAndConnectToTunnel: failed to pair device: %w", err)
	}
	tunnelInfo, err := ts.createTunnelListener()
	if err != nil {
		return Tunnel{}, fmt.Errorf("ManualPairAndConnectToTunnel: failed to create tunnel listener: %w", err)
	}
	t, err := connectToTunnel(ctx, tunnelInfo, addr, device)
	if err != nil {
		return Tunnel{}, fmt.Errorf("ManualPairAndConnectToTunnel: failed to connect to tunnel: %w", err)
	}
	return t, nil
}

func getUntrustedTunnelServicePort(addr string, device ios.DeviceEntry) (int, error) {
	rsdService, err := ios.NewWithAddrDevice(addr, device)
	if err != nil {
		return 0, fmt.Errorf("getUntrustedTunnelServicePort: failed to connect to RSD service: %w", err)
	}
	defer rsdService.Close()
	handshakeResponse, err := rsdService.Handshake()
	if err != nil {
		return 0, fmt.Errorf("getUntrustedTunnelServicePort: failed to perform RSD handshake: %w", err)
	}

	port := handshakeResponse.GetPort(untrustedTunnelServiceName)
	if port == 0 {
		return 0, fmt.Errorf("getUntrustedTunnelServicePort: could not find port for '%s'", untrustedTunnelServiceName)
	}
	return port, nil
}

func connectToTunnel(ctx context.Context, info tunnelListener, addr string, device ios.DeviceEntry) (Tunnel, error) {
	logrus.WithField("address", addr).WithField("port", info.TunnelPort).Info("connect to tunnel endpoint on device")

	conf, err := createTlsConfig(info)
	if err != nil {
		return Tunnel{}, err
	}

	conn, err := quic.DialAddr(ctx, fmt.Sprintf("[%s]:%d", addr, info.TunnelPort), conf, &quic.Config{
		EnableDatagrams: true,
		KeepAlivePeriod: 1 * time.Second,
	})
	if err != nil {
		return Tunnel{}, err
	}

	err = conn.SendDatagram(make([]byte, 1024))
	if err != nil {
		return Tunnel{}, err
	}

	stream, err := conn.OpenStream()
	if err != nil {
		return Tunnel{}, err
	}

	tunnelInfo, err := exchangeCoreTunnelParameters(stream)
	stream.Close()
	if err != nil {
		return Tunnel{}, fmt.Errorf("could not exchange tunnel parameters. %w", err)
	}

	utunIface, err := setupTunnelInterface(tunnelInfo)
	if err != nil {
		return Tunnel{}, fmt.Errorf("could not setup tunnel interface. %w", err)
	}

	// we want a copy of the parent ctx here, but it shouldn't time out/be cancelled at the same time.
	// doing it like this allows us to have a context with a timeout for the tunnel creation, but the tunnel itself
	tunnelCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))

	go func() {
		err := forwardDataToInterface(tunnelCtx, conn, utunIface)
		if err != nil {
			logrus.WithError(err).Error("failed to forward data to tunnel interface")
		}
	}()

	go func() {
		err := forwardDataToDevice(tunnelCtx, tunnelInfo.ClientParameters.Mtu, utunIface, conn)
		if err != nil {
			logrus.WithError(err).Error("failed to forward data to the device")
		}
	}()

	closeFunc := func() error {
		cancel()
		quicErr := conn.CloseWithError(0, "")
		utunErr := utunIface.Close()
		return errors.Join(quicErr, utunErr)
	}

	return Tunnel{
		Address: tunnelInfo.ServerAddress,
		RsdPort: int(tunnelInfo.ServerRSDPort),
		Udid:    device.Properties.SerialNumber,
		closer:  closeFunc,
	}, nil
}

func setupTunnelInterface(tunnelInfo tunnelParameters) (io.ReadWriteCloser, error) {
	if runtime.GOOS == "windows" {
		return setupWindowsTUN(tunnelInfo)
	}
	ifce, err := water.New(water.Config{
		DeviceType: water.TUN,
	})
	if err != nil {
		return nil, fmt.Errorf("setupTunnelInterface: failed creating TUN device %w", err)
	}

	const prefixLength = 64 // TODO: this could be calculated from the netmask provided by the device

	setIpAddr := exec.Command("ifconfig", ifce.Name(), "inet6", "add", fmt.Sprintf("%s/%d", tunnelInfo.ClientParameters.Address, prefixLength))
	err = runCmd(setIpAddr)
	if err != nil {
		return nil, fmt.Errorf("setupTunnelInterface: failed to set IP address for interface: %w", err)
	}

	// FIXME: we need to reduce the tunnel interface MTU so that the OS takes care of splitting the payloads into
	// smaller packets. If we use a larger number here, the QUIC tunnel won't send the packets properly
	// This is only necessary on MacOS, on Linux we can't set the MTU to a value less than 1280 (minimum for IPv6)
	if runtime.GOOS == "darwin" {
		ifceMtu := 1202
		setMtu := exec.Command("ifconfig", ifce.Name(), "mtu", fmt.Sprintf("%d", ifceMtu), "up")
		err = runCmd(setMtu)
		if err != nil {
			return nil, fmt.Errorf("setupTunnelInterface: failed to configure MTU: %w", err)
		}
	}

	enableIfce := exec.Command("ifconfig", ifce.Name(), "up")
	err = runCmd(enableIfce)
	if err != nil {
		return nil, fmt.Errorf("setupTunnelInterface: failed to enable interface %s: %w", ifce.Name(), err)
	}

	return ifce, nil
}

func runCmd(cmd *exec.Cmd) error {
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("runCmd: failed to exeute command (stderr: %s): %w", buf.String(), err)
	}
	return nil
}

func createTlsConfig(info tunnelListener) (*tls.Config, error) {
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		SignatureAlgorithm:    x509.SHA256WithRSA,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	cert, err := x509.CreateCertificate(rand.Reader, template, template, &info.PrivateKey.PublicKey, info.PrivateKey)
	if err != nil {
		return nil, err
	}
	privateKeyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(info.PrivateKey),
		},
	)
	certPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})
	cert5, err := tls.X509KeyPair(certPem, privateKeyPem)

	conf := &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert5},
		ClientAuth:         tls.NoClientCert,
		NextProtos:         []string{"RemotePairingTunnelProtocol"},
		CurvePreferences:   []tls.CurveID{tls.CurveP256},
	}
	return conf, nil
}

func forwardDataToDevice(ctx context.Context, mtu uint64, r io.Reader, conn quic.Connection) error {
	packet := make([]byte, mtu)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			n, err := r.Read(packet)
			if err != nil {
				return fmt.Errorf("could not read packet. %w", err)
			}
			err = conn.SendDatagram(packet[:n])
			if err != nil {
				return fmt.Errorf("could not write packet. %w", err)
			}
		}
	}
}

func forwardDataToInterface(ctx context.Context, conn quic.Connection, w io.Writer) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			b, err := conn.ReceiveDatagram(ctx)
			if err != nil {
				return fmt.Errorf("failed to read datagram. %w", err)
			}
			_, err = w.Write(b)
			if err != nil {
				return fmt.Errorf("failed to forward data. %w", err)
			}
		}
	}
}

func exchangeCoreTunnelParameters(stream io.ReadWriteCloser) (tunnelParameters, error) {
	rq, err := json.Marshal(map[string]interface{}{
		"type": "clientHandshakeRequest",
		"mtu":  1280,
	})
	if err != nil {
		return tunnelParameters{}, err
	}

	buf := bytes.NewBuffer(nil)
	// Write on bytes.Buffer never returns an error
	_, _ = buf.Write([]byte("CDTunnel\000"))
	_ = buf.WriteByte(byte(len(rq)))
	_, _ = buf.Write(rq)

	_, err = stream.Write(buf.Bytes())
	if err != nil {
		return tunnelParameters{}, err
	}

	header := make([]byte, len("CDTunnel")+2)
	n, err := stream.Read(header)
	if err != nil {
		return tunnelParameters{}, fmt.Errorf("could not header read from stream. %w", err)
	}

	bodyLen := header[len(header)-1]

	res := make([]byte, bodyLen)
	n, err = stream.Read(res)
	if err != nil {
		return tunnelParameters{}, fmt.Errorf("could not read from stream. %w", err)
	}

	var parameters tunnelParameters
	err = json.Unmarshal(res[:n], &parameters)
	if err != nil {
		return tunnelParameters{}, err
	}
	return parameters, nil
}
