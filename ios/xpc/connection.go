package xpc

import (
	"io"

	"golang.org/x/net/http2"
)

type Connection struct {
	connectionCloser io.Closer
	framer           *http2.Framer
	msgId            uint64
	clientServer     io.ReadWriter
	serverClient     io.ReadWriter
}

func New(clientServer io.ReadWriter, serverClient io.ReadWriter, closer io.Closer) (*Connection, error) {
	return &Connection{
		connectionCloser: closer,
		msgId:            1,
		clientServer:     clientServer,
		serverClient:     serverClient,
	}, nil
}

func (c *Connection) Receive() (map[string]interface{}, error) {
	msg, err := DecodeMessage(c.serverClient)
	if err != nil {
		return nil, err
	}
	return msg.Body, nil
}

// Send sends the passed data as XPC message.
// Additional flags can be passed via the flags argument (the default ones are AlwaysSetFlag and if data != nil DataFlag)
func (c *Connection) Send(data map[string]interface{}, flags ...uint32) error {
	f := AlwaysSetFlag
	if data != nil {
		f |= DataFlag
	}
	for _, flag := range flags {
		f |= flag
	}
	msg := Message{
		Flags: f,
		Body:  data,
		Id:    c.msgId,
	}
	return EncodeMessage(c.clientServer, msg)
}

func (c *Connection) Close() error {
	return c.connectionCloser.Close()
}
