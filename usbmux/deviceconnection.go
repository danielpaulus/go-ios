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
	Send(message []byte) error
	Reader() io.Reader
	Writer() io.Writer
	EnableSessionSsl(pairRecord PairRecord) error
	EnableSessionSslServerMode(pairRecord PairRecord)
	EnableSessionSslHandshakeOnly(pairRecord PairRecord) error
	EnableSessionSslServerModeHandshakeOnly(pairRecord PairRecord)
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

func NewDeviceConnectionWithConn(conn net.Conn) *DeviceConnection {
	return &DeviceConnection{muxSocket: "", c: conn}
}

//Connect connects to the USB multiplexer daemon using  the default address: '/var/run/usbmuxd'
func (conn *DeviceConnection) Connect() {
	conn.ConnectToSocketAddress(conn.muxSocket)
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
func (conn *DeviceConnection) Send(bytes []byte) error {
	_, err := conn.c.Write(bytes)
	if err != nil {
		log.Errorf("Failed sending: %s", err)
		conn.Close()
		return err
	}
	return nil
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
	tlsConn, _ := conn.createServerTlsConn(pairRecord)

	conn.c = net.Conn(tlsConn)
}

func (conn *DeviceConnection) EnableSessionSslServerModeHandshakeOnly(pairRecord PairRecord) {
	conn.createServerTlsConn(pairRecord)
}

//EnableSessionSsl wraps the underlying net.Conn in a client tls.Conn using the pairRecord.
func (conn *DeviceConnection) EnableSessionSsl(pairRecord PairRecord) error {
	tlsConn, err := conn.createClientTlsConn(pairRecord)
	if err != nil {
		return err
	}
	conn.c = net.Conn(tlsConn)
	return nil
}

func (conn *DeviceConnection) EnableSessionSslHandshakeOnly(pairRecord PairRecord) error {
	_, err := conn.createClientTlsConn(pairRecord)
	if err != nil {
		return err
	}
	return nil
}

func (conn *DeviceConnection) createClientTlsConn(pairRecord PairRecord) (*tls.Conn, error) {
	cert5, err := tls.X509KeyPair(pairRecord.HostCertificate, pairRecord.HostPrivateKey)
	if err != nil {
		log.Error("Error SSL:" + err.Error())
		return nil, err
	}
	conf := &tls.Config{
		//We always trust whatever the phone sends, I do not see an issue here as probably
		//nobody would build a fake iphone to hack this library.
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert5},
		ClientAuth:         tls.NoClientCert,
	}

	tlsConn := tls.Client(conn.c, conf)
	err = tlsConn.Handshake()
	if err != nil {
		log.Info("Handshake error", err)
		return nil, err
	}
	log.Debug("enable session ssl on", &conn.c, " and wrap with tlsConn", &tlsConn)
	return tlsConn, nil
}

func (conn *DeviceConnection) createServerTlsConn(pairRecord PairRecord) (*tls.Conn, error) {
	//we can just use the hostcert and key here, normally the device has its own pair of cert and key
	//but we do not know the device private key. funny enough, host has been signed by the same root cert
	//so it will be accepted by clients
	cert5, err := tls.X509KeyPair(pairRecord.HostCertificate, pairRecord.HostPrivateKey)
	if err != nil {
		log.Error("Error SSL:" + err.Error())
		return nil, err
	}
	conf := &tls.Config{
		//We always trust whatever the phone sends, I do not see an issue here as probably
		//nobody would build a fake iphone to hack this library.
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert5},
		ClientAuth:         tls.NoClientCert,
	}
	tlsConn := tls.Server(conn.c, conf)
	err = tlsConn.Handshake()
	if err != nil {
		log.Info("Handshake error", err)
		return nil, err
	}
	log.Debug("enable session ssl on", &conn.c, " and wrap with tlsConn", &tlsConn)
	return tlsConn, nil
}
