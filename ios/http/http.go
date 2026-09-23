package http

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

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

// maxFrameSize is the largest DATA frame payload we send. It's the HTTP/2
// default for SETTINGS_MAX_FRAME_SIZE, which every peer has to accept.
const maxFrameSize = 16384

// HttpConnection is a wrapper around a http2.Framer that provides a simple interface to read and write http2 streams for iOS17+.
type HttpConnection struct {
	framer             *http2.Framer
	clientServerStream *bytes.Buffer
	serverClientStream *bytes.Buffer
	closer             io.Closer
	csIsOpen           *atomic.Bool
	scIsOpen           *atomic.Bool

	// mu guards the flow-control state and the additional streams below.
	mu sync.Mutex
	// peerInitialWindow is the peer's SETTINGS_INITIAL_WINDOW_SIZE, the send
	// window every newly opened stream starts with.
	peerInitialWindow int64
	// connSendWindow is the connection level send window. Only writes on
	// additional streams wait for it, see Stream.Write.
	connSendWindow int64
	// streams holds the additional client initiated streams (5, 7, ...) that
	// are used for XPC file transfers.
	streams      map[uint32]*streamState
	nextStreamId uint32
}

type streamState struct {
	buf        bytes.Buffer
	sendWindow int64
	reset      bool
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
		http2.Setting{ID: http2.SettingInitialWindowSize, Val: 1048576},
	)
	if err != nil {
		return nil, fmt.Errorf("NewHttpConnection: could not write settings. %w", err)
	}

	err = framer.WriteWindowUpdate(uint32(InitStream), 983041)
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
		framer:             framer,
		clientServerStream: bytes.NewBuffer(nil),
		serverClientStream: bytes.NewBuffer(nil),
		closer:             rw,
		csIsOpen:           &atomic.Bool{},
		scIsOpen:           &atomic.Bool{},
		peerInitialWindow:  peerInitialWindow,
		connSendWindow:     defaultWindowSize,
		streams:            map[uint32]*streamState{},
		nextStreamId:       uint32(ServerClient) + 2,
	}, nil
}

func (r *HttpConnection) ReadClientServerStream(p []byte) (int, error) {
	for r.clientServerStream.Len() < len(p) {
		err := r.readDataFrame()
		if err != nil {
			return 0, fmt.Errorf("ReadClientServerStream: %w", err)
		}
	}
	return r.clientServerStream.Read(p)
}

func (r *HttpConnection) WriteClientServerStream(p []byte) (int, error) {
	return r.write(p, uint32(ClientServer), r.csIsOpen)
}

func (r *HttpConnection) WriteServerClientStream(p []byte) (int, error) {
	return r.write(p, uint32(ServerClient), r.scIsOpen)
}

func (r *HttpConnection) write(p []byte, stream uint32, isOpen *atomic.Bool) (int, error) {
	if isOpen.CompareAndSwap(false, true) {
		err := r.framer.WriteHeaders(http2.HeadersFrameParam{
			StreamID:   stream,
			EndHeaders: true,
		})
		if err != nil {
			return 0, fmt.Errorf("write: could not send headers. %w", err)
		}
	}
	return r.Write(p, stream)
}

func (r *HttpConnection) Write(p []byte, streamId uint32) (int, error) {
	err := r.framer.WriteData(streamId, false, p)
	if err != nil {
		return 0, fmt.Errorf("Write: could not write data. %w", err)
	}
	r.mu.Lock()
	r.connSendWindow -= int64(len(p))
	r.mu.Unlock()
	return len(p), nil
}

func (r *HttpConnection) readDataFrame() error {
	for {
		isData, err := r.processFrame()
		if err != nil {
			return fmt.Errorf("readDataFrame: %w", err)
		}
		if isData {
			return nil
		}
	}
}

