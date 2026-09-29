package http

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/danielpaulus/go-ios/ios/golog"
	"golang.org/x/net/http2"
)

const logModule = "go-ios/http"

type StreamId uint32

const (
	InitStream   = StreamId(0)
	ClientServer = StreamId(1)
	ServerClient = StreamId(3)
)

// defaultWindowSize is the initial HTTP/2 flow-control window (RFC 9113 6.9.2)
// that applies until the peer announces a different one.
const defaultWindowSize = 65535

// recvWindowSize is the flow-control window we grant the peer, for the
// connection as well as for every stream. It is announced as
// SETTINGS_INITIAL_WINDOW_SIZE and replenished with WINDOW_UPDATE frames.
const recvWindowSize = 1048576

// windowUpdateThreshold is how much of a receive window may be used up before
// it is replenished, so that not every DATA frame needs a WINDOW_UPDATE.
const windowUpdateThreshold = recvWindowSize / 2

// maxFrameSize is the largest DATA frame payload we send. It's the HTTP/2
// default for SETTINGS_MAX_FRAME_SIZE, which every peer has to accept.
const maxFrameSize = 16384

// HttpConnection is a wrapper around a http2.Framer that provides a simple interface to read and write http2 streams for iOS17+.
type HttpConnection struct {
	closer io.Closer

	framer *http2.Framer

	// framerReadMu guards reading frames. A http2.Framer is not safe for concurrent
	// reads, and the payload of a frame is only valid until the next read, so a
	// frame has to be dispatched to its stream before the next one is read.
	framerReadMu sync.Mutex
	// framerWriteMu guards writing frames, a http2.Framer encodes every frame into
	// the same buffer.
	framerWriteMu sync.Mutex

	// mu guards the flow-control state and the additional streams below.
	mu sync.Mutex
	// peerInitialWindow is the peer's SETTINGS_INITIAL_WINDOW_SIZE, the send
	// window every newly opened stream starts with.
	peerInitialWindow int64
	// connSendWindow is the connection level send window. Only writes on
	// additional streams wait for it, see Stream.Write.
	connSendWindow int64
	// connRecvUnacked counts the bytes received on the connection that have not
	// been granted back to the peer with a WINDOW_UPDATE yet.
	connRecvUnacked int64
	// streams holds the additional client initiated streams (5, 7, ...) that
	// are used for XPC file transfers.
	streams      map[uint32]*streamState
	nextStreamId uint32
}

type streamState struct {
	buf        bytes.Buffer
	sendWindow int64
	// recvUnacked counts the bytes received on this stream that have not been
	// granted back to the peer with a WINDOW_UPDATE yet.
	recvUnacked int64
	reset       bool
}

func (r *HttpConnection) Close() error {
	return r.closer.Close()
}

func NewHttpConnection(rw io.ReadWriteCloser) (*HttpConnection, error) {
	framer := http2.NewFramer(rw, rw)

	_, err := rw.Write([]byte("PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n"))
	if err != nil {
		return nil, fmt.Errorf("NewHttpConnection: could not write PRI. %w", err)
	}

	err = framer.WriteSettings(
		http2.Setting{ID: http2.SettingMaxConcurrentStreams, Val: 100},
		http2.Setting{ID: http2.SettingInitialWindowSize, Val: recvWindowSize},
	)
	if err != nil {
		return nil, fmt.Errorf("NewHttpConnection: could not write settings. %w", err)
	}

	// SETTINGS_INITIAL_WINDOW_SIZE doesn't apply to the connection window, so it
	// is raised to the same size explicitly (RFC 9113 6.9.2)
	err = framer.WriteWindowUpdate(uint32(InitStream), recvWindowSize-defaultWindowSize)
	if err != nil {
		return nil, fmt.Errorf("NewHttpConnection: could not write window update. %w", err)
	}
	//
	frame, err := framer.ReadFrame()
	if err != nil {
		return nil, fmt.Errorf("NewHttpConnection: could not read frame. %w", err)
	}
	peerInitialWindow := int64(defaultWindowSize)
	if frame.Header().Type == http2.FrameSettings {
		settings := frame.(*http2.SettingsFrame)
		v, ok := settings.Value(http2.SettingInitialWindowSize)
		if ok {
			framer.SetMaxReadFrameSize(v)
			peerInitialWindow = int64(v)
		}
		err := framer.WriteSettingsAck()
		if err != nil {
			return nil, fmt.Errorf("NewHttpConnection: could not write settings ack. %w", err)
		}
	} else {
		golog.Warn("expected setttings frame", "module", logModule, "frame", frame.Header().String())
	}

	return &HttpConnection{
		framer:            framer,
		closer:            rw,
		peerInitialWindow: peerInitialWindow,
		connSendWindow:    defaultWindowSize,
		streams:           map[uint32]*streamState{},
		nextStreamId:      uint32(1),
	}, nil
}

