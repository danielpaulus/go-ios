package usbmux

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"reflect"

	log "github.com/sirupsen/logrus"
)

//MuxConnection provides a Send Method for sending Messages to UsbMuxD and a ResponseChannel to
//receive the responses.
type MuxConnection struct {
	tag             uint32
	deviceConn      *DeviceConnection
	ResponseChannel chan []byte
}

//NewUsbMuxConnection creates a new MuxConnection by connecting to the usbmuxd Socket.
func NewUsbMuxConnection() *MuxConnection {
	var conn MuxConnection
	var deviceConn DeviceConnection
	conn.tag = 0
	conn.deviceConn = &deviceConn
	deviceConn.connect(&conn)
	conn.ResponseChannel = make(chan []byte)
	return &conn
}

//Close closes the underlying socket connection.
func (muxConn *MuxConnection) Close() {
	muxConn.deviceConn.close()
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

func (muxConn *MuxConnection) Send(msg interface{}) {
	muxConn.deviceConn.send(msg)
}

//Encode serializes a MuxMessage struct to a Plist and returns the []byte of its
//string representation
func (muxConn *MuxConnection) Encode(message interface{}) ([]byte, error) {
	log.Debug("UsbMux send", reflect.TypeOf(message), " on ", &muxConn.deviceConn.c)
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

//Decode reads all bytes for the next MuxMessage from r io.Reader and
//sends them to the ResponseChannel
func (muxConn *MuxConnection) Decode(r io.Reader) error {
	var muxHeader usbmuxHeader

	err := binary.Read(r, binary.LittleEndian, &muxHeader)
	if err != nil {
		return err
	}

	payloadBytes := make([]byte, muxHeader.Length-16)
	n, err := io.ReadFull(r, payloadBytes)
	if err != nil {
		return err
	}
	if n != int(muxHeader.Length-16) {
		return errors.New("Invalid UsbMux Payload")
	}
	log.Debug("UsbMux Receive on ", &muxConn.deviceConn.c)
	muxConn.ResponseChannel <- payloadBytes
	return nil
}
