package ios

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"

	log "github.com/sirupsen/logrus"
)

// PlistCodec is a codec for PLIST based services with [4 byte big endian length][plist-payload] based messages
type PlistCodec struct{}

// NewPlistCodec create a codec for PLIST based services with [4 byte big endian length][plist-payload] based messages
func NewPlistCodec() PlistCodec {
	return PlistCodec{}
}

// Encode encodes a LockDown Struct to a byte[] with the lockdown plist format.
// It returns a byte array that contains a 4 byte length unsigned big endian integer
// followed by the plist as a string
func (plistCodec PlistCodec) Encode(message interface{}) ([]byte, error) {
	stringContent := ToPlist(message)
	log.Tracef("Lockdown send %v", reflect.TypeOf(message))
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

// Decode reads a Lockdown Message from the provided reader and
// sends it to the ResponseChannel
func (plistCodec PlistCodec) Decode(r io.Reader) ([]byte, error) {
	if r == nil {
		return nil, errors.New("Reader was nil")
	}
	buf := make([]byte, 4)
	err := binary.Read(r, binary.BigEndian, buf)
	if err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(buf)
	payloadBytes := make([]byte, length)
	n, err := io.ReadFull(r, payloadBytes)
	if err != nil {
		return nil, fmt.Errorf("lockdown Payload had incorrect size: %d expected: %d original error: %s", n, length, err)
	}
	return payloadBytes, nil
}
