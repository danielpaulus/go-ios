package usbmux

import (
	"crypto/tls"
	"io"
	"net"
	"reflect"

	log "github.com/sirupsen/logrus"
)

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
	SendForProtocolUpgradeSSL(muxConnection *MuxConnection, message interface{}, newCodec Codec, pairRecord PairRecord) []byte
	SendForSslUpgrade(lockDownConn *LockDownConnection, pairRecord PairRecord) StartSessionResponse
	Send(message interface{})
	Listen(activeCodec Codec, c net.Conn)
	WaitForDisconnect() error
	StopReadingAfterNextMessage()
	ResumeReadingWithNewCodec(codec Codec)
	SetCodec(codec Codec)
	EnableSessionSsl(pairRecord PairRecord)
	ResumeReading()
}

//DeviceConnection wraps the net.Conn to the ios Device and has support for
//switching Codecs and enabling SSL
type DeviceConnection struct {
	c                 net.Conn
	activeCodec       Codec
	stop              chan struct{}
	disconnectChannel chan error
	muxSocket         string
}

func NewDeviceConnection(socketToConnectTo string) *DeviceConnection {
	return &DeviceConnection{muxSocket: socketToConnectTo}
}

//Connect connects to the USB multiplexer daemon using  the default address: '/var/run/usbmuxd'
func (conn *DeviceConnection) Connect(activeCodec Codec) {
	conn.ConnectToSocketAddress(activeCodec, conn.muxSocket)
}

func (conn *DeviceConnection) Listen(activeCodec Codec, c net.Conn) {
	conn.stop = make(chan struct{})
	conn.c = c
	conn.activeCodec = activeCodec
	conn.startReading()
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
	_, err = conn.c.Write(bytes)
	if err != nil {
		log.Fatalf("Failed sending: %s", err)
	}
}

func reader(conn *DeviceConnection) {
	for {
		err := conn.activeCodec.Decode(conn.c)
		select {
		case <-conn.stop:
			//ignore error for stopped connection, we stop reading for protocol upgrades
			return
		default:
			if err != nil {
				log.Info("Connection disconnected")
				conn.activeCodec.Decode(nil)
				conn.disconnectChannel <- err
			}
		}
	}
}

//WaitForDisconnect blocks until the connection disconnects and returns the error that caused the disconnect
func (conn *DeviceConnection) WaitForDisconnect() error {
	reason := <-conn.disconnectChannel
	return reason
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

//SendForProtocolUpgradeSSL does the same as SendForProtocolUpgrade and in addition to that enables SSL on the service connection.
func (conn *DeviceConnection) SendForProtocolUpgradeSSL(muxConnection *MuxConnection, message interface{}, newCodec Codec, pairRecord PairRecord) []byte {
	log.Debug("Protocol update to ", reflect.TypeOf(newCodec), " on ", &conn.c)
	conn.stopReadingAfterNextMessage()
	conn.Send(message)
	responseBytes := <-muxConnection.ResponseChannel
	conn.activeCodec = newCodec
	conn.EnableSessionSsl(pairRecord)
	conn.startReading()
	return responseBytes
}

func (conn *DeviceConnection) SetCodecAfterNextMessage(newCodec Codec, channel chan []byte) []byte {
	conn.stopReadingAfterNextMessage()
	msg := <-channel
	conn.activeCodec = newCodec
	conn.startReading()
	return msg
}

func (conn *DeviceConnection) StopReadingAfterNextMessage() {
	conn.stopReadingAfterNextMessage()
}
func (conn *DeviceConnection) ResumeReadingWithNewCodec(codec Codec) {
	conn.activeCodec = codec
	conn.startReading()
}
func (conn *DeviceConnection) SetCodec(codec Codec) {
	conn.activeCodec = codec
}

func (conn *DeviceConnection) stopReadingAfterNextMessage() {
	var sig struct{}
	go func() { conn.stop <- sig }()
}

func (conn *DeviceConnection) startReading() {
	go reader(conn)
}

func (conn *DeviceConnection) ResumeReading() {
	conn.startReading()
}

//SendForSslUpgrade Start Session and enable SSL
func (conn *DeviceConnection) SendForSslUpgrade(lockDownConn *LockDownConnection, pairRecord PairRecord) StartSessionResponse {
	conn.stopReadingAfterNextMessage()
	conn.Send(newStartSessionRequest(pairRecord.HostID, pairRecord.SystemBUID))
	resp := <-lockDownConn.ResponseChannel
	response := startSessionResponsefromBytes(resp)
	lockDownConn.sessionID = response.SessionID
	if response.EnableSessionSSL {
		conn.EnableSessionSsl(pairRecord)
		conn.startReading()
	}
	return response
}

func (conn *DeviceConnection) EnableSessionSsl(pairRecord PairRecord) {
	cert5, error5 := tls.X509KeyPair(pairRecord.HostCertificate, pairRecord.HostPrivateKey)
	if error5 != nil {
		log.Error("Error SSL:" + error5.Error())
		return
	}
	conf := &tls.Config{
		//We always trust whatever the phone sends, I do not see an issue here as probably
		//nobody would build a fake iphone to hack this library.
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert5},
		ClientAuth:         tls.NoClientCert,
	}

	tlsConn := tls.Client(conn.c, conf)
	log.Debug("enable session ssl on", &conn.c, " and wrap with tlsConn", &tlsConn)
	conn.c = net.Conn(tlsConn)
}
