// Package xpc contains a connection stuct and the codec for the xpc protocol.
// The xpc protocol is used to communicate with services on iOS17+ devices.
package xpc

import (
	"fmt"
	"io"
)

// Connection represents a http2 based connection to an XPC service on an iOS17 device.
type Connection struct {
	connectionCloser io.Closer
	msgId            uint64
	clientServer     io.ReadWriter
	serverClient     io.ReadWriter
	openStream       StreamOpener
}

// StreamOpener opens an additional stream on the underlying connection. RemoteXPC
// sends the payload of each FileTransfer object on a stream of its own.
type StreamOpener func() (io.ReadWriteCloser, error)

// New creates a new connection to an XPC service on an iOS17 device.
func New(clientServer io.ReadWriter, serverClient io.ReadWriter, openStream StreamOpener, closer io.Closer) (*Connection, error) {
	return &Connection{
		connectionCloser: closer,
		msgId:            1,
		clientServer:     clientServer,
		serverClient:     serverClient,
		openStream:       openStream,
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

// FileTransferStream carries the payload of a FileTransfer object
type FileTransferStream struct {
	rwc io.ReadWriteCloser
	id  uint64
}

// OpenFileTransfer opens a new stream for the payload of the FileTransfer object with
// the given id. The peer accepts it only after it received the message that
// references the FileTransfer, so call WaitAccepted after sending that message.
func (c *Connection) OpenFileTransfer(id uint64) (*FileTransferStream, error) {
	if c.openStream == nil {
		return nil, fmt.Errorf("OpenFileTransfer: connection does not support file transfers")
	}
	rwc, err := c.openStream()
	if err != nil {
		return nil, fmt.Errorf("OpenFileTransfer: %w", err)
	}
	err = EncodeMessage(rwc, Message{
		Flags: AlwaysSetFlag | FileOpenFlag,
		Id:    id,
	})
	if err != nil {
		return nil, fmt.Errorf("OpenFileTransfer: failed to send file open message: %w", err)
	}
	return &FileTransferStream{rwc: rwc, id: id}, nil
}

// WaitAccepted blocks until the peer is ready to receive the payload
func (f *FileTransferStream) WaitAccepted() error {
	msg, err := DecodeMessage(f.rwc)
	if err != nil {
		return fmt.Errorf("WaitAccepted: %w", err)
	}
	if msg.Flags&FileOpenReplyFlag == 0 || msg.Id != f.id {
		return fmt.Errorf("WaitAccepted: unexpected reply with flags 0x%x and id %d for file transfer %d", msg.Flags, msg.Id, f.id)
	}
	return nil
}

// Write sends payload data
func (f *FileTransferStream) Write(p []byte) (int, error) {
	return f.rwc.Write(p)
}

// Close marks the end of the payload
func (f *FileTransferStream) Close() error {
	return f.rwc.Close()
}

func (c *Connection) Close() error {
	return c.connectionCloser.Close()
}
