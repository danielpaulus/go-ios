package usbmux

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"reflect"

	log "github.com/sirupsen/logrus"
)

type PlistCodec struct {
	ResponseChannel chan []byte
}

func NewPlistCodec(responseChannel chan []byte) *PlistCodec {
	var codec PlistCodec
	codec.ResponseChannel = responseChannel
	return &codec
}

//Encode encodes a LockDown Struct to a byte[] with the lockdown plist format.
//It returns a byte array that contains a 4 byte length unsigned big endian integer
//followed by the plist as a string
func (plistCodec *PlistCodec) Encode(message interface{}) ([]byte, error) {
	stringContent := ToPlist(message)
	log.Debug("Lockdown send", reflect.TypeOf(message))
	if log.IsLevelEnabled(log.TraceLevel) {
		print(stringContent)
	}
	buf := new(bytes.Buffer)
	length := len(stringContent)
	messageLength := uint32(length)

	err := binary.Write(buf, binary.BigEndian, messageLength)
	if err != nil {
		return nil, err
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
		return fmt.Errorf("lockdown Payload had incorrect size: %d original error: %s", n, err)
	}
	if log.IsLevelEnabled(log.TraceLevel) {
		println(hex.Dump(payloadBytes))
	}
	plistCodec.ResponseChannel <- payloadBytes
	return nil
}
