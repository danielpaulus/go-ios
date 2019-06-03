package usbmux_test

import (
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"usbmuxd/usbmux"

	log "github.com/sirupsen/logrus"
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