func (r *HttpConnection) Write(p []byte, streamId uint32) (int, error) {
	r.framerWriteMu.Lock()
	err := r.framer.WriteData(streamId, false, p)
	r.framerWriteMu.Unlock()
	if err != nil {
		return 0, fmt.Errorf("Write: could not write data. %w", err)
	}
	r.mu.Lock()
	r.connSendWindow -= int64(len(p))
	r.mu.Unlock()
	return len(p), nil
}

// processFrame reads and handles a single frame. Callers waiting for stream
// data or for flow-control windows to open call it repeatedly until their
// condition is met.
//
// Only one goroutine reads from the connection at a time, and it dispatches the
// frames of all streams. ready is therefore evaluated again once this goroutine
// owns the read side: another goroutine may have delivered what the caller is
// waiting for in the meantime, and reading one more frame would block until the
// peer happens to send another one.
func (r *HttpConnection) processFrame(ready func() bool) error {
	r.framerReadMu.Lock()
	defer r.framerReadMu.Unlock()
	if ready() {
		return nil
	}
	f, err := r.framer.ReadFrame()
	if err != nil {
		return fmt.Errorf("could not read frame. %w", err)
	}
	switch f.Header().Type {
	case http2.FrameData:
		d := f.(*http2.DataFrame)
		// the whole frame payload counts against the receive windows, the
		// padding included (RFC 9113 6.9.1)
		size := int64(d.Header().Length)
		r.mu.Lock()
		s, ok := r.streams[d.StreamID]
		streamIncrement := int64(0)
		if ok {
			s.buf.Write(d.Data())
			// a stream that the peer is done with doesn't need its window back
			if !s.reset && !d.StreamEnded() {
				streamIncrement = ackReceived(&s.recvUnacked, size)
			}
		}
		connIncrement := ackReceived(&r.connRecvUnacked, size)
		r.mu.Unlock()
		if !ok {
			return fmt.Errorf("unknown stream id %d", d.StreamID)
		}
		// the received data is buffered without a bound, so the windows are
		// replenished right away instead of when a reader consumes the data
		if err := r.writeWindowUpdate(uint32(InitStream), connIncrement); err != nil {
			return err
		}
		if err := r.writeWindowUpdate(d.StreamID, streamIncrement); err != nil {
			return err
		}
	case http2.FrameGoAway:
		return fmt.Errorf("received GOAWAY")
	case http2.FrameSettings:
		s := f.(*http2.SettingsFrame)
		if s.Flags&http2.FlagSettingsAck != http2.FlagSettingsAck {
			if v, ok := s.Value(http2.SettingInitialWindowSize); ok {
				r.updateInitialWindow(int64(v))
			}
			r.framerWriteMu.Lock()
			err := r.framer.WriteSettingsAck()
			r.framerWriteMu.Unlock()
			if err != nil {
				return fmt.Errorf("could not write settings ack. %w", err)
			}
		}
	case http2.FrameWindowUpdate:
		w := f.(*http2.WindowUpdateFrame)
		r.mu.Lock()
		if w.StreamID == uint32(InitStream) {
			r.connSendWindow += int64(w.Increment)
		} else if s, ok := r.streams[w.StreamID]; ok {
			s.sendWindow += int64(w.Increment)
		}
		r.mu.Unlock()
	case http2.FrameRSTStream:
		rst := f.(*http2.RSTStreamFrame)
		r.mu.Lock()
		s, ok := r.streams[rst.StreamID]
		if ok {
			s.reset = true
		}
		r.mu.Unlock()
		// The device resets file transfer streams once it received all data.
		// That's no reason to fail reads on the XPC streams.
		if !ok {
			return fmt.Errorf("got RST frame with error code: %s", rst.ErrCode.String())
		}
	default:
		break
	}
	return nil
}

// updateInitialWindow applies a new SETTINGS_INITIAL_WINDOW_SIZE of the peer,
// which also adjusts the send windows of all open streams (RFC 9113 6.9.2).
func (r *HttpConnection) updateInitialWindow(v int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delta := v - r.peerInitialWindow
	r.peerInitialWindow = v
	for _, s := range r.streams {
		s.sendWindow += delta
	}
}

// ackReceived adds the n received bytes to unacked and returns the increment to
// grant back with a WINDOW_UPDATE frame, or 0 while the window still has room.
func ackReceived(unacked *int64, n int64) int64 {
	*unacked += n
	if *unacked < windowUpdateThreshold {
		return 0
	}
	increment := *unacked
	*unacked = 0
	return increment
}

