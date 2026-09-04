package display

import (
	"errors"
	"fmt"
	"net"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/golog"
)

// The device sends RTP to a host address, which only a kernel tunnel gives a
// real interface; the userspace stack registers TCP only.
var ErrUserspaceTunnelUnsupported = errors.New("media streams require a kernel tunnel: start the tunnel without --userspace")

// Sized to absorb a burst while the reader is between reads. Overflowing it
// makes the device see packet loss and throttle its encoder.
const receiverReadBuffer = 1024 * 1024

// Receiver is the UDP socket the device streams RTP to. Its address and port go
// to StartVideoStream; the caller then uses Read or Drain.
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
		// Not fatal, so the stream still starts, but at warn: the smaller buffer
		// drops frames under a burst, and the device throttles when it sees that.
		golog.Warn("could not enlarge the RTP receive buffer, keeping the default",
			"module", logModule, "error", err)
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

// Read receives one RTP packet and returns its length. Callers that want the
// video read here; callers that only need the stream to exist use Drain.
func (r *Receiver) Read(packet []byte) (int, error) {
	n, _, err := r.conn.ReadFromUDP(packet)
	if err != nil {
		if errors.Is(err, net.ErrClosed) {
			return 0, net.ErrClosed
		}
		return 0, fmt.Errorf("the RTP stream stopped being received: %w", err)
	}
	return n, nil
}

// Drain discards packets until the Receiver is closed, which is the only way to
// stop it. Run it in a goroutine: unread, the socket fills and the device throttles.
func (r *Receiver) Drain() error {
	buf := make([]byte, 65535)
	for {
		if _, err := r.Read(buf); err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			return err
		}
	}
}

// Close closes the socket, which also unblocks Drain.
func (r *Receiver) Close() error {
	return r.conn.Close()
}

// hostTunnelAddress returns the host's address on the tunnel. Dialing sends
// nothing; it just makes the OS bind the local address the device can reach.
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
