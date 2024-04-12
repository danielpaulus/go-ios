package syslog

import (
	"bufio"
	"io"

	"github.com/danielpaulus/go-ios/ios"
)

const (
	usbmuxdServiceName string = "com.apple.syslog_relay"
	shimServiceName           = "com.apple.syslog_relay.shim.remote"
)

// Connection exposes the LogReader channel which send the LogMessages as strings.
type Connection struct {
	closer         io.Closer
	bufferedReader *bufio.Reader
}

// New returns a new SysLog Connection for the given DeviceID and Udid
// It will create LogReader as a buffered Channel because Syslog is very verbose.
func New(device ios.DeviceEntry) (*Connection, error) {
	if !device.SupportsRsd() {
		return NewWithUsbmuxdConnection(device)
	}
	return NewWithShimConnection(device)
}

// NewWithUsbmuxdConnection connects to the syslog_relay service on the device over the usbmuxd socket
func NewWithUsbmuxdConnection(device ios.DeviceEntry) (*Connection, error) {
	deviceConn, err := ios.ConnectToService(device, usbmuxdServiceName)
	if err != nil {
		return &Connection{}, err
	}
	return &Connection{
		closer:         deviceConn,
		bufferedReader: bufio.NewReader(deviceConn),
	}, nil
}

// NewWithShimConnection connects to the syslog_relay service over a tunnel interface and the service port
// is obtained from remote service discovery
func NewWithShimConnection(device ios.DeviceEntry) (*Connection, error) {
	deviceConn, err := ios.ConnectToShimService(device, shimServiceName)
	if err != nil {
		return nil, err
	}
	return &Connection{
		closer:         deviceConn,
		bufferedReader: bufio.NewReader(deviceConn),
	}, nil
}

// ReadLogMessage this is a blocking function that will return individual log messages received from syslog.
// Call it in an endless for loop in a separate go routine.
func (sysLogConn *Connection) ReadLogMessage() (string, error) {
	logmsg, err := sysLogConn.bufferedReader.ReadString(0)
	if err != nil {
		return "", err
	}
	return logmsg, nil
}

// Close closes the underlying UsbMuxConnection
func (sysLogConn *Connection) Close() error {
	return sysLogConn.closer.Close()
}
