// Package openstdio is used to open a new socket connection that can be used to connect an app launched with appservice
package openstdio

import (
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/google/uuid"
)

// Connection is a single connection to the stdio socket
type Connection struct {
	// ID is required whenever we need to tell iOS which socket should be used
	// (for example when we launch an app we need to pass this id so that the system knows which socket should be used)
	ID         uuid.UUID
	connection io.ReadWriteCloser
}

// NewOpenStdIoSocket creates a new stdio-socket on the device and the returned connection can be used
// to read and write from this socket
func NewOpenStdIoSocket(device ios.DeviceEntry) (Connection, error) {
	if device.Rsd == nil {
		return Connection{}, errors.New("NewOpenStdIoSocket: no rsd device found")
	}
	port := device.Rsd.GetPort("com.apple.coredevice.openstdiosocket")
	conn, err := ios.ConnectTUNDevice(device.Address, port, device)
	if err != nil {
		return Connection{}, fmt.Errorf("NewOpenStdIoSocket: failed to open connection: %w", err)
	}

	uuidBytes := make([]byte, 16)
	_, err = conn.Read(uuidBytes)
	if err != nil {
		return Connection{}, fmt.Errorf("NewOpenStdIoSocket: failed to read UUID: %w", err)
	}

	u, err := uuid.FromBytes(uuidBytes)
	if err != nil {
		return Connection{}, fmt.Errorf("NewOpenStdIoSocket: failed to parse UUID: %w", err)
	}

	return Connection{
		ID:         u,
		connection: conn,
	}, nil
}

// Read reads from the connected stdio-socket
func (s Connection) Read(p []byte) (n int, err error) {
	if s.connection == nil {
		return 0, fmt.Errorf("Read: not connected to service")
	}
	n, err = s.connection.Read(p)
	if err != nil && errors.Is(err, net.ErrClosed) {
		return n, io.EOF
	}
	return n, err
}

// Write writes to the connected stdio-socket
func (s Connection) Write(p []byte) (n int, err error) {
	if s.connection == nil {
		return 0, fmt.Errorf("Write: not connected to service")
	}
	return s.connection.Write(p)
}

// Close closes the connected stdio-socket
func (s Connection) Close() error {
	if s.connection == nil {
		return fmt.Errorf("Close: not connected to service")
	}
	return s.connection.Close()
}
