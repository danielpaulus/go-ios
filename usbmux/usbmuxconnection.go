package usbmux

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"

	log "github.com/sirupsen/logrus"
	"howett.net/plist"
)

//DefaultUsbmuxdSocket this is the unix domain socket address to connect to. The default is "/var/run/usbmuxd"
var DefaultUsbmuxdSocket = "/var/run/usbmuxd"

//MuxConnection provides a Send Method for sending Messages to UsbMuxD and a ResponseChannel to
//receive the responses.
type MuxConnection struct {
	//tag will be incremented for every message, so responses can be correlated to requests
	tag        uint32
	deviceConn DeviceConnectionInterface
}

//NewUsbMuxConnection creates a new MuxConnection by connecting to the usbmuxd Socket.
func NewUsbMuxConnection() *MuxConnection {
	return NewUsbMuxConnectionToSocket(DefaultUsbmuxdSocket)
}

//NewUsbMuxConnectionToSocket creates a new MuxConnection by connecting to the specified usbmuxd Socket.
func NewUsbMuxConnectionToSocket(socket string) *MuxConnection {
	muxConnection := &MuxConnection{tag: 0, deviceConn: NewDeviceConnection(socket)}
	muxConnection.deviceConn.Connect()
	return muxConnection
}

//NewUsbMuxConnectionWithConn creates a new MuxConnection in listening mode for proxy use.
func NewUsbMuxConnectionWithConn(c net.Conn) *MuxConnection {
	return &MuxConnection{tag: 0, deviceConn: NewDeviceConnectionWithConn(c)}
}

// NewUsbMuxConnectionWithDeviceConnection creates a new MuxConnection with from an already initialized DeviceConnectionInterface
// (only needed for testing)
func NewUsbMuxConnectionWithDeviceConnection(deviceConn DeviceConnectionInterface) *MuxConnection {
	muxConn := &MuxConnection{tag: 0, deviceConn: deviceConn}
	muxConn.deviceConn.Connect()
	return muxConn
}

//Close dereferences this MuxConn from the underlying DeviceConnections and it returns the DeviceConnection for later use.
func (muxConn *MuxConnection) Close() DeviceConnectionInterface {
	conn := muxConn.deviceConn
	muxConn.deviceConn = nil
	return conn
}

type MuxMessage struct {
	Header  UsbmuxHeader
	Payload []byte
}

//UsbmuxHeader contains the header for plist messages for the usbmux daemon.
type UsbmuxHeader struct {
	Length  uint32
	Version uint32
	Request uint32
	Tag     uint32
}

func newUsbmuxHeader(length uint32, tag uint32) *UsbmuxHeader {
	header := UsbmuxHeader{}
	header.Length = length
	header.Version = 1
	header.Request = 8
	header.Tag = tag
	return &header
}

func getHeader(length int, tag uint32) []byte {
	buf := new(bytes.Buffer)
	header := newUsbmuxHeader(16+uint32(length), tag)
	tag++
	errs := binary.Write(buf, binary.LittleEndian, header)
	if errs != nil {
		log.Fatalf("binary.Write failed: %v", errs)
	}
	return buf.Bytes()
}

// Send sends and encodes a Plist using the usbmux Encoder
func (muxConn *MuxConnection) Send(msg interface{}) error {
	bytes, err := muxConn.Encode(msg)
	if err != nil {
		log.Error("Error sending mux")
		return err
	}
	return muxConn.deviceConn.Send(bytes)
}

//SendMuxMessage serializes and sends a MuxMessage to the underlying DeviceConnection.
func (muxConn *MuxConnection) SendMuxMessage(msg MuxMessage) error {
	if muxConn.deviceConn == nil {
		return io.EOF
	}
	err := binary.Write(muxConn.deviceConn.Writer(), binary.LittleEndian, msg.Header)
	if err != nil {
		return err
	}
	return muxConn.deviceConn.Send(msg.Payload)
}

//ReadMessage blocks until the next muxMessage is available on the underlying DeviceConnection and returns it.
func (muxConn *MuxConnection) ReadMessage() (*MuxMessage, error) {
	reader := muxConn.deviceConn.Reader()
	msg, err := muxConn.Decode(reader)
	if err != nil {
		return &MuxMessage{}, err
	}
	return msg, nil
}

//Encode serializes a MuxMessage struct to a Plist and returns the []byte of its
//string representation
func (muxConn *MuxConnection) Encode(message interface{}) ([]byte, error) {
	log.Debug("UsbMux send", reflect.TypeOf(message), " on ", &muxConn.deviceConn)
	//stringContent := ToPlist(message)

	var err error
	mbytes, err := plist.MarshalIndent(message, plist.XMLFormat, " ")

	var buffer bytes.Buffer

	headerBytes := getHeader(len(mbytes), muxConn.tag)
	buffer.Write(headerBytes)
	_, err = buffer.Write(mbytes)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

//Decode reads all bytes for the next MuxMessage from r io.Reader and
//sends them to the ResponseChannel
func (muxConn MuxConnection) Decode(r io.Reader) (*MuxMessage, error) {
	if r == nil {
		return nil, errors.New("Reader was nil")
	}

	var muxHeader UsbmuxHeader

	err := binary.Read(r, binary.LittleEndian, &muxHeader)
	if err != nil {
		return nil, err
	}

	payloadBytes := make([]byte, muxHeader.Length-16)
	n, err := io.ReadFull(r, payloadBytes)
	if err != nil {
		return nil, fmt.Errorf("Error '%s' while reading usbmux package. Only %d bytes received instead of %d", err.Error(), n, muxHeader.Length-16)
	}
	log.Debug("UsbMux Receive on ", &muxConn.deviceConn)

	return &MuxMessage{muxHeader, payloadBytes}, nil
}
