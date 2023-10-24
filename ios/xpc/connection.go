package xpc

import (
	"bytes"

	"golang.org/x/net/http2"
)

const (
	rootChannel  = uint32(1)
	replyChannel = uint32(3)
)

type FramerDataWriter struct {
	Framer    http2.Framer
	StreamID  uint32
	EndStream bool
}

func (writer FramerDataWriter) Write(p []byte) (int, error) {
	err := writer.Framer.WriteData(writer.StreamID, writer.EndStream, p)

	return len(p), err
}

type Connection struct {
	framer *http2.Framer
	msgId  uint64
}

func New(framer *http2.Framer) *Connection {
	return &Connection{
		framer: framer,
		msgId:  1,
	}
}

func (c *Connection) Send(data map[string]interface{}) error {
	err := EncodeData(FramerDataWriter{
		Framer:    *c.framer,
		StreamID:  rootChannel,
		EndStream: false,
	}, data, c.msgId, false)
	if err != nil {
		return err
	}

	c.msgId += 1

	firstResponse, err := c.framer.ReadFrame()
	if err != nil {
		return err
	}

	print(firstResponse)

	secondResponse, err := c.framer.ReadFrame()
	if err != nil {
		return err
	}

	print(secondResponse)

	thirdResponse, err := c.framer.ReadFrame()
	if err != nil {
		return err
	}

	print(thirdResponse)

	fourthResponse, err := c.framer.ReadFrame()
	if err != nil {
		return err
	}

	print(fourthResponse)

	dataFrame := fourthResponse.(*http2.DataFrame)
	_, err = DecodeMessage(bytes.NewReader(dataFrame.Data()))
	if err != nil {
		return err
	}

	fifthResponse, err := c.framer.ReadFrame()
	if err != nil {
		return err
	}

	print(fifthResponse)

	sixthResponse, err := c.framer.ReadFrame()
	if err != nil {
		return err
	}

	print(sixthResponse)

	seventhResponse, err := c.framer.ReadFrame()
	if err != nil {
		return err
	}

	print(seventhResponse)

	return nil
}
