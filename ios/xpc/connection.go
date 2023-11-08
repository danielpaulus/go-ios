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
	connectionCloser io.Closer
	framer           *http2.Framer
	msgId            uint64
}

func New(readWriteCloser io.ReadWriteCloser) (*Connection, error) {
	httpMagic := "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n"
	_, err := readWriteCloser.Write([]byte(httpMagic))
	if err != nil {
		return nil, err
	}

	framer := http2.NewFramer(readWriteCloser, readWriteCloser)

	conn := &Connection{
		connectionCloser: readWriteCloser,
		framer:           framer,
		msgId:            1,
	}

	err = exchangeSettings(framer)
	if err != nil {
		return nil, err
	}

	err = exchangeData(conn, framer)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func exchangeSettings(framer *http2.Framer) error {
	err := framer.WriteSettings(
		http2.Setting{ID: http2.SettingMaxConcurrentStreams, Val: xpcMaxConcurrentStreams},
		http2.Setting{ID: http2.SettingInitialWindowSize, Val: xpcInitialWindowSize},
	)
	if err != nil {
		return err
	}

	err = framer.WriteWindowUpdate(allChannels, xpcUpdatedWindowSize)
	if err != nil {
		return err
	}

	err = framer.WriteHeaders(http2.HeadersFrameParam{StreamID: rootChannel, EndHeaders: true})
	if err != nil {
		return err
	}

	firstReadFrame, err := framer.ReadFrame()
	if err != nil {
		return err
	} else {
		if firstReadFrame.Header().Type != http2.FrameSettings {
			return errors.New("Received unexpected frame from XPC connection")
		}
		// TODO : figure out if need to act on this frame
	}

	err = framer.WriteSettingsAck()
	if err != nil {
		return err
	}

	return nil
}

func exchangeData(conn *Connection, framer *http2.Framer) error {
	err := EncodeMessage(framerDataWriter{
		Framer:   *framer,
		StreamID: rootChannel,
	}, Message{
		Flags: alwaysSetFlag | dataFlag,
		Body:  map[string]interface{}{},
		Id:    conn.msgId,
	})
	if err != nil {
		return err
	}

	_, err = conn.Receive() // TODO : figure out if need to act on this frame
	if err != nil {
		return err
	}

	err = EncodeMessage(framerDataWriter{
		Framer:   *framer,
		StreamID: rootChannel,
	}, Message{
		Flags: 0x201, // alwaysSetFlag | 0x200
		Body:  nil,
		Id:    conn.msgId,
	})
	if err != nil {
		return err
	}

	_, err = conn.Receive() // TODO : figure out if need to act on this frame
	if err != nil {
		return err
	}

	err = framer.WriteHeaders(http2.HeadersFrameParam{StreamID: replyChannel, EndHeaders: true})
	if err != nil {
		return err
	}

	err = EncodeMessage(framerDataWriter{
		Framer:   *framer,
		StreamID: replyChannel,
	}, Message{
		Flags: initHandshakeFlag | alwaysSetFlag,
		Body:  nil,
		Id:    conn.msgId,
	})
	if err != nil {
		return err
	}

	_, err = conn.Receive() // TODO : figure out if need to act on this frame
	if err != nil {
		return err
	}

	conn.msgId += 1

	return nil
}

func (c *Connection) Receive() (map[string]interface{}, error) {
	for {
		response, err := c.framer.ReadFrame()
		if err != nil {
			return nil, err
		}

		if response.Header().Type == http2.FrameData {
			dataFrame := response.(*http2.DataFrame)
			msg, err := DecodeMessage(bytes.NewReader(dataFrame.Data()))
			if err != nil {
				return nil, err
			}
			return msg.Body, nil
		}
	}
}

func (c *Connection) Send(data map[string]interface{}) error {
	err := EncodeMessage(framerDataWriter{
		Framer:   *c.framer,
		StreamID: rootChannel,
	}, Message{
		Flags: alwaysSetFlag | dataFlag,
		Body:  data,
		Id:    c.msgId,
	})
	if err != nil {
		return err
	}

	c.msgId += 1

	return nil
}

func (c *Connection) Close() error {
	return c.connectionCloser.Close()
}
