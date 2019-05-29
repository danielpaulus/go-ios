package usbmux

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

type DummyCodec struct {
	received chan []byte
	send     chan []byte
}

func TestDeviceConnection(t *testing.T) {
	path, cleanup := CreateSocketFilePath("socket")
	defer cleanup()
	received := make(chan []byte)
	send := make(chan []byte)
	serverCleanup := StartServer(path, received, send)
	defer serverCleanup()

	dummyCodec := DummyCodec{received: make(chan []byte), send: make(chan []byte)}
	var deviceConn DeviceConnection
	deviceConn.connectToSocketAddress(&dummyCodec, path)
	message := make([]byte, 1)
	go func() { deviceConn.send(message) }()
	encoderShouldEncode := <-dummyCodec.send
	assert.ElementsMatch(t, message, encoderShouldEncode)
	serverShouldHaveReceived := <-received
	assert.ElementsMatch(t, message, serverShouldHaveReceived)

	send <- message
	decoderShouldDecode := <-dummyCodec.received
	assert.ElementsMatch(t, message, decoderShouldDecode)
	deviceConn.close()
}

func (codec *DummyCodec) Encode(message interface{}) ([]byte, error) {
	codec.send <- message.([]byte)
	return message.([]byte), nil
}
func (codec *DummyCodec) Decode(r io.Reader) error {
	buffer := make([]byte, 1)
	_, _ = r.Read(buffer)
	codec.received <- buffer
	return nil
}
