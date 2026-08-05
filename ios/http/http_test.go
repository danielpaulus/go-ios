package http

import (
	"bytes"
	"io"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/http2"
)

// fakeXpcServer speaks just enough http2 to answer the handshake NewHttpConnection
// performs and to push frames to the client afterwards.
type fakeXpcServer struct {
	framer *http2.Framer
	// fileStreamWindow accumulates WINDOW_UPDATE increments the client granted for the
	// file transfer stream under test.
	fileStreamWindow atomic.Int64
}

// startFakeXpcServer completes the server side of the handshake and starts a drain
// goroutine that keeps consuming client frames so client writes never block on the
// synchronous in-memory pipe.
func startFakeXpcServer(t *testing.T, fileStreamId uint32) (*HttpConnection, *fakeXpcServer) {
	t.Helper()
	clientConn, serverConn := net.Pipe()
	deadline := time.Now().Add(10 * time.Second)
	require.NoError(t, clientConn.SetDeadline(deadline))
	require.NoError(t, serverConn.SetDeadline(deadline))
	t.Cleanup(func() {
		clientConn.Close()
		serverConn.Close()
	})

	s := &fakeXpcServer{}
	handshakeDone := make(chan error, 1)
	go func() {
		preface := make([]byte, len(http2.ClientPreface))
		if _, err := io.ReadFull(serverConn, preface); err != nil {
			handshakeDone <- err
			return
		}
		f := http2.NewFramer(serverConn, serverConn)
		s.framer = f
		// client settings and window update
		for i := 0; i < 2; i++ {
			if _, err := f.ReadFrame(); err != nil {
				handshakeDone <- err
				return
			}
		}
		if err := f.WriteSettings(http2.Setting{ID: http2.SettingInitialWindowSize, Val: 1 << 20}); err != nil {
			handshakeDone <- err
			return
		}
		// client settings ack
		if _, err := f.ReadFrame(); err != nil {
			handshakeDone <- err
			return
		}
		handshakeDone <- nil
		// drain all further client frames (headers, data, window updates)
		for {
			frame, err := f.ReadFrame()
			if err != nil {
				return
			}
			if wu, ok := frame.(*http2.WindowUpdateFrame); ok && wu.StreamID == fileStreamId {
				s.fileStreamWindow.Add(int64(wu.Increment))
			}
		}
	}()

	h, err := NewHttpConnection(clientConn)
	require.NoError(t, err)
	require.NoError(t, <-handshakeDone)
	return h, s
}

// TestFileStreamChunkReassembly downloads a payload that arrives in many DATA frames on a
// file transfer side stream, interleaved with data for the regular server-client stream,
// and verifies chunk reassembly, EOF on END_STREAM, buffering of the interleaved stream
// and flow-control window replenishment.
func TestFileStreamChunkReassembly(t *testing.T) {
	const streamId = uint32(2)
	h, s := startFakeXpcServer(t, streamId)

	stream, err := NewFileStreamReadWriter(h, streamId)
	require.NoError(t, err)
	// opening the stream from the client side must not fail (headers + payload frame)
	_, err = stream.Write([]byte("xpc open message"))
	require.NoError(t, err)

	payload := bytes.Repeat([]byte("dyld-shared-cache!"), 20*1024) // 360 KiB, above the window update threshold
	interleaved := []byte("reply-channel-data")

	serverErr := make(chan error, 1)
	go func() {
		const chunkSize = 16000
		for i := 0; i < len(payload); i += chunkSize {
			end := i + chunkSize
			if end > len(payload) {
				end = len(payload)
			}
			if err := s.framer.WriteData(streamId, false, payload[i:end]); err != nil {
				serverErr <- err
				return
			}
			if i == 0 {
				// push data for the regular server-client stream in the middle of the
				// file transfer, it must get buffered instead of corrupting the file
				if err := s.framer.WriteData(uint32(ServerClient), false, interleaved); err != nil {
					serverErr <- err
					return
				}
			}
		}
		serverErr <- s.framer.WriteData(streamId, true, nil)
	}()

	received, err := io.ReadAll(stream)
	require.NoError(t, err)
	require.NoError(t, <-serverErr)
	assert.Equal(t, payload, received)

	// reading again keeps returning EOF
	_, err = stream.Read(make([]byte, 1))
	assert.ErrorIs(t, err, io.EOF)

	// the interleaved frame must be readable on the regular server-client stream
	scBuf := make([]byte, len(interleaved))
	_, err = io.ReadFull(NewStreamReadWriter(h, ServerClient), scBuf)
	require.NoError(t, err)
	assert.Equal(t, interleaved, scBuf)

	// the client must have granted the full payload size back to the stream window
	require.Eventually(t, func() bool {
		return s.fileStreamWindow.Load() == int64(len(payload))
	}, 5*time.Second, 10*time.Millisecond, "expected window updates for %d bytes, got %d", len(payload), s.fileStreamWindow.Load())
}

func TestFileStreamUnknownStreamStillErrors(t *testing.T) {
	const streamId = uint32(2)
	h, s := startFakeXpcServer(t, streamId)

	stream, err := NewFileStreamReadWriter(h, streamId)
	require.NoError(t, err)

	serverErr := make(chan error, 1)
	go func() {
		// data for a stream that was never registered must be rejected
		serverErr <- s.framer.WriteData(6, false, []byte("bogus"))
	}()

	_, err = stream.Read(make([]byte, 1))
	assert.ErrorContains(t, err, "unknown stream id 6")
	require.NoError(t, <-serverErr)
}

func TestNewFileStreamReadWriterRejectsInvalidIds(t *testing.T) {
	h := &HttpConnection{fileStreams: map[uint32]*fileStream{}, pendingWindow: map[uint32]uint32{}}

	for _, reserved := range []uint32{uint32(InitStream), uint32(ClientServer), uint32(ServerClient)} {
		_, err := NewFileStreamReadWriter(h, reserved)
		assert.ErrorContains(t, err, "reserved", "stream id %d", reserved)
	}

	_, err := NewFileStreamReadWriter(h, 2)
	require.NoError(t, err)
	_, err = NewFileStreamReadWriter(h, 2)
	assert.ErrorContains(t, err, "already registered")
}