// writeWindowUpdate grants increment bytes of receive window back to the peer.
// The WINDOW_UPDATE is only sent when increment > 0, otherwise this is a no-op.
func (r *HttpConnection) writeWindowUpdate(streamId uint32, increment int64) error {
	if increment <= 0 {
		return nil
	}
	r.framerWriteMu.Lock()
	err := r.framer.WriteWindowUpdate(streamId, uint32(increment))
	r.framerWriteMu.Unlock()
	if err != nil {
		return fmt.Errorf("could not write window update for stream %d. %w", streamId, err)
	}
	return nil
}

// Stream is an additional client initiated HTTP/2 stream. RemoteXPC uses those
// for transferring the payload of file transfer objects.
//
// Writes honor the peer's flow-control windows and read frames from the
// connection while they wait for window updates. Every stream of a connection
// can be read and written from its own goroutine.
type Stream struct {
	h  *HttpConnection
	id uint32
}

// OpenStream opens a new client initiated stream.
func (r *HttpConnection) OpenStream() (*Stream, error) {
	r.mu.Lock()
	id := r.nextStreamId
	// Client controlled streams are always odd numbered (RFC 9113 5.1.1)
	r.nextStreamId += 2
	r.streams[id] = &streamState{sendWindow: r.peerInitialWindow}
	r.mu.Unlock()

	r.framerWriteMu.Lock()
	err := r.framer.WriteHeaders(http2.HeadersFrameParam{
		StreamID:   id,
		EndHeaders: true,
	})
	r.framerWriteMu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("OpenStream: could not send headers for stream %d. %w", id, err)
	}
	return &Stream{h: r, id: id}, nil
}

// Read blocks until len(p) bytes were received on this stream
func (s *Stream) Read(p []byte) (int, error) {
	for {
		s.h.mu.Lock()
		st := s.h.streams[s.id]
		if st.buf.Len() >= len(p) {
			n, err := st.buf.Read(p)
			s.h.mu.Unlock()
			return n, err
		}
		reset := st.reset
		s.h.mu.Unlock()
		if reset {
			return 0, fmt.Errorf("Read: stream %d was reset by the peer", s.id)
		}
		if err := s.h.processFrame(func() bool { return s.readable(len(p)) }); err != nil {
			return 0, fmt.Errorf("Read: %w", err)
		}
	}
}

// readable reports whether a read of n bytes can complete, either because the
// data arrived or because the peer reset the stream.
func (s *Stream) readable(n int) bool {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()
	st := s.h.streams[s.id]
	return st.reset || st.buf.Len() >= n
}

// Write sends p as DATA frames on this stream, waiting for the peer to open its
// flow-control windows whenever they are exhausted.
func (s *Stream) Write(p []byte) (int, error) {
	written := 0
	for written < len(p) {
		n, err := s.sendWindow(len(p) - written)
		if err != nil {
			return written, fmt.Errorf("Write: %w", err)
		}
		if n == 0 {
			if err := s.h.processFrame(s.writable); err != nil {
				return written, fmt.Errorf("Write: failed waiting for window update. %w", err)
			}
			continue
		}
		s.h.framerWriteMu.Lock()
		err = s.h.framer.WriteData(s.id, false, p[written:written+n])
		s.h.framerWriteMu.Unlock()
		if err != nil {
			return written, fmt.Errorf("Write: could not write data on stream %d. %w", s.id, err)
		}
		written += n
	}
	return written, nil
}

// writable reports whether a write can make progress, either because both send
// windows have room or because the peer reset the stream.
func (s *Stream) writable() bool {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()
	st := s.h.streams[s.id]
	return st.reset || (st.sendWindow > 0 && s.h.connSendWindow > 0)
}

// sendWindow reserves up to want bytes of the stream and connection windows and
// returns how many bytes may be sent right now.
func (s *Stream) sendWindow(want int) (int, error) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()
	st := s.h.streams[s.id]
	if st.reset {
		return 0, fmt.Errorf("stream %d was reset by the peer", s.id)
	}
	n := int64(min(want, maxFrameSize))
	n = min(n, st.sendWindow, s.h.connSendWindow)
	if n <= 0 {
		return 0, nil
	}
	st.sendWindow -= n
	s.h.connSendWindow -= n
	return int(n), nil
}

// Close half-closes the stream by sending an empty DATA frame with END_STREAM
func (s *Stream) Close() error {
	s.h.framerWriteMu.Lock()
	err := s.h.framer.WriteData(s.id, true, nil)
	s.h.framerWriteMu.Unlock()
	if err != nil {
		return fmt.Errorf("Close: could not end stream %d. %w", s.id, err)
	}
	return nil
}
