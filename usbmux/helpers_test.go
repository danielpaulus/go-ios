package usbmux_test

import (
	"github.com/danielpaulus/go-ios/usbmux"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func CreateSocketFilePath(name string) (string, func()) {
	path, err := ioutil.TempDir("", "goios_")
	if err != nil {
		log.Fatal(err)
	}
	socketFilePath := filepath.Join(path, name)
	cleanup := func() {
		os.RemoveAll(path)
	}
	return socketFilePath, cleanup
}

func StartMuxServer(path string, receivedMessages chan []byte, sendMessage chan []byte) func() {
	// listen on all interfaces
	codec := usbmux.MuxConnection{}
	codec.ResponseChannel = make(chan []byte)

	ln, err := net.Listen("unix", path)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		// accept connection on port
		conn, _ := ln.Accept()
		go func() {
			for {
				msg := <-sendMessage
				_, err := conn.Write(msg)
				if err != nil {
					log.Error(err)
				}
			}
		}()
		// run loop forever (or until ctrl-c)
		for {
			// will listen for message to process ending in newline (\n)
			go func() { codec.Decode(conn) }()
			message := <-codec.ResponseChannel
			receivedMessages <- message

		}
	}()
	return func() { ln.Close() }
}

func StartServer(path string, receivedMessages chan []byte, sendMessage chan []byte) func() {
	// listen on all interfaces

	ln, err := net.Listen("unix", path)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		// accept connection on port
		conn, _ := ln.Accept()
		go func() {
			for {
				msg := <-sendMessage
				_, err := conn.Write(msg)
				if err != nil {
					log.Error(err)
				}
			}
		}()
		// run loop forever (or until ctrl-c)
		for {
			// will listen for message to process ending in newline (\n)
			buffer := make([]byte, 1)
			_, err := conn.Read(buffer)
			if err != nil {
				log.Error(err)
			}
			receivedMessages <- buffer

		}
	}()
	return func() { ln.Close() }
}

func GenericMockUsbmuxdIntegrationTest(t *testing.T, commandToInvoke func() interface{}, whatUsbmuxShouldReceive interface{}, whatTheServerShouldRespond interface{}) interface{} {
	path, cleanup := CreateSocketFilePath("socket")
	defer cleanup()
	serverReceiver := make(chan []byte)
	serverSender := make(chan []byte)
	serverCleanup := StartMuxServer(path, serverReceiver, serverSender)
	defer serverCleanup()

	usbmux.UsbmuxdSocket = path

	returnValue := make(chan interface{})
	go func() {
		list := commandToInvoke()
		returnValue <- list
	}()
	serverHasReceived := <-serverReceiver

	readDevicesPlist := usbmux.ToPlist(whatUsbmuxShouldReceive)

	assert.Equal(t, readDevicesPlist, string(serverHasReceived))

	muxCodec := usbmux.MuxConnection{}

	bytes, err := muxCodec.Encode(whatTheServerShouldRespond)
	if assert.NoError(t, err) {
		serverSender <- bytes
		receivedList := <-returnValue
		return receivedList

	}
	log.Fatal("TestFailed")
	return nil
}
