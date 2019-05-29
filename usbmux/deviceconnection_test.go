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
	//setup dummy server
	path, cleanup := CreateSocketFilePath("socket")
	defer cleanup()
	serverReceiver := make(chan []byte)
	serverSender := make(chan []byte)
	serverCleanup := StartServer(path, serverReceiver, serverSender)
	defer serverCleanup()
	dummyCodec := DummyCodec{received: make(chan []byte), send: make(chan []byte)}
	//setup device connection
	var deviceConn DeviceConnection
	deviceConn.connectToSocketAddress(&dummyCodec, path)

	//check that deviceconnection passes messages through the active
	//encoder
	message := make([]byte, 1)
	go func() { deviceConn.send(message) }()
	encoderShouldEncode := <-dummyCodec.send
	assert.ElementsMatch(t, message, encoderShouldEncode)
	serverShouldHaveReceived := <-serverReceiver
	assert.ElementsMatch(t, message, serverShouldHaveReceived)

	//check that deviceConnection correctly passes received messages through
	//the active decoder
	serverSender <- message
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
