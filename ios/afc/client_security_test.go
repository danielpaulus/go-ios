package afc

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"
)

// buildHeader serializes an AFC header into its 40-byte little-endian wire form.
func buildHeader(magic, entireLen, thisLen, packetNum uint64, op opcode) []byte {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, header{
		Magic:     magic,
		EntireLen: entireLen,
		ThisLen:   thisLen,
		PacketNum: packetNum,
		Operation: op,
	})
	return buf.Bytes()
}

// readPacketFrom feeds the given bytes into a Client and runs readPacket. It
// must never panic; a malformed header must surface as an error instead.
func readPacketFrom(t *testing.T, raw []byte) (packet, error) {
	t.Helper()
	c := &Client{connection: nopReadWriteCloser{bytes.NewReader(raw)}}
	return c.readPacket()
}

type nopReadWriteCloser struct{ io.Reader }

func (nopReadWriteCloser) Write(p []byte) (int, error) { return len(p), nil }
func (nopReadWriteCloser) Close() error                { return nil }

// TestReadPacketThisLenUnderflow covers HIGH-14: ThisLen < headerSize would make
// h.ThisLen-headerSize wrap to a huge uint64 and panic in make([]byte, ...).
func TestReadPacketThisLenUnderflow(t *testing.T) {
	// ThisLen (10) is smaller than the 40-byte header size.
	raw := buildHeader(magic, 100, 10, 1, readDir)
	_, err := readPacketFrom(t, raw)
	if err == nil {
		t.Fatal("expected an error for ThisLen smaller than header size, got nil")
	}
}

// TestReadPacketEntireLenUnderflow covers HIGH-14: EntireLen < ThisLen would make
// h.EntireLen-h.ThisLen wrap to a huge uint64 and panic in make([]byte, ...).
func TestReadPacketEntireLenUnderflow(t *testing.T) {
	// EntireLen (40) is smaller than ThisLen (50).
	raw := buildHeader(magic, 40, 50, 1, readDir)
	_, err := readPacketFrom(t, raw)
	if err == nil {
		t.Fatal("expected an error for EntireLen smaller than ThisLen, got nil")
	}
}

// TestReadPacketShortStatusHeaderPayload covers HIGH-15: a status packet whose
// header payload is shorter than 8 bytes would index out of range when decoding
// the status code.
func TestReadPacketShortStatusHeaderPayload(t *testing.T) {
	// ThisLen = headerSize + 4 => a 4-byte header payload, shorter than 8.
	raw := buildHeader(magic, headerSize+4, headerSize+4, 1, status)
	raw = append(raw, []byte{1, 2, 3, 4}...)
	_, err := readPacketFrom(t, raw)
	if err == nil {
		t.Fatal("expected an error for a status header payload shorter than 8 bytes, got nil")
	}
}
