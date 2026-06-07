package tunnel

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

// ipv6HeaderLen is the size of the fixed IPv6 header in bytes. The 16-bit
// payload-length field lives at offset 4..6 and counts every byte after this
// fixed header (including any extension headers).
const ipv6HeaderLen = 40

// framedIPv6Reader turns a raw byte stream that carries back-to-back bare IPv6
// packets — such as the CoreDeviceProxy lockdown tunnel — into a packet
// boundary-preserving io.Reader: each Read returns exactly one complete IPv6
// packet, never a partial one and never two coalesced together.
//
// This matters because the lockdown tunnel is a TCP byte stream with no 1:1
// relationship between a socket Read and an IP packet. Consumers that assume
// "one Read == one packet" (e.g. gVisor's link-endpoint dispatch loop in
// rwcendpoint.go) corrupt or drop traffic under TCP coalescing without this
// reframing. The kernel TUN path performs the equivalent framing inline in
// forwardTCPToInterface (tunnel_lockdown.go).
type framedIPv6Reader struct {
	br *bufio.Reader
}

func newFramedIPv6Reader(r io.Reader) *framedIPv6Reader {
	return &framedIPv6Reader{br: bufio.NewReader(r)}
}

// Read fills p with exactly one IPv6 packet and returns its total length. The
// supplied buffer must be large enough to hold a full packet; if it is not,
// Read fails loudly rather than truncating, because a partial packet would
// desync the stream for every packet that follows.
func (r *framedIPv6Reader) Read(p []byte) (int, error) {
	if len(p) < ipv6HeaderLen {
		return 0, fmt.Errorf("framedIPv6Reader: buffer of %d bytes is smaller than the IPv6 header", len(p))
	}
	if _, err := io.ReadFull(r.br, p[:ipv6HeaderLen]); err != nil {
		return 0, fmt.Errorf("framedIPv6Reader: failed to read IPv6 header: %w", err)
	}
	if p[0]>>4 != 6 {
		return 0, fmt.Errorf("framedIPv6Reader: not an IPv6 packet: expected version 6, got %d", p[0]>>4)
	}
	payloadLength := int(binary.BigEndian.Uint16(p[4:6]))
	total := ipv6HeaderLen + payloadLength
	if total > len(p) {
		// Failing loudly is intentional. The peer negotiated the MTU in the
		// handshake, so a frame larger than the buffer means the stream is
		// already desynced; the gVisor dispatch loop treats this as fatal and
		// tears the inbound path down rather than silently corrupting every
		// subsequent packet (the bug this reader exists to prevent).
		return 0, fmt.Errorf("framedIPv6Reader: packet of %d bytes exceeds buffer of %d bytes", total, len(p))
	}
	if _, err := io.ReadFull(r.br, p[ipv6HeaderLen:total]); err != nil {
		return 0, fmt.Errorf("framedIPv6Reader: failed to read payload of length %d: %w", payloadLength, err)
	}
	return total, nil
}

// framedIPv6Conn wraps a lockdown tunnel connection so that reads preserve IPv6
// packet boundaries (via framedIPv6Reader) while writes and Close pass straight
// through to the underlying connection. Writes are already packet-aligned: each
// outbound IP packet is written in a single Write, and the device reframes on
// its side using the IPv6 header length.
type framedIPv6Conn struct {
	io.WriteCloser
	r *framedIPv6Reader
}

// newFramedIPv6Conn returns conn wrapped so that Read returns one IPv6 packet at
// a time. It must be constructed only after any handshake bytes have been
// consumed from conn, so the buffered reader starts on a packet boundary.
func newFramedIPv6Conn(conn io.ReadWriteCloser) *framedIPv6Conn {
	return &framedIPv6Conn{
		WriteCloser: conn,
		r:           newFramedIPv6Reader(conn),
	}
}

func (c *framedIPv6Conn) Read(p []byte) (int, error) { return c.r.Read(p) }
