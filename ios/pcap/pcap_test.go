package pcap

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/gopacket/pcapgo"
	"github.com/lunixbochs/struc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"howett.net/plist"
)

// fakeDeviceConnection is an in-memory stand-in for the pcapd service
// connection. Close unblocks a pending Reader() read, like closing a real
// device connection does.
type fakeDeviceConnection struct {
	r *io.PipeReader
}

func (f *fakeDeviceConnection) Reader() io.Reader { return f.r }
func (f *fakeDeviceConnection) Close() error      { return f.r.Close() }

func testPacketHeader(payloadLen int) IOSPacketHeader {
	return IOSPacketHeader{
		HdrSize:        PacketHeaderSize,
		Version:        2,
		PacketSize:     uint32(payloadLen),
		Type:           1,
		IO:             1,
		ProtocolFamily: 2,
		FramePreLength: 14,
		IFName:         "en0",
		Pid:            -1,
		ProcName:       "testproc",
		Pid2:           -1,
		ProcName2:      "testproc",
		TsSec:          1700000000,
		TsUsec:         123456,
	}
}

// packHeader serializes an IOSPacketHeader the way pcapd does on the wire.
func packHeader(t *testing.T, iph IOSPacketHeader) []byte {
	t.Helper()
	var buf bytes.Buffer
	err := struc.Pack(&buf, &iph)
	require.NoError(t, err)
	return buf.Bytes()
}

// makeFrame builds one full pcapd wire frame: a 4 byte big endian length
// followed by a binary plist containing header+payload as a data element.
func makeFrame(t *testing.T, iph IOSPacketHeader, payload []byte) []byte {
	t.Helper()
	raw := append(packHeader(t, iph), payload...)
	plistBytes, err := plist.Marshal(raw, plist.BinaryFormat)
	require.NoError(t, err)
	frame := make([]byte, 4, 4+len(plistBytes))
	binary.BigEndian.PutUint32(frame, uint32(len(plistBytes)))
	return append(frame, plistBytes...)
}

func resetFilters(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		Pid = int32(-2)
		ProcName = ""
	})
	Pid = int32(-2)
	ProcName = ""
}

func TestFromBytes(t *testing.T) {
	payload := []byte{0xde, 0xad, 0xbe, 0xef}
	plistBytes, err := plist.Marshal(payload, plist.BinaryFormat)
	require.NoError(t, err)
	decoded, err := fromBytes(plistBytes)
	require.NoError(t, err)
	assert.Equal(t, payload, decoded)

	_, err = fromBytes([]byte("not a plist"))
	assert.Error(t, err)
}

func TestGetPacket(t *testing.T) {
	resetFilters(t)
	payload := []byte{0x01, 0x02, 0x03, 0x04, 0x05}

	t.Run("parses header and payload", func(t *testing.T) {
		iph, packet, err := getPacket(append(packHeader(t, testPacketHeader(len(payload))), payload...))
		require.NoError(t, err)
		assert.Equal(t, payload, packet)
		assert.Equal(t, PacketHeaderSize, iph.HdrSize)
		assert.Equal(t, int32(-1), iph.Pid)
		assert.Equal(t, "testproc", trimNull(iph.ProcName))
		assert.Equal(t, "en0", trimNull(iph.IFName))
		assert.Equal(t, 1700000000, iph.TsSec)
		assert.Equal(t, 123456, iph.TsUsec)
	})

	t.Run("prepends fake ethernet header when FramePreLength is zero", func(t *testing.T) {
		hdr := testPacketHeader(len(payload))
		hdr.FramePreLength = 0
		_, packet, err := getPacket(append(packHeader(t, hdr), payload...))
		require.NoError(t, err)
		require.Len(t, packet, len(payload)+14)
		assert.Equal(t, payload, packet[14:])
	})

	t.Run("skips extended header bytes (iOS 15 beta4)", func(t *testing.T) {
		hdr := testPacketHeader(len(payload))
		hdr.HdrSize = PacketHeaderSize + 4
		raw := append(packHeader(t, hdr), 0xaa, 0xbb, 0xcc, 0xdd)
		_, packet, err := getPacket(append(raw, payload...))
		require.NoError(t, err)
		assert.Equal(t, payload, packet)
	})

	t.Run("pid filter drops other pids", func(t *testing.T) {
		resetFilters(t)
		Pid = 42
		_, packet, err := getPacket(append(packHeader(t, testPacketHeader(len(payload))), payload...))
		require.NoError(t, err)
		assert.Empty(t, packet)

		hdr := testPacketHeader(len(payload))
		hdr.Pid = 42
		_, packet, err = getPacket(append(packHeader(t, hdr), payload...))
		require.NoError(t, err)
		assert.Equal(t, payload, packet)
	})

	t.Run("process name filter drops other processes", func(t *testing.T) {
		resetFilters(t)
		ProcName = "other"
		_, packet, err := getPacket(append(packHeader(t, testPacketHeader(len(payload))), payload...))
		require.NoError(t, err)
		assert.Empty(t, packet)

		ProcName = "test"
		_, packet, err = getPacket(append(packHeader(t, testPacketHeader(len(payload))), payload...))
		require.NoError(t, err)
		assert.Equal(t, payload, packet)
	})
}