// processFrame reads and handles a single frame. It reports whether the frame
// was a DATA frame so that callers waiting for stream data can re-check their
// buffers.
func (r *HttpConnection) processFrame() (bool, error) {
	f, err := r.framer.ReadFrame()
	if err != nil {
		return false, fmt.Errorf("could not read frame. %w", err)
	}
	switch f.Header().Type {
	case http2.FrameData:
		d := f.(*http2.DataFrame)
		switch d.StreamID {
		case 1:
			r.clientServerStream.Write(d.Data())
		case 3:
			r.serverClientStream.Write(d.Data())
		default:
			r.mu.Lock()
			s, ok := r.streams[d.StreamID]
			if ok {
				s.buf.Write(d.Data())
			}
			r.mu.Unlock()
			if !ok {
				return false, fmt.Errorf("unknown stream id %d", d.StreamID)
			}
		}
		return true, nil
	case http2.FrameGoAway:
		return false, fmt.Errorf("received GOAWAY")
	case http2.FrameSettings:
		s := f.(*http2.SettingsFrame)
		if s.Flags&http2.FlagSettingsAck != http2.FlagSettingsAck {
			if v, ok := s.Value(http2.SettingInitialWindowSize); ok {
				r.updateInitialWindow(int64(v))
			}
			err := r.framer.WriteSettingsAck()
			if err != nil {
				return false, fmt.Errorf("could not write settings ack. %w", err)
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
			return false, fmt.Errorf("got RST frame with error code: %s", rst.ErrCode.String())
		}
	default:
		break
	}
	return false, nil
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

func (r *HttpConnection) ReadServerClientStream(p []byte) (int, error) {
	for r.serverClientStream.Len() < len(p) {
		err := r.readDataFrame()
		if err != nil {
			return 0, err
		}
	}
	return r.serverClientStream.Read(p)
}

type HttpStreamReadWriter struct {
	h        *HttpConnection
	streamId uint32
}

func NewStreamReadWriter(h *HttpConnection, streamId StreamId) HttpStreamReadWriter {
	return HttpStreamReadWriter{
		h:        h,
		streamId: uint32(streamId),
	}
}

func (h HttpStreamReadWriter) Read(p []byte) (n int, err error) {
	if h.streamId == 1 {
		return h.h.ReadClientServerStream(p)
	}
	if h.streamId == 3 {
		return h.h.ReadServerClientStream(p)
	}
	return 0, fmt.Errorf("Read: unknown stream id %d", h.streamId)
}

func (h HttpStreamReadWriter) Write(p []byte) (n int, err error) {
	if h.streamId == 1 {
		return h.h.WriteClientServerStream(p)
	}
	if h.streamId == 3 {
		return h.h.WriteServerClientStream(p)
	}
	return 0, fmt.Errorf("Write: unknown stream id %d", h.streamId)
}

// Stream is an additional client initiated HTTP/2 stream. RemoteXPC uses those
// for transferring the payload of file transfer objects.
//
// Writes honor the peer's flow-control windows and read frames from the
// connection while they wait for window updates. A Stream must therefore not be
// written while another goroutine reads from the same HttpConnection.
type Stream struct {
	h  *HttpConnection
	id uint32
}

// OpenStream opens a new client initiated stream.
func (r *HttpConnection) OpenStream() (*Stream, error) {
	r.mu.Lock()
	id := r.nextStreamId
	r.nextStreamId += 2
	r.streams[id] = &streamState{sendWindow: r.peerInitialWindow}
	r.mu.Unlock()

	err := r.framer.WriteHeaders(http2.HeadersFrameParam{
		StreamID:   id,
		EndHeaders: true,
	})
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
		if _, err := s.h.processFrame(); err != nil {
			return 0, fmt.Errorf("Read: %w", err)
		}
	}
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
			if _, err := s.h.processFrame(); err != nil {
				return written, fmt.Errorf("Write: failed waiting for window update. %w", err)
			}
			continue
		}
		if err := s.h.framer.WriteData(s.id, false, p[written:written+n]); err != nil {
			return written, fmt.Errorf("Write: could not write data on stream %d. %w", s.id, err)
		}
		written += n
	}
	return written, nil
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
	if err := s.h.framer.WriteData(s.id, true, nil); err != nil {
		return fmt.Errorf("Close: could not end stream %d. %w", s.id, err)
	}
	return nil
}
