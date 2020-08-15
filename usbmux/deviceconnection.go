package usbmux

import (
	"crypto/tls"
	"io"
	"net"

	log "github.com/sirupsen/logrus"
)

// DeviceConnectionInterface contains a physical network connection to a usbmuxd socket.
type DeviceConnectionInterface interface {
	Connect()
	ConnectToSocketAddress(socketAddress string)
	Close()
	Send(message []byte)
	Listen(c net.Conn)
	Reader() io.Reader
	Writer() io.Writer
	EnableSessionSsl(pairRecord PairRecord) error
	EnableSessionSslServerMode(pairRecord PairRecord)
}

//DeviceConnection wraps the net.Conn to the ios Device and has support for
//switching Codecs and enabling SSL
type DeviceConnection struct {
	c         net.Conn
	muxSocket string
}

//NewDeviceConnection creates a new DeviceConnection pointing to the given socket waiting for a call to Connect()
func NewDeviceConnection(socketToConnectTo string) *DeviceConnection {
	return &DeviceConnection{muxSocket: socketToConnectTo}
}

//Connect connects to the USB multiplexer daemon using  the default address: '/var/run/usbmuxd'
func (conn *DeviceConnection) Connect() {
	conn.ConnectToSocketAddress(conn.muxSocket)
}

func (conn *DeviceConnection) Listen(c net.Conn) {
	conn.c = c
}

//ConnectToSocketAddress connects to the USB multiplexer with a specified socket addres
func (conn *DeviceConnection) ConnectToSocketAddress(socketAddress string) {
	c, err := net.Dial("unix", socketAddress)
	if err != nil {
		log.Fatal("Could not connect to usbmuxd socket, is it running?", err)
	}
	log.Debug("Opening connection:", &c)
	conn.c = c

}

//Close closes the network connection
func (conn *DeviceConnection) Close() {
	log.Debug("Closing connection:", &conn.c)
	conn.c.Close()
}

//Send sends a message
func (conn *DeviceConnection) Send(bytes []byte) {
	_, err := conn.c.Write(bytes)
	if err != nil {
		log.Errorf("Failed sending: %s", err)
		conn.Close()
	}
}

//Reader exposes the underlying net.Conn as io.Reader
func (conn *DeviceConnection) Reader() io.Reader {
	return conn.c
}

//Writer exposes the underlying net.Conn as io.Writer
func (conn *DeviceConnection) Writer() io.Writer {
	return conn.c
}

//EnableSessionSslServerMode wraps the underlying net.Conn in a server tls.Conn using the pairRecord.
func (conn *DeviceConnection) EnableSessionSslServerMode(pairRecord PairRecord) {
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
	tlsConn := tls.Server(conn.c, conf)
	err := tlsConn.Handshake()
	if err != nil {
		log.Info("Handshake error", err)
	}
	log.Debug("enable session ssl on", &conn.c, " and wrap with tlsConn", &tlsConn)
	conn.c = net.Conn(tlsConn)
}

//EnableSessionSsl wraps the underlying net.Conn in a client tls.Conn using the pairRecord.
func (conn *DeviceConnection) EnableSessionSsl(pairRecord PairRecord) error {
	cert5, error5 := tls.X509KeyPair(pairRecord.HostCertificate, pairRecord.HostPrivateKey)
	if error5 != nil {
		log.Error("Error SSL:" + error5.Error())
		return error5
	}
	conf := &tls.Config{
		//We always trust whatever the phone sends, I do not see an issue here as probably
		//nobody would build a fake iphone to hack this library.
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert5},
		ClientAuth:         tls.NoClientCert,
	}

	tlsConn := tls.Client(conn.c, conf)
	err := tlsConn.Handshake()
	if err != nil {
		log.Info("Handshake error", err)
	}
	log.Debug("enable session ssl on", &conn.c, " and wrap with tlsConn", &tlsConn)
	conn.c = net.Conn(tlsConn)
	return nil
}
