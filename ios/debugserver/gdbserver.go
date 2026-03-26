package debugserver

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
)

type GDBServer struct {
	rw    io.ReadWriter
	noAck bool
}

// NewGDBServer implements wire level GDBServer protocol
func NewGDBServer(rw io.ReadWriter) *GDBServer {
	return &GDBServer{rw: rw}
}

func (*GDBServer) chksum(pck string) string {
	sum := 0
	for _, b := range pck {
		sum += int(b)
	}
	return hex.EncodeToString([]byte{byte(sum % 256)})
}

func (g *GDBServer) formatPacket(pck string) string {
	if g.noAck {
		return "$" + pck + "#" + g.chksum(pck)
	}
	return "+$" + pck + "#" + g.chksum(pck)
}

// SetNoAckMode should be called after QStartNoAckMode succeeds.
func (g *GDBServer) SetNoAckMode() {
	g.noAck = true
}

// Recv reads the next GDB RSP packet, skipping any ACK bytes (+/-).
// Returns the packet data (between $ and #).
func (g *GDBServer) Recv() (string, error) {
	// State machine:
	// 1. Skip bytes until we see '$'
	// 2. Read until '#'
	// 3. Read 2 more bytes (checksum)
	// 4. Return content between $ and #

	oneByte := make([]byte, 1)

	// Skip until '$'
	for {
		_, err := io.ReadFull(g.rw, oneByte)
		if err != nil {
			return "", fmt.Errorf("recv: waiting for '$': %w", err)
		}
		if oneByte[0] == '$' {
			break
		}
		// Skip +, -, and any other non-$ bytes
	}

	// Read until '#'
	var payload bytes.Buffer
	for {
		_, err := io.ReadFull(g.rw, oneByte)
		if err != nil {
			return "", fmt.Errorf("recv: reading payload: %w", err)
		}
		if oneByte[0] == '#' {
			break
		}
		payload.WriteByte(oneByte[0])
	}

	// Read 2-byte checksum (discard — we trust the connection)
	checksumBuf := make([]byte, 2)
	_, err := io.ReadFull(g.rw, checksumBuf)
	if err != nil {
		return "", fmt.Errorf("recv: reading checksum: %w", err)
	}

	return payload.String(), nil
}

func (g *GDBServer) Send(req string) error {
	pck := g.formatPacket(req)
	if _, err := g.rw.Write([]byte(pck)); err != nil {
		return err
	}
	return nil
}

func (g *GDBServer) Request(req string) (string, error) {
	if err := g.Send(req); err != nil {
		return "", err
	}
	return g.Recv()
}
