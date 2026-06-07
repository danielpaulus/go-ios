package tunnel

import (
	"context"
	"fmt"
	"net"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/http"
	"github.com/danielpaulus/go-ios/ios/tunnel/tlspsk"
)

// ManualPairAndConnectToTunnelTCP establishes a tunnel over the TLS-PSK TCP
// transport that iOS 18.2+ uses in place of QUIC (which Apple removed). It
// mirrors ManualPairAndConnectToTunnel up to and including pairing, then asks the
// device for a TCP listener and wraps the dialed connection in a TLS 1.2 PSK
// session keyed by the pair-verify shared secret.
//
// The tunnel data plane — the CDTunnel parameter exchange plus the raw
// IPv6-length-framed packet stream — is byte-for-byte identical to the lockdown
// CoreDeviceProxy tunnel, so it is reused via connectToTunnelLockdown; only the
// transport underneath (a TLS-PSK conn instead of the lockdown service conn)
// differs.
func ManualPairAndConnectToTunnelTCP(ctx context.Context, device ios.DeviceEntry, p PairRecordManager) (Tunnel, error) {
	addr, err := ios.FindDeviceInterfaceAddress(ctx, device)
	if err != nil {
		return Tunnel{}, fmt.Errorf("ManualPairAndConnectToTunnelTCP: failed to find device interface address: %w", err)
	}

	servicePort, err := getUntrustedTunnelServicePort(addr, device)
	if err != nil {
		return Tunnel{}, fmt.Errorf("ManualPairAndConnectToTunnelTCP: could not find port for '%s'", untrustedTunnelServiceName)
	}
	conn, err := ios.ConnectTUNDevice(addr, servicePort, device)
	if err != nil {
		return Tunnel{}, fmt.Errorf("ManualPairAndConnectToTunnelTCP: failed to connect to tunnel service: %w", err)
	}
	h, err := http.NewHttpConnection(conn)
	if err != nil {
		return Tunnel{}, fmt.Errorf("ManualPairAndConnectToTunnelTCP: failed to create HTTP2 connection: %w", err)
	}
	xpcConn, err := ios.CreateXpcConnection(h)
	if err != nil {
		return Tunnel{}, fmt.Errorf("ManualPairAndConnectToTunnelTCP: failed to create RemoteXPC connection: %w", err)
	}
	ts := newTunnelServiceWithXpc(xpcConn, h, p)

	if err := ts.ManualPair(); err != nil {
		return Tunnel{}, fmt.Errorf("ManualPairAndConnectToTunnelTCP: failed to pair device: %w", err)
	}

	tunnelPort, err := ts.createTcpTunnelListener()
	if err != nil {
		return Tunnel{}, fmt.Errorf("ManualPairAndConnectToTunnelTCP: failed to create tcp tunnel listener: %w", err)
	}

	tunnelAddr := fmt.Sprintf("[%s]:%d", addr, tunnelPort)
	tcpConn, err := net.Dial("tcp", tunnelAddr)
	if err != nil {
		return Tunnel{}, fmt.Errorf("ManualPairAndConnectToTunnelTCP: failed to dial tunnel port %s: %w", tunnelAddr, err)
	}
	tlsConn, err := tlspsk.Client(tcpConn, ts.sharedSecret)
	if err != nil {
		return Tunnel{}, fmt.Errorf("ManualPairAndConnectToTunnelTCP: TLS-PSK handshake failed: %w", err)
	}

	return connectToTunnelLockdown(ctx, device, tlsConn)
}
