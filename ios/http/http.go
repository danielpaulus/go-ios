package http

import (
	"bytes"
	"fmt"
	"io"
	"sync/atomic"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
)

type StreamId uint32

const (
	InitStream   = StreamId(0)
	ClientServer = StreamId(1)
	ServerClient = StreamId(3)
)

// HttpConnection is a wrapper around a http2.Framer that provides a simple interface to read and write http2 streams for iOS17+.
type HttpConnection struct {
	framer             *http2.Framer
	clientServerStream *bytes.Buffer
	serverClientStream *bytes.Buffer
	closer             io.Closer
	csIsOpen           *atomic.Bool
	scIsOpen           *atomic.Bool
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
		log.WithField("frame", frame.Header().String()).
			Warn("expected setttings frame")
	}

	return &HttpConnection{
		framer:             framer,
		clientServerStream: bytes.NewBuffer(nil),
		serverClientStream: bytes.NewBuffer(nil),
		closer:             rw,
		csIsOpen:           &atomic.Bool{},
		scIsOpen:           &atomic.Bool{},
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
				return fmt.Errorf("readDataFrame: unknown stream id %d", d.StreamID)
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
