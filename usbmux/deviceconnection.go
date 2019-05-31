package usbmux

import (
	"io"
	"net"
	"reflect"

	log "github.com/sirupsen/logrus"
)

const usbmuxdSocket = "/var/run/usbmuxd"

//Codec is an interface with methods to Encode and Decode iOS Messages for all different protocols.
type Codec interface {
	//Encode converts a given message to a byte array
	Encode(interface{}) ([]byte, error)
	//Decode will be called by a DeviceConnection and provide it with a io.Reader to read raw bytes from.
	Decode(io.Reader) error
}

// DeviceConnectionInterface contains a physical network connection to a usbmuxd socket.
type DeviceConnectionInterface interface {
	Connect(activeCodec Codec)
	ConnectToSocketAddress(activeCodec Codec, socketAddress string)
	Close()
	SendForProtocolUpgrade(muxConnection *MuxConnection, message interface{}, newCodec Codec) []byte
	Send(message interface{})
}

//DeviceConnection wraps the net.Conn to the ios Device and has support for
//switching Codecs and enabling SSL
type DeviceConnection struct {
	c           net.Conn
	activeCodec Codec
	stop        chan struct{}
}

//Connect connects to the USB multiplexer daemon using  the default address: '/var/run/usbmuxd'
func (conn *DeviceConnection) Connect(activeCodec Codec) {
	conn.ConnectToSocketAddress(activeCodec, usbmuxdSocket)
}

//ConnectToSocketAddress connects to the USB multiplexer with a specified socket addres
func (conn *DeviceConnection) ConnectToSocketAddress(activeCodec Codec, socketAddress string) {
	c, err := net.Dial("unix", socketAddress)
	if err != nil {
		log.Fatal("Could not connect to usbmuxd socket, is it running?", err)
	}
	log.Debug("Opening connection:", &c)
	conn.stop = make(chan struct{})
	conn.c = c
	conn.activeCodec = activeCodec
	conn.startReading()
}

//Close closes the network connection
func (conn *DeviceConnection) Close() {
	log.Debug("Closing connection:", &conn.c)
	var sig struct{}
	go func() { conn.stop <- sig }()
	conn.c.Close()
}

//Send sends a message
func (conn *DeviceConnection) Send(message interface{}) {
	bytes, err := conn.activeCodec.Encode(message)
	if err != nil {
		log.Errorf("Deviceconnection failed sending data %s", err)
		conn.Close()
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
				conn.Close()
				return
			}
		}
	}

}

//SendForProtocolUpgrade takes care of the complicated protocol upgrade process of iOS/Usbmux.
//First, a Connect Message is sent to usbmux using the UsbMux Codec
//Second, wait for the Mux Response also in UsbMuxCodec and stop reading immediately after receiving it
//since this is network connection, it could be that the MuxResponse is immediately followed by
//Data from the Codec. In that case, attempting a read with UsbMux usually results in fatal connection loss.
//To Prevent this, stop reading immediately after reading the response.
//Third, set the new codec and start reading again
//It returns the usbMuxResponse as a []byte
func (conn *DeviceConnection) SendForProtocolUpgrade(muxConnection *MuxConnection, message interface{}, newCodec Codec) []byte {
	log.Debug("Protocol update to ", reflect.TypeOf(newCodec), " on ", &conn.c)
	conn.stopReadingAfterNextMessage()
	conn.Send(message)
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