func TestCaptureWritesPacketsUntilCanceled(t *testing.T) {
	resetFilters(t)
	payloads := [][]byte{
		bytes.Repeat([]byte{0x11}, 20),
		bytes.Repeat([]byte{0x22}, 40),
		bytes.Repeat([]byte{0x33}, 60),
	}

	pr, pw := io.Pipe()
	conn := &fakeDeviceConnection{r: pr}
	var out bytes.Buffer
	require.NoError(t, writePcapHeader(&out))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	captureDone := make(chan error, 1)
	go func() { captureDone <- capture(ctx, conn, &out) }()

	// io.Pipe writes complete only once capture has consumed the frame.
	for _, payload := range payloads {
		_, err := pw.Write(makeFrame(t, testPacketHeader(len(payload)), payload))
		require.NoError(t, err)
	}
	cancel()

	select {
	case err := <-captureDone:
		require.NoError(t, err, "capture must return nil on context cancellation")
	case <-time.After(5 * time.Second):
		t.Fatal("capture did not stop after context cancellation")
	}

	// The finalized output must be a valid pcap stream containing every packet.
	r, err := pcapgo.NewReader(bytes.NewReader(out.Bytes()))
	require.NoError(t, err, "output is not a valid pcap stream")
	for i, payload := range payloads {
		data, ci, err := r.ReadPacketData()
		require.NoError(t, err, "packet %d missing from pcap output", i)
		assert.Equal(t, payload, data, "packet %d payload mismatch", i)
		assert.Equal(t, len(payload), ci.CaptureLength)
		assert.Equal(t, len(payload), ci.Length)
		assert.Equal(t, time.Unix(1700000000, 123456*1000).UTC(), ci.Timestamp.UTC())
	}
	_, _, err = r.ReadPacketData()
	assert.ErrorIs(t, err, io.EOF, "pcap output must end cleanly after the captured packets")
}

func TestCaptureStopsOnDeadline(t *testing.T) {
	resetFilters(t)
	pr, pw := io.Pipe()
	defer pw.Close()
	conn := &fakeDeviceConnection{r: pr}
	var out bytes.Buffer
	require.NoError(t, writePcapHeader(&out))

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	captureDone := make(chan error, 1)
	go func() { captureDone <- capture(ctx, conn, &out) }()

	payload := bytes.Repeat([]byte{0x44}, 10)
	_, err := pw.Write(makeFrame(t, testPacketHeader(len(payload)), payload))
	require.NoError(t, err)
	// No more frames arrive; the deadline must stop the blocked read.

	select {
	case err := <-captureDone:
		require.NoError(t, err, "capture must return nil on deadline")
	case <-time.After(5 * time.Second):
		t.Fatal("capture did not stop on context deadline")
	}

	r, err := pcapgo.NewReader(bytes.NewReader(out.Bytes()))
	require.NoError(t, err)
	data, _, err := r.ReadPacketData()
	require.NoError(t, err)
	assert.Equal(t, payload, data)
	_, _, err = r.ReadPacketData()
	assert.ErrorIs(t, err, io.EOF)
}

func TestCaptureReturnsReadErrorWhenNotCanceled(t *testing.T) {
	resetFilters(t)
	pr, pw := io.Pipe()
	conn := &fakeDeviceConnection{r: pr}
	var out bytes.Buffer

	captureDone := make(chan error, 1)
	go func() { captureDone <- capture(context.Background(), conn, &out) }()
	require.NoError(t, pw.CloseWithError(io.ErrUnexpectedEOF))

	select {
	case err := <-captureDone:
		require.Error(t, err, "capture must surface connection errors when the context is still active")
	case <-time.After(5 * time.Second):
		t.Fatal("capture did not stop on connection error")
	}
}

// closeTrackingConn records how often Close is called and blocks in Close
// until released, so a test can prove capture waits for the watchdog's
// conn.Close() to finish before returning (no concurrent double-close).
type closeTrackingConn struct {
	r        *io.PipeReader
	release  chan struct{}
	closes   int32
	inClose  chan struct{}
	closedCh chan struct{}
}

func (c *closeTrackingConn) Reader() io.Reader { return c.r }

func (c *closeTrackingConn) Close() error {
	if atomic.AddInt32(&c.closes, 1) == 1 {
		close(c.inClose)
		<-c.release
		err := c.r.Close()
		close(c.closedCh)
		return err
	}
	return c.r.Close()
}

func TestCaptureWaitsForWatchdogCloseBeforeReturning(t *testing.T) {
	resetFilters(t)
	pr, _ := io.Pipe()
	conn := &closeTrackingConn{
		r:        pr,
		release:  make(chan struct{}),
		inClose:  make(chan struct{}),
		closedCh: make(chan struct{}),
	}
	var out bytes.Buffer
	require.NoError(t, writePcapHeader(&out))

	ctx, cancel := context.WithCancel(context.Background())
	captureDone := make(chan error, 1)
	go func() { captureDone <- capture(ctx, conn, &out) }()

	cancel()
	// The watchdog is now inside Close, blocked before it finishes.
	select {
	case <-conn.inClose:
	case <-time.After(5 * time.Second):
		t.Fatal("watchdog did not call Close after cancellation")
	}

	// capture must not have returned yet: it owns the connection and must wait
	// for the watchdog's Close to complete so it never races the caller's own
	// deferred Close.
	select {
	case <-captureDone:
		t.Fatal("capture returned before the watchdog's Close finished")
	case <-time.After(100 * time.Millisecond):
	}

	close(conn.release)
	select {
	case err := <-captureDone:
		require.NoError(t, err, "capture must return nil on cancellation")
	case <-time.After(5 * time.Second):
		t.Fatal("capture did not return after Close finished")
	}
	<-conn.closedCh
	assert.Equal(t, int32(1), atomic.LoadInt32(&conn.closes), "connection must be closed exactly once by capture")
}

func trimNull(s string) string {
	return string(bytes.Trim([]byte(s), "\x00"))
}
