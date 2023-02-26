package syslog

import (
	"bufio"
	"errors"
	"io"

	"github.com/danielpaulus/go-ios/ios"
)

const serviceName string = "com.apple.syslog_relay"

// Connection exposes the LogReader channel which send the LogMessages as strings.
type Connection struct {
	deviceConn     ios.DeviceConnectionInterface
	bufferedReader *bufio.Reader
}

// New returns a new SysLog Connection for the given DeviceID and Udid
// It will create LogReader as a buffered Channel because Syslog is very verbose.
func New(device ios.DeviceEntry) (*Connection, error) {
	deviceConn, err := ios.ConnectToService(device, serviceName)
	if err != nil {
		return &Connection{}, err
	}
	return &Connection{deviceConn: deviceConn}, nil
}

// ReadLogMessage this is a blocking function that will return individual log messages received from syslog.
// Call it in an endless for loop in a separate go routine.
func (sysLogConn *Connection) ReadLogMessage() (string, error) {
	reader := sysLogConn.deviceConn.Reader()
	logmsg, err := sysLogConn.Decode(reader)
	if err != nil {
		return "", err
	}
	return logmsg, nil
}

// Encode returns only and error because syslog is read only.
func (sysLogConn *Connection) Encode(message interface{}) ([]byte, error) {
	return nil, errors.New("Syslog is readonly")
}

// Decode wraps the reader into a buffered reader and reads nullterminated strings from it
// syslog is very verbose, so the decoder sends the decoded strings to a bufferedChannel
// in a non blocking style.
// Do not call this manually, it is used by the underlying DeviceConnection.
func (sysLogConn *Connection) Decode(r io.Reader) (string, error) {
	if sysLogConn.bufferedReader == nil {
		sysLogConn.bufferedReader = bufio.NewReader(r)
	}

	stringmessage, err := sysLogConn.bufferedReader.ReadString(0)
	if err != nil {
		return "", err
	}
	return stringmessage, nil
}

// Close closes the underlying UsbMuxConnection
func (sysLogConn *Connection) Close() {
	sysLogConn.deviceConn.Close()
}
