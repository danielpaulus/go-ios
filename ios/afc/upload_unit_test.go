package afc

import (
	"bytes"
	"encoding/binary"
	"io"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ackingConn records every packet the client writes and answers each request
// with a success status packet, mimicking a device that accepts all writes.
type ackingConn struct {
	out     bytes.Buffer
	pending bytes.Buffer
}

func (c *ackingConn) Write(p []byte) (int, error) { return c.out.Write(p) }

func (c *ackingConn) Read(p []byte) (int, error) {
	if c.pending.Len() == 0 {
		statusPayload := make([]byte, 8)
		binary.LittleEndian.PutUint64(statusPayload, errSuccess)
		c.pending.Write(encodeAfcPacket(status, statusPayload, nil))
	}
	return c.pending.Read(p)
}

func (c *ackingConn) Close() error { return nil }

// decodeSentPackets parses the raw bytes the client wrote into AFC packets.
func decodeSentPackets(t *testing.T, raw []byte) []packet {
	t.Helper()
	r := bytes.NewReader(raw)
	var packets []packet
	for r.Len() > 0 {
		var h header
		assert.NoError(t, binary.Read(r, binary.LittleEndian, &h))
		headerPayload := make([]byte, h.ThisLen-headerSize)
		_, err := io.ReadFull(r, headerPayload)
		assert.NoError(t, err)
		payload := make([]byte, h.EntireLen-h.ThisLen)
		_, err = io.ReadFull(r, payload)
		assert.NoError(t, err)
		packets = append(packets, packet{Header: h, HeaderPayload: headerPayload, Payload: payload})
	}
	return packets
}

// TestFileReadFromChunksLargeUploads verifies that File.ReadFrom splits an
// upload into fileWrite packets of at most maxTransferSize payload bytes and
// that the reassembled payload matches the input.
func TestFileReadFromChunksLargeUploads(t *testing.T) {
	data := make([]byte, 2*maxTransferSize+1234)
	rand.New(rand.NewSource(1)).Read(data)

	conn := &ackingConn{}
	f := &File{client: &Client{connection: conn}, handle: 7}

	n, err := f.ReadFrom(bytes.NewReader(data))
	assert.NoError(t, err)
	assert.EqualValues(t, len(data), n)

	packets := decodeSentPackets(t, conn.out.Bytes())
	assert.Len(t, packets, 3)
	var reassembled []byte
	for _, p := range packets {
		assert.Equal(t, fileWrite, p.Header.Operation)
		// the header payload carries the file handle
		assert.EqualValues(t, 7, binary.LittleEndian.Uint64(p.HeaderPayload))
		assert.LessOrEqual(t, len(p.Payload), maxTransferSize)
		reassembled = append(reassembled, p.Payload...)
	}
	assert.Len(t, packets[0].Payload, maxTransferSize)
	assert.Len(t, packets[1].Payload, maxTransferSize)
	assert.Len(t, packets[2].Payload, 1234)
	assert.Equal(t, data, reassembled)
}

// TestFileReadFromEmptyReader verifies that an empty upload does not send any
// fileWrite packets.
func TestFileReadFromEmptyReader(t *testing.T) {
	conn := &ackingConn{}
	f := &File{client: &Client{connection: conn}, handle: 1}

	n, err := f.ReadFrom(bytes.NewReader(nil))
	assert.NoError(t, err)
	assert.EqualValues(t, 0, n)
	assert.Empty(t, conn.out.Bytes())
}
