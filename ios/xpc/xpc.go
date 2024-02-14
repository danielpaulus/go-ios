// Package xpc contains a connection stuct and the codec for the xpc protocol.
// The xpc protocol is used to communicate with services on iOS17+ devices.
package xpc

import (
	"fmt"
	"io"

	"golang.org/x/net/http2"
)

// Connection represents a http2 based connection to an XPC service on an iOS17 device.
type Connection struct {
	connectionCloser io.Closer
	framer           *http2.Framer
	msgId            uint64
	clientServer     io.ReadWriter
	serverClient     io.ReadWriter
}

// New creates a new connection to an XPC service on an iOS17 device.
func New(clientServer io.ReadWriter, serverClient io.ReadWriter, closer io.Closer) (*Connection, error) {
	return &Connection{
		connectionCloser: closer,
		msgId:            1,
		clientServer:     clientServer,
		serverClient:     serverClient,
	}, nil
}

func (c *Connection) ReceiveOnServerClientStream() (map[string]interface{}, error) {
	msg, err := DecodeMessage(c.serverClient)
	if err != nil {
		return nil, fmt.Errorf("ReceiveOnServerClientStream: %w", err)
	}
	return msg.Body, nil
}

func (c *Connection) ReceiveOnClientServerStream() (map[string]interface{}, error) {
	return c.receiveOnStream(c.clientServer)
}

func (c *Connection) receiveOnStream(r io.Reader) (map[string]interface{}, error) {
	msg, err := DecodeMessage(r)
	if err != nil {
		return nil, fmt.Errorf("receiveOnStream: %w", err)
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
