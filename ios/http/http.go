package http

import (
	"bytes"
	"fmt"
	"io"
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

// fileStreamWindowUpdateThreshold is how many received bytes are batched before the
// flow-control windows (connection and file stream) are replenished. It must stay well
// below the advertised initial window size (1 MiB) so large transfers never stall.
const fileStreamWindowUpdateThreshold = 256 * 1024

// fileStream buffers the raw content the device sends on an additional HTTP2 stream
// during an XPC file transfer. ended is set once the device half-closes the stream.
type fileStream struct {
	buf    bytes.Buffer
	ended  bool
	isOpen atomic.Bool
}

// HttpConnection is a wrapper around a http2.Framer that provides a simple interface to read and write http2 streams for iOS17+.
type HttpConnection struct {
	framer             *http2.Framer
	clientServerStream *bytes.Buffer
	serverClientStream *bytes.Buffer
	closer             io.Closer
	csIsOpen           *atomic.Bool
	scIsOpen           *atomic.Bool
	fileStreams        map[uint32]*fileStream
	pendingWindow      map[uint32]uint32
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
	if frame.Header().Type == http2.FrameSettings {
		settings := frame.(*http2.SettingsFrame)
		v, ok := settings.Value(http2.SettingInitialWindowSize)
		if ok {
			framer.SetMaxReadFrameSize(v)
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
		fileStreams:        map[uint32]*fileStream{},
		pendingWindow:      map[uint32]uint32{},
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
	return len(p), nil
}

func (r *HttpConnection) readDataFrame() error {
	for {
		f, err := r.framer.ReadFrame()
		if err != nil {
			return fmt.Errorf("readDataFrame: could not read frame. %w", err)
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
				s, ok := r.fileStreams[d.StreamID]
				if !ok {
					return fmt.Errorf("readDataFrame: unknown stream id %d", d.StreamID)
				}
				s.buf.Write(d.Data())
				if d.StreamEnded() {
					s.ended = true
				}
				// Flow control accounts for the whole DATA frame payload, including the
				// Pad Length byte and padding (RFC 7540 §6.9.1), so replenish by the frame
				// header Length, not just the unpadded data d.Data() exposes. Crediting only
				// the unpadded length would leak window on padded frames and eventually stall.
				if err := r.replenishReceiveWindow(d.StreamID, d.Length, s.ended); err != nil {
					return fmt.Errorf("readDataFrame: %w", err)
				}
			}
			return nil
		case http2.FrameGoAway:
			return fmt.Errorf("received GOAWAY")
		case http2.FrameSettings:
			s := f.(*http2.SettingsFrame)
			if s.Flags&http2.FlagSettingsAck != http2.FlagSettingsAck {
				err := r.framer.WriteSettingsAck()
				if err != nil {
					return fmt.Errorf("readDataFrame: could not write settings ack. %w", err)
				}
			}
		case http2.FrameRSTStream:
			r := f.(*http2.RSTStreamFrame)
			return fmt.Errorf("readDataFrame: got RST frame with error code: %s", r.ErrCode.String())
		default:
			break
		}
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

// replenishReceiveWindow grants received bytes back to the device's flow-control windows
// (connection and file stream) so transfers larger than the initial window size keep
// flowing. Increments are batched until fileStreamWindowUpdateThreshold to limit frame
// overhead; flush forces out any remainder (used when a stream ends).
func (r *HttpConnection) replenishReceiveWindow(streamId uint32, received uint32, flush bool) error {
	pending := r.pendingWindow[streamId] + received
	if pending < fileStreamWindowUpdateThreshold && !flush {
		r.pendingWindow[streamId] = pending
		return nil
	}
	delete(r.pendingWindow, streamId)
	if pending == 0 {
		return nil
	}
	if err := r.framer.WriteWindowUpdate(uint32(InitStream), pending); err != nil {
		return fmt.Errorf("replenishReceiveWindow: could not update connection window: %w", err)
	}
	if err := r.framer.WriteWindowUpdate(streamId, pending); err != nil {
		return fmt.Errorf("replenishReceiveWindow: could not update window of stream %d: %w", streamId, err)
	}
	return nil
}

// registerFileStream reserves an additional stream id for an XPC file transfer so that
// incoming DATA frames on it are buffered instead of rejected.
func (r *HttpConnection) registerFileStream(streamId uint32) (*fileStream, error) {
	switch StreamId(streamId) {
	case InitStream, ClientServer, ServerClient:
		return nil, fmt.Errorf("registerFileStream: stream id %d is reserved", streamId)
	}
	if _, ok := r.fileStreams[streamId]; ok {
		return nil, fmt.Errorf("registerFileStream: stream id %d is already registered", streamId)
	}
	s := &fileStream{}
	r.fileStreams[streamId] = s
	return s, nil
}

// readFileStream reads buffered file content of the given stream, pumping frames off the
// connection as needed. It returns io.EOF once the device half-closed the stream and all
// buffered content was consumed.
func (r *HttpConnection) readFileStream(streamId uint32, p []byte) (int, error) {
	s, ok := r.fileStreams[streamId]
	if !ok {
		return 0, fmt.Errorf("readFileStream: stream id %d is not registered", streamId)
	}
	for s.buf.Len() == 0 {
		if s.ended {
			return 0, io.EOF
		}
		if err := r.readDataFrame(); err != nil {
			return 0, fmt.Errorf("readFileStream: %w", err)
		}
	}
	return s.buf.Read(p)
}

// FileStreamReadWriter reads and writes an additional HTTP2 stream used by RemoteXPC file
// transfers (e.g. com.apple.dt.remoteFetchSymbols): the client opens the stream and the
// device streams the raw file content back on it. Read returns io.EOF once the device
// half-closes the stream. Like HttpConnection it is not safe for concurrent use.
type FileStreamReadWriter struct {
	h        *HttpConnection
	streamId uint32
	stream   *fileStream
}

// NewFileStreamReadWriter registers the given stream id (must not be one of the reserved
// XPC streams 0, 1 and 3) on the connection and returns a ReadWriter for it.
func NewFileStreamReadWriter(h *HttpConnection, streamId uint32) (FileStreamReadWriter, error) {
	s, err := h.registerFileStream(streamId)
	if err != nil {
		return FileStreamReadWriter{}, fmt.Errorf("NewFileStreamReadWriter: %w", err)
	}
	return FileStreamReadWriter{h: h, streamId: streamId, stream: s}, nil
}

func (f FileStreamReadWriter) Read(p []byte) (int, error) {
	return f.h.readFileStream(f.streamId, p)
}

func (f FileStreamReadWriter) Write(p []byte) (int, error) {
	return f.h.write(p, f.streamId, &f.stream.isOpen)
}
