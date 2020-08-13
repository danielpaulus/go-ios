package usbmux

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"reflect"

	log "github.com/sirupsen/logrus"
)

//DefaultUsbmuxdSocket this is the unix domain socket address to connect to. The default is "/var/run/usbmuxd"
var DefaultUsbmuxdSocket = "/var/run/usbmuxd"

//MuxConnection provides a Send Method for sending Messages to UsbMuxD and a ResponseChannel to
//receive the responses.
type MuxConnection struct {
	//tag will be incremented for every message, so responses can be correlated to requests
	tag             uint32
	deviceConn      DeviceConnectionInterface
	ResponseChannel chan []byte
	singleDecode    bool
	decodeSignal    chan interface{}
	stopSignal      chan interface{}
}

//NewUsbMuxConnection creates a new MuxConnection by connecting to the usbmuxd Socket.
func NewUsbMuxConnection() *MuxConnection {
	return NewUsbMuxConnectionToSocket(DefaultUsbmuxdSocket)
}

//NewUsbMuxConnectionToSocket creates a new MuxConnection by connecting to the specified usbmuxd Socket.
func NewUsbMuxConnectionToSocket(socket string) *MuxConnection {
	var conn MuxConnection
	conn.tag = 0
	conn.ResponseChannel = make(chan []byte)
	conn.singleDecode = false
	conn.deviceConn = NewDeviceConnection(socket)
	conn.deviceConn.Connect(&conn)
	return &conn
}

//NewUsbMuxServerConnection creates a new MuxConnection in listening mode for proxy use.
func NewUsbMuxServerConnection(c net.Conn) *MuxConnection {
	var conn MuxConnection
	conn.tag = 0
	conn.singleDecode = true
	conn.decodeSignal = make(chan interface{})
	conn.stopSignal = make(chan interface{})
	conn.ResponseChannel = make(chan []byte)
	conn.deviceConn = NewDeviceConnection("")
	conn.deviceConn.Listen(&conn, c)
	return &conn
}

// NewUsbMuxConnectionWithDeviceConnection creates a new MuxConnection with from an already initialized DeviceConnectionInterface
// (only needed for testing)
func NewUsbMuxConnectionWithDeviceConnection(deviceConn DeviceConnectionInterface) *MuxConnection {
	var conn MuxConnection
	conn.tag = 0
	conn.singleDecode = false
	conn.ResponseChannel = make(chan []byte)
	deviceConn.Connect(&conn)
	conn.deviceConn = deviceConn
	return &conn
}

//Close closes the underlying socket connection.
func (muxConn *MuxConnection) Close() {
	muxConn.deviceConn.Close()
}

type usbmuxHeader struct {
	Length  uint32
	Version uint32
	Request uint32
	Tag     uint32
}

func newUsbmuxHeader(length uint32, tag uint32) *usbmuxHeader {
	header := usbmuxHeader{}
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
func (muxConn *MuxConnection) Send(msg interface{}) {
	muxConn.deviceConn.Send(msg)
}

//Encode serializes a MuxMessage struct to a Plist and returns the []byte of its
//string representation
func (muxConn *MuxConnection) Encode(message interface{}) ([]byte, error) {
	log.Debug("UsbMux send", reflect.TypeOf(message), " on ", &muxConn.deviceConn)
	stringContent := ToPlist(message)
	var err error
	var buffer bytes.Buffer

	headerBytes := getHeader(len(stringContent), muxConn.tag)
	buffer.Write(headerBytes)
	_, err = buffer.Write([]byte(stringContent))
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (muxConn *MuxConnection) StartDecode() {
	var i interface{}
	muxConn.decodeSignal <- i
}
func (muxConn *MuxConnection) StopDecoding() {
	var i interface{}
	muxConn.stopSignal <- i
}

//Decode reads all bytes for the next MuxMessage from r io.Reader and
//sends them to the ResponseChannel
func (muxConn MuxConnection) Decode(r io.Reader) error {
	if r == nil {
		muxConn.ResponseChannel <- nil
		return nil
	}
	if muxConn.singleDecode {
		select {
		case <-muxConn.stopSignal:
			return nil
		case <-muxConn.decodeSignal:
			log.Info("usbmux codec rcv decode")
		}
	}

	var muxHeader usbmuxHeader

	err := binary.Read(r, binary.LittleEndian, &muxHeader)
	if err != nil {
		return err
	}

	payloadBytes := make([]byte, muxHeader.Length-16)
	n, err := io.ReadFull(r, payloadBytes)
	if err != nil {
		return fmt.Errorf("Error '%s' while reading usbmux package. Only %d bytes received instead of %d", err.Error(), n, muxHeader.Length-16)
	}
	log.Debug("UsbMux Receive on ", &muxConn.deviceConn)
	muxConn.ResponseChannel <- payloadBytes
	return nil
}
