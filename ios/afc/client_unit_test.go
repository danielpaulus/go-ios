package afc

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

// rwc adapts a bytes.Buffer (read side) and discards writes so we can drive a
// Client without a real device connection.
type rwc struct {
	r io.Reader
}

func (c *rwc) Read(p []byte) (int, error)  { return c.r.Read(p) }
func (c *rwc) Write(p []byte) (int, error) { return len(p), nil }
func (c *rwc) Close() error                { return nil }

// encodeAfcPacket builds a raw AFC packet for the given operation and payload.
func encodeAfcPacket(op opcode, headerPayload, payload []byte) []byte {
	thisLen := headerSize + uint64(len(headerPayload))
	h := header{
		Magic:     magic,
		EntireLen: thisLen + uint64(len(payload)),
		ThisLen:   thisLen,
		PacketNum: 0,
		Operation: op,
	}
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, h)
	buf.Write(headerPayload)
	buf.Write(payload)
	return buf.Bytes()
}

// TestFileReadOversizedPayload verifies that File.Read never returns a count
// larger than len(p), even when the device sends more bytes than were
// requested (which would violate the io.Reader contract).
func TestFileReadOversizedPayload(t *testing.T) {
	payload := []byte("0123456789") // 10 bytes
	raw := encodeAfcPacket(fileRead, nil, payload)

	f := &File{
		client: &Client{connection: &rwc{r: bytes.NewReader(raw)}},
		handle: 1,
	}

	p := make([]byte, 4) // smaller than the payload
	n, err := f.Read(p)
	assert.NoError(t, err)
	assert.LessOrEqual(t, n, len(p))
	assert.Equal(t, 4, n)
	assert.Equal(t, []byte("0123"), p)
}

// TestFileReadEOF verifies that an empty payload is reported as io.EOF.
func TestFileReadEOF(t *testing.T) {
	raw := encodeAfcPacket(fileRead, nil, nil)

	f := &File{
		client: &Client{connection: &rwc{r: bytes.NewReader(raw)}},
		handle: 1,
	}

	p := make([]byte, 8)
	n, err := f.Read(p)
	assert.Equal(t, io.EOF, err)
	assert.Equal(t, 0, n)
}

// TestAfcErrorDescriptiveMessage verifies afcError.Error() includes the
// human-readable description for a known code.
func TestAfcErrorDescriptiveMessage(t *testing.T) {
	err := afcError{code: errObjectNotFound}
	assert.Contains(t, err.Error(), "ObjectNotFound")
	assert.Contains(t, err.Error(), "8")
}
