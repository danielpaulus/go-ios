package tunnel

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"
	"testing/iotest"
)

// buildIPv6Packet builds a minimal well-formed IPv6 packet: a 40-byte header
// with the version nibble set to 6 and the payload-length field populated,
// followed by payload bytes.
func buildIPv6Packet(payload []byte) []byte {
	pkt := make([]byte, ipv6HeaderLen+len(payload))
	pkt[0] = 6 << 4
	binary.BigEndian.PutUint16(pkt[4:6], uint16(len(payload)))
	copy(pkt[ipv6HeaderLen:], payload)
	return pkt
}

func readOne(t *testing.T, r io.Reader, bufSize int) []byte {
	t.Helper()
	buf := make([]byte, bufSize)
	n, err := r.Read(buf)
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	return buf[:n]
}

func TestFramedIPv6Reader_SinglePacket(t *testing.T) {
	pkt := buildIPv6Packet([]byte("hello world"))
	r := newFramedIPv6Reader(bytes.NewReader(pkt))

	got := readOne(t, r, 1500)
	if !bytes.Equal(got, pkt) {
		t.Fatalf("got %d bytes, want %d (%x vs %x)", len(got), len(pkt), got, pkt)
	}
}

// TestFramedIPv6Reader_Coalesced is the core regression test: two packets that
// arrive in a single underlying read (TCP coalescing) must be returned one at a
// time, not merged into a single oversized "packet".
func TestFramedIPv6Reader_Coalesced(t *testing.T) {
	p1 := buildIPv6Packet([]byte("first packet payload"))
	p2 := buildIPv6Packet([]byte("second"))
	stream := append(append([]byte{}, p1...), p2...)

	r := newFramedIPv6Reader(bytes.NewReader(stream))

	if got := readOne(t, r, 1500); !bytes.Equal(got, p1) {
		t.Fatalf("packet 1: got %x, want %x", got, p1)
	}
	if got := readOne(t, r, 1500); !bytes.Equal(got, p2) {
		t.Fatalf("packet 2: got %x, want %x", got, p2)
	}
}

// TestFramedIPv6Reader_Split is the other half of the regression: a packet that
// dribbles in one byte at a time (TCP segmentation) must be reassembled into a
// single complete packet rather than truncated.
func TestFramedIPv6Reader_Split(t *testing.T) {
	p1 := buildIPv6Packet(bytes.Repeat([]byte{0xAB}, 200))
	p2 := buildIPv6Packet(bytes.Repeat([]byte{0xCD}, 37))
	stream := append(append([]byte{}, p1...), p2...)

	// OneByteReader forces every underlying read to return a single byte.
	r := newFramedIPv6Reader(iotest.OneByteReader(bytes.NewReader(stream)))

	if got := readOne(t, r, 1500); !bytes.Equal(got, p1) {
		t.Fatalf("packet 1: got %d bytes, want %d", len(got), len(p1))
	}
	if got := readOne(t, r, 1500); !bytes.Equal(got, p2) {
		t.Fatalf("packet 2: got %d bytes, want %d", len(got), len(p2))
	}
}

func TestFramedIPv6Reader_NotIPv6(t *testing.T) {
	pkt := buildIPv6Packet([]byte("x"))
	pkt[0] = 4 << 4 // pretend it's IPv4
	r := newFramedIPv6Reader(bytes.NewReader(pkt))

	if _, err := r.Read(make([]byte, 1500)); err == nil {
		t.Fatal("expected an error for a non-IPv6 packet, got nil")
	}
}

func TestFramedIPv6Reader_BufferTooSmallForHeader(t *testing.T) {
	r := newFramedIPv6Reader(bytes.NewReader(buildIPv6Packet([]byte("data"))))
	if _, err := r.Read(make([]byte, ipv6HeaderLen-1)); err == nil {
		t.Fatal("expected an error when the buffer is smaller than the IPv6 header, got nil")
	}
}

func TestFramedIPv6Reader_PacketExceedsBuffer(t *testing.T) {
	pkt := buildIPv6Packet(bytes.Repeat([]byte{0x01}, 100)) // total 140 bytes
	r := newFramedIPv6Reader(bytes.NewReader(pkt))
	if _, err := r.Read(make([]byte, 100)); err == nil {
		t.Fatal("expected an error when a packet exceeds the read buffer, got nil")
	}
}

func TestFramedIPv6Reader_EOF(t *testing.T) {
	r := newFramedIPv6Reader(bytes.NewReader(nil))
	if _, err := r.Read(make([]byte, 1500)); err == nil {
		t.Fatal("expected an error on EOF, got nil")
	}
}

// TestFramedIPv6Conn_WritePassthrough verifies writes go straight to the
// underlying connection unframed (the device reframes on its side).
func TestFramedIPv6Conn_WritePassthrough(t *testing.T) {
	var written bytes.Buffer
	rwc := &readWriteCloser{Reader: bytes.NewReader(buildIPv6Packet([]byte("in"))), Writer: &written}
	c := newFramedIPv6Conn(rwc)

	payload := []byte("outbound packet bytes")
	if _, err := c.Write(payload); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if !bytes.Equal(written.Bytes(), payload) {
		t.Fatalf("write passthrough: got %x, want %x", written.Bytes(), payload)
	}
}

// readWriteCloser adapts separate reader/writer into an io.ReadWriteCloser for tests.
type readWriteCloser struct {
	io.Reader
	io.Writer
}

func (readWriteCloser) Close() error { return nil }
