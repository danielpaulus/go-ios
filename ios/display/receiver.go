package display

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/danielpaulus/go-ios/ios"
)

// ErrUserspaceTunnelUnsupported is returned when a media stream is requested on
// a device reached through the userspace tunnel.
//
// Media streaming is the one place where the device connects back to the host:
// it pushes RTP to an address the host nominates. The kernel tunnel assigns that
// address to a real TUN/utun interface, so a normal UDP socket is reachable. The
// userspace tunnel has no OS interface - only a localhost TCP proxy - so inbound
// UDP cannot reach a host socket. Start the tunnel without --userspace to use
// touch input.
var ErrUserspaceTunnelUnsupported = errors.New("media streams require a kernel tunnel: start the tunnel without --userspace")

// receiverReadBuffer sizes the kernel receive buffer for the RTP socket. The
// payload is discarded, but a small buffer makes the device's sender see drops
// and back off, which has shown up as encoder stalls in captured sessions.
const receiverReadBuffer = 1024 * 1024

// Receiver is the UDP sink the device streams RTP to. go-ios does not decode the
// stream: the packets exist only because dtuhidd gates touch input on a running
// media stream, so Drain reads and discards them.
type Receiver struct {
	conn *net.UDPConn
	ip   string
	port int
}

// OpenReceiver binds a UDP socket on the host's tunnel address, ready to be
// nominated as a stream's receiver. Close it when the stream is stopped.
func OpenReceiver(device ios.DeviceEntry) (*Receiver, error) {
	if device.UserspaceTUN {
		return nil, ErrUserspaceTunnelUnsupported
	}
	hostIP, err := hostTunnelAddress(device)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp6", &net.UDPAddr{IP: hostIP, Port: 0})
	if err != nil {
		return nil, fmt.Errorf("OpenReceiver: failed to bind a UDP socket on %s: %w", hostIP, err)
	}
	if err := conn.SetReadBuffer(receiverReadBuffer); err != nil {
		// Not fatal: an undersized buffer only risks dropped frames, and the
		// frames are discarded anyway.
		conn.SetReadBuffer(0)
	}
	local, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		conn.Close()
		return nil, fmt.Errorf("OpenReceiver: unexpected local address type %T", conn.LocalAddr())
	}
	return &Receiver{conn: conn, ip: hostIP.String(), port: local.Port}, nil
}

// IP is the host address the device should send RTP to.
func (r *Receiver) IP() string { return r.ip }

// Port is the UDP port the device should send RTP to.
func (r *Receiver) Port() int { return r.port }

// Drain reads and discards packets until ctx is cancelled or the socket is
// closed. Run it in its own goroutine for the lifetime of the stream: without a
// reader the socket buffer fills and the device's encoder can stall.
func (r *Receiver) Drain(ctx context.Context) {
	buf := make([]byte, 65535)
	for {
		if ctx.Err() != nil {
			return
		}
		if _, _, err := r.conn.ReadFromUDP(buf); err != nil {
			// A closed socket or a cancelled context both end the stream; any
			// other read error would repeat, so stop either way.
			return
		}
	}
}

// Close closes the socket, which also unblocks Drain.
func (r *Receiver) Close() error {
	return r.conn.Close()
}

// hostTunnelAddress finds the host's own address on the tunnel to this device.
//
// Dialing a UDP socket toward the device sends nothing but makes the OS pick the
// route and bind a local address, which is the tunnel interface address the
// device can reach us on. Reading it back avoids having to expose the tunnel's
// client parameters through the ios.DeviceEntry API.
func hostTunnelAddress(device ios.DeviceEntry) (net.IP, error) {
	if device.Address == "" {
		return nil, fmt.Errorf("hostTunnelAddress: device has no tunnel address, start a tunnel first")
	}
	// Port 9 is discard; nothing is ever sent to it.
	conn, err := net.Dial("udp6", net.JoinHostPort(device.Address, "9"))
	if err != nil {
		return nil, fmt.Errorf("hostTunnelAddress: no route to device %s, is the tunnel up? %w", device.Address, err)
	}
	defer conn.Close()

	local, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return nil, fmt.Errorf("hostTunnelAddress: unexpected local address type %T", conn.LocalAddr())
	}
	return local.IP, nil
}
