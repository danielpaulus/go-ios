package xpc

import (
	"bytes"
	"errors"
	"io"

	"golang.org/x/net/http2"
)

const (
	allChannels             = uint32(0)
	rootChannel             = uint32(1)
	replyChannel            = uint32(3)
	xpcMaxConcurrentStreams = uint32(100)
	xpcInitialWindowSize    = uint32(1048576)
	xpcUpdatedWindowSize    = uint32(983041)
)

type framerDataWriter struct {
	Framer   http2.Framer
	StreamID uint32
}

func (writer framerDataWriter) Write(p []byte) (int, error) {
	err := writer.Framer.WriteData(writer.StreamID, false, p)

	return len(p), err
}

type Connection struct {
	framer *http2.Framer
	msgId  uint64
}

func New(reader io.Reader, writer io.Writer) (*Connection, error) {
	framer := http2.NewFramer(writer, reader)

	conn := &Connection{
		framer: framer,
		msgId:  1,
	}

	err := framer.WriteSettings(
		http2.Setting{ID: http2.SettingMaxConcurrentStreams, Val: xpcMaxConcurrentStreams},
		http2.Setting{ID: http2.SettingInitialWindowSize, Val: xpcInitialWindowSize},
	)
	if err != nil {
		return nil, err
	}

	err = framer.WriteWindowUpdate(allChannels, xpcUpdatedWindowSize)
	if err != nil {
		return nil, err
	}

	err = framer.WriteHeaders(http2.HeadersFrameParam{StreamID: rootChannel, EndHeaders: true})
	if err != nil {
		return nil, err
	}

	firstReadFrame, err := framer.ReadFrame()
	if err != nil {
		return nil, err
	} else {
		if firstReadFrame.Header().Type != http2.FrameSettings {
			return nil, errors.New("Received unexpected frame from XPC connection")
		}
		// TODO : figure out if need to act on this frame
	}

	err = framer.WriteSettingsAck()
	if err != nil {
		return nil, err
	}

	err = EncodeData(framerDataWriter{
		Framer:   *framer,
		StreamID: rootChannel,
	}, conn.msgId, map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	EncodeEmpty(framerDataWriter{
		Framer:   *framer,
		StreamID: rootChannel,
	}, conn.msgId, 0x201, false) // TODO : figure out why 0x201 (0x1 for always set, 0x200 for ?)
	if err != nil {
		return nil, err
	}

	err = framer.WriteHeaders(http2.HeadersFrameParam{StreamID: replyChannel, EndHeaders: true})
	if err != nil {
		return nil, err
	}

	EncodeEmpty(framerDataWriter{
		Framer:   *framer,
		StreamID: replyChannel,
	}, conn.msgId, 0, true)
	if err != nil {
		return nil, err
	}

	conn.msgId += 1

	_, err = conn.Receive() // TODO : figure out if need to act on this frame
	if err != nil {
		return nil, err
	}
	_, err = conn.Receive() // TODO : figure out if need to act on this frame
	if err != nil {
		return nil, err
	}
	_, err = conn.Receive() // TODO : figure out if need to act on this frame
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (c *Connection) Receive() (map[string]interface{}, error) {
	for true {
		response, err := c.framer.ReadFrame()
		if err != nil {
			return nil, err
		}

		if response.Header().Type == http2.FrameData {
			dataFrame := response.(*http2.DataFrame)
			if dataFrame.Length == 24 { // empty payload
				return nil, nil
			}
			msg, err := DecodeMessage(bytes.NewReader(dataFrame.Data()))
			if err != nil {
				return nil, err
			}
			return msg.Body, nil
		}
	}

	// Impossible to reach here
	return map[string]interface{}{}, nil
}

func (c *Connection) Send(data map[string]interface{}) error {
	err := EncodeData(framerDataWriter{
		Framer:   *c.framer,
		StreamID: rootChannel,
	}, c.msgId, data)
	if err != nil {
		return err
	}

	c.msgId += 1

	return nil
}
