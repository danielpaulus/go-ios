package usbmux

import (
	"io"
	"net"
	"os"
	"reflect"

	log "github.com/sirupsen/logrus"
)

const usbmuxdSocket = "/var/run/usbmuxd"

var usbmuxdSocketEnv = os.Getenv("USBMUXD_SOCKET_ADDRESS")

//Codec is an interface with methods to Encode and Decode iOS Messages for all different protocols.
type Codec interface {
	//Encode converts a given message to a byte array
	Encode(interface{}) ([]byte, error)
	//Decode will be called by a DeviceConnection and provide it with a io.Reader to read raw bytes from.
	Decode(io.Reader) error
}

//DeviceConnection wraps the net.Conn to the ios Device and has support for
//switching Codecs and enabling SSL
type DeviceConnection struct {
	c           net.Conn
	activeCodec Codec
	stop        chan struct{}
}

//connect connects to /var/run/usbmuxd by default or to whatever is specified in the
// "USBMUXD_SOCKET_ADDRESS" env var.
func (conn *DeviceConnection) connect(activeCodec Codec) {
	var address = usbmuxdSocket
	if usbmuxdSocketEnv != "" {
		address = usbmuxdSocketEnv
	}

	c, err := net.Dial("unix", address)
	if err != nil {
		log.Fatal("Could not connect to usbmuxd socket, is it running?", err)
	}
	log.Debug("Opening connection:", &c)
	conn.stop = make(chan struct{})
	conn.c = c
	conn.activeCodec = activeCodec
	conn.startReading()

}

func (conn *DeviceConnection) close() {
	log.Debug("Closing connection:", &conn.c)
	var sig struct{}
	go func() { conn.stop <- sig }()
	conn.c.Close()
}

func (conn *DeviceConnection) send(message interface{}) {
	bytes, err := conn.activeCodec.Encode(message)
	if err != nil {
		log.Errorf("Deviceconnection failed sending data %s", err)
		conn.close()
		return
	}
	conn.c.Write(bytes)
}

func reader(conn *DeviceConnection) {
	for {
		err := conn.activeCodec.Decode(conn.c)
		select {
		case <-conn.stop:
			//ignore error for stopped connection
			return
		default:
			if err != nil {
				log.Errorf("Failed decoding/reading %s", err)
				conn.close()
				return
			}
		}
	}

}

//sendForProtocolUpgrade takes care of the complicated protocol upgrade process of iOS/Usbmux.
//First, a Connect Message is sent to usbmux using the UsbMux Codec
//Second, wait for the Mux Response also in UsbMuxCodec and stop reading immediately after receiving it
//since this is network connection, it could be that the MuxResponse is immediately followed by
//Data from the Codec. In that case, attempting a read with UsbMux usually results in fatal connection loss.
//To Prevent this, stop reading immediately after reading the response.
//Third, set the new codec and start reading again
//It returns the usbMuxResponse as a []byte
func (conn *DeviceConnection) sendForProtocolUpgrade(muxConnection *MuxConnection, message interface{}, newCodec Codec) []byte {
	log.Debug("Protocol update to ", reflect.TypeOf(newCodec), " on ", &conn.c)
	conn.stopReadingAfterNextMessage()
	conn.send(message)
	responseBytes := <-muxConnection.ResponseChannel
	conn.activeCodec = newCodec
	conn.startReading()
	return responseBytes
}

func (conn *DeviceConnection) stopReadingAfterNextMessage() {
	var sig struct{}
	go func() { conn.stop <- sig }()
}

func (conn *DeviceConnection) startReading() {
	go reader(conn)
}
