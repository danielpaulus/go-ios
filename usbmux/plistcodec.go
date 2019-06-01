package usbmux

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"reflect"

	log "github.com/sirupsen/logrus"
)

type PlistCodec struct {
	deviceConnection DeviceConnectionInterface
	ResponseChannel  chan []byte
}

func NewPlistCodecFromMuxConnection(muxConn *MuxConnection, responseChannel chan []byte) *PlistCodec {
	return NewPlistCodec(muxConn.deviceConn, responseChannel)
}

func NewPlistCodec(deviceConnection DeviceConnectionInterface, responseChannel chan []byte) *PlistCodec {
	var codec PlistCodec
	codec.deviceConnection = deviceConnection
	codec.ResponseChannel = responseChannel
	return &codec
}

//Encode encodes a LockDown Struct to a byte[] with the lockdown plist format.
//It returns a byte array that contains a 4 byte length unsigned big endian integer
//followed by the plist as a string
func (plistCodec *PlistCodec) Encode(message interface{}) ([]byte, error) {
	stringContent := ToPlist(message)
	log.Debug("Lockdown send", reflect.TypeOf(message), " on ", &plistCodec.deviceConnection)
	buf := new(bytes.Buffer)
	length := len(stringContent)
	messageLength := uint32(length)

	errs := binary.Write(buf, binary.BigEndian, messageLength)
	if errs != nil {
		return nil, errs
	}
	buf.Write([]byte(stringContent))
	return buf.Bytes(), nil
}

//Decode reads a Lockdown Message from the provided reader and
//sends it to the ResponseChannel
func (plistCodec *PlistCodec) Decode(r io.Reader) error {
	buf := make([]byte, 4)

	err := binary.Read(r, binary.BigEndian, buf)
	if err != nil {
		return err
	}
	length := binary.BigEndian.Uint32(buf)
	payloadBytes := make([]byte, length)
	n, err := io.ReadFull(r, payloadBytes)
	if err != nil {
		return err
	}
	if n != int(length) {
		return errors.New("lockdown Payload had incorrect size")
	}
	log.Debug("Lockdown receive on ", &plistCodec.deviceConnection)
	plistCodec.ResponseChannel <- payloadBytes
	return nil
}
