package syslog

import (
	"bufio"
	"errors"
	"io"

	"github.com/danielpaulus/go-ios/usbmux"
)

const serviceName string = "com.apple.syslog_relay"

//Connection exposes the LogReader channel which send the LogMessages as strings.
type Connection struct {
	muxConn        *usbmux.MuxConnection
	logReader      chan string
	bufferedReader *bufio.Reader
}

//New returns a new SysLog Connection for the given DeviceID and Udid
//It will create LogReader as a buffered Channel because Syslog is very verbose.
func New(deviceID int, udid string, pairRecord usbmux.PairRecord) (*Connection, error) {
	startServiceResponse := usbmux.StartService(deviceID, udid, serviceName)
	var sysLogConn Connection
	sysLogConn.muxConn = usbmux.NewUsbMuxConnection()
	err := sysLogConn.muxConn.ConnectWithStartServiceResponse(deviceID, *startServiceResponse, &sysLogConn, pairRecord)
	if err != nil {
		return &Connection{}, err
	}
	sysLogConn.logReader = make(chan string, 200)

	return &sysLogConn, nil
}

//ReadLogMessage this is a blocking function that will return individual log messages received from syslog.
//Call it in an endless for loop in a separate go routine.
func (sysLogConn *Connection) ReadLogMessage() string {
	return <-sysLogConn.logReader
}

//Encode returns only and error because syslog is read only.
func (sysLogConn *Connection) Encode(message interface{}) ([]byte, error) {
	return nil, errors.New("Syslog is readonly")
}

//Decode wraps the reader into a buffered reader and reads nullterminated strings from it
//syslog is very verbose, so the decoder sends the decoded strings to a bufferedChannel
//in a non blocking style.
//Do not call this manually, it is used by the underlying DeviceConnection.
func (sysLogConn *Connection) Decode(r io.Reader) error {
	if sysLogConn.bufferedReader == nil {
		sysLogConn.bufferedReader = bufio.NewReader(r)
	}

	stringmessage, err := sysLogConn.bufferedReader.ReadString(0)
	if err != nil {
		return err
	}
	select {
	case sysLogConn.logReader <- stringmessage:
	default:

	}
	return nil
}

//Close closes the underlying UsbMuxConnection
func (sysLogConn *Connection) Close() {
	sysLogConn.muxConn.Close()
}
