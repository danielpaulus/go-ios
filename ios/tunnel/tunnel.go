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
	"github.com/danielpaulus/go-ios/ios"
	"io"
	"math/big"
	"os/exec"
	"strings"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/sirupsen/logrus"
	"github.com/songgao/water"
)

type Tunnel struct {
	Address string
	RsdPort int

	quicConn   quic.Connection
	utunCloser io.Closer
	ctxCancel  context.CancelFunc
}

func (t Tunnel) Close() error {
	t.ctxCancel()
	quicErr := t.quicConn.CloseWithError(0, "")
	utunErr := t.utunCloser.Close()
	return errors.Join(quicErr, utunErr)
}

// ManualPairAndConnectToTunnel tries to verify an existing pairing, and if this fails it triggers a new manual pairing process.
// After a successful pairing a tunnel for this device gets started and the tunnel information is returned
func ManualPairAndConnectToTunnel(ctx context.Context, device ios.DeviceEntry, p PairRecordManager) (Tunnel, error) {
	addr, err := ios.FindDeviceInterfaceAddress(ctx, device)
	if err != nil {
		return Tunnel{}, fmt.Errorf("ManualPairAndConnectToTunnel: failed to find device ethernet interface: %w", err)
	}

	port, err := getUntrustedTunnelServicePort(addr)
	if err != nil {
		return Tunnel{}, fmt.Errorf("ManualPairAndConnectToTunnel: could not find port for '%s'", UntrustedTunnelServiceName)
	}
	h, err := ios.ConnectToHttp2WithAddr(addr, port)
	if err != nil {
		return Tunnel{}, fmt.Errorf("ManualPairAndConnectToTunnel: failed to create HTTP2 connection: %w", err)
	}

	xpcConn, err := ios.CreateXpcConnection(h)
	if err != nil {
		return Tunnel{}, fmt.Errorf("ManualPairAndConnectToTunnel: failed to create RemoteXPC connection: %w", err)
	}
	ts := NewTunnelServiceWithXpc(xpcConn, h, p)

	err = ts.ManualPair()
	if err != nil {
		return Tunnel{}, fmt.Errorf("ManualPairAndConnectToTunnel: failed to pair device: %w", err)
	}
	tunnelInfo, err := ts.createTunnelListener()
	if err != nil {
		return Tunnel{}, fmt.Errorf("ManualPairAndConnectToTunnel: failed to create tunnel listener: %w", err)
	}
	t, err := ConnectToTunnel(ctx, tunnelInfo, addr)
	if err != nil {
		return Tunnel{}, fmt.Errorf("ManualPairAndConnectToTunnel: failed to connect to tunnel: %w", err)
	}
	return t, nil
}

func getUntrustedTunnelServicePort(addr string) (int, error) {
	rsdService, err := ios.NewWithAddr(addr)
	if err != nil {
		return 0, fmt.Errorf("getUntrustedTunnelServicePort: failed to connect to RSD service: %w", err)
	}
	defer rsdService.Close()
	handshakeResponse, err := rsdService.Handshake()
	if err != nil {
		return 0, fmt.Errorf("getUntrustedTunnelServicePort: failed to perform RSD handshake: %w", err)
	}

	port := handshakeResponse.GetPort(UntrustedTunnelServiceName)
	if port == 0 {
		return 0, fmt.Errorf("getUntrustedTunnelServicePort: could not find port for '%s'", UntrustedTunnelServiceName)
	}
	return port, nil
}

func ConnectToTunnel(ctx context.Context, info TunnelListener, addr string) (Tunnel, error) {
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

	err = conn.SendDatagram(make([]byte, 1))
	if err != nil {
		return Tunnel{}, err
	}

	tunnelInfo, err := exchangeCoreTunnelParameters(conn)
	if err != nil {
		return Tunnel{}, fmt.Errorf("could not exchange tunnel parameters. %w", err)
	}

	utunIface, err := setupTunnelInterface(err, tunnelInfo)
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

	return Tunnel{
		Address:    tunnelInfo.ServerAddress,
		RsdPort:    int(tunnelInfo.ServerRSDPort),
		quicConn:   conn,
		utunCloser: utunIface,
		ctxCancel:  cancel,
	}, nil
}

func setupTunnelInterface(err error, tunnelInfo TunnelInfo) (*water.Interface, error) {
	ifce, err := water.New(water.Config{
		DeviceType: water.TUN,
	})
	if err != nil {
		logrus.Fatal(err)
	}

	const prefixLength = 64 // TODO: this could be calculated from the netmask provided by the device
	cmd := exec.Command("ifconfig", ifce.Name(), "inet6", "add", fmt.Sprintf("%s/%d", tunnelInfo.ClientParameters.Address, prefixLength))
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("could not get stderr. %w", err)
	}
	stdErrOutput := &strings.Builder{}
	go func() {
		_, _ = io.Copy(stdErrOutput, stderr)
	}()
	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	err = cmd.Wait()
	if err != nil {
		return nil, fmt.Errorf("cmd failed: '%s'", stdErrOutput.String())
	}
	return ifce, nil
}

func createTlsConfig(info TunnelListener) (*tls.Config, error) {
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

func exchangeCoreTunnelParameters(conn quic.Connection) (TunnelInfo, error) {
	stream, err := conn.OpenStream()
	if err != nil {
		return TunnelInfo{}, err
	}
	defer stream.Close()

	rq, err := json.Marshal(map[string]interface{}{
		"type": "clientHandshakeRequest",
		"mtu":  1280,
	})
	if err != nil {
		return TunnelInfo{}, err
	}

	buf := bytes.NewBuffer(nil)
	buf.Write([]byte("CDTunnel\000"))
	buf.WriteByte(byte(len(rq)))
	buf.Write(rq)

	_, err = stream.Write(buf.Bytes())
	if err != nil {
		return TunnelInfo{}, err
	}

	header := make([]byte, len("CDTunnel")+2)
	n, err := stream.Read(header)
	if err != nil {
		return TunnelInfo{}, fmt.Errorf("could not header read from stream. %w", err)
	}

	bodyLen := header[len(header)-1]

	res := make([]byte, bodyLen)
	n, err = stream.Read(res)
	if err != nil {
		return TunnelInfo{}, fmt.Errorf("could not read from stream. %w", err)
	}

	var tunnelInfo TunnelInfo
	err = json.Unmarshal(res[:n], &tunnelInfo)
	if err != nil {
		return TunnelInfo{}, err
	}
	return tunnelInfo, nil
}
