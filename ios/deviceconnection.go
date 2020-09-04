package ios

import (
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
	"io"
	"net"

	log "github.com/sirupsen/logrus"
)

// DeviceConnectionInterface contains a physical network connection to a usbmuxd socket.
type DeviceConnectionInterface interface {
	ConnectToSocketAddress(socketAddress string)
	Close()
	Send(message []byte) error
	Reader() io.Reader
	Writer() io.Writer
	EnableSessionSsl(pairRecord PairRecord) error
	EnableSessionSslServerMode(pairRecord PairRecord)
	EnableSessionSslHandshakeOnly(pairRecord PairRecord) error
	EnableSessionSslServerModeHandshakeOnly(pairRecord PairRecord)
	DisableSessionSSL()
}

//DeviceConnection wraps the net.Conn to the ios Device and has support for
//switching Codecs and enabling SSL
type DeviceConnection struct {
	c               net.Conn
	unencryptedConn net.Conn
}

//NewDeviceConnection creates a new DeviceConnection pointing to the given socket waiting for a call to Connect()
func NewDeviceConnection(socketToConnectTo string) *DeviceConnection {
	conn := &DeviceConnection{}
	conn.ConnectToSocketAddress(socketToConnectTo)
	return conn
}

//NewDeviceConnectionWithConn create a DeviceConnection with a already connected network conn.
func NewDeviceConnectionWithConn(conn net.Conn) *DeviceConnection {
	return &DeviceConnection{c: conn}
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
	n, err := conn.c.Write(bytes)
	if n < len(bytes) {
		log.Fatalf("DeviceConnection failed writing %d bytes, only %d sent", len(bytes), n)
	}
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

//DisableSessionSSL is a hack to go back from SSL to an unencrypted conn without closing the connection.
//It is only used for the debug proxy because certain MAC applications actually disable SSL, use the connection
//to send unencrypted messages just to then enable SSL again without closing the connection
func (conn *DeviceConnection) DisableSessionSSL() {
	/*
		Sometimes, apple tools will remove SSL from a lockdown connection after StopSession was received.
		After that they will issue a StartSession command on the same connection in plaintext just to then enable SSL again.
		I only know of Accessibility Inspector doing this, but there might be other tools too.
		This is not really supported by any library afaik so I added this hack to make it work.
	*/

	//First send a close write
	conn.c.(*tls.Conn).CloseWrite()
	//Use the underlying conn again to receive unencrypted bytes
	conn.c = conn.unencryptedConn
	/*read the first 5 bytes of the SSL encrypted CLOSE message we get.
	Because it is a Close message, we can throw it away. We cannot forward it to the client though, because
	we use a different SSL connection there.
	First five bytes are usually: 15 03 03 XX XX where XX XX is the length of the encrypted payload
	*/
	header := make([]byte, 5)

	io.ReadFull(conn.c, header)
	log.Trace(hex.Dump(header))
	length := binary.BigEndian.Uint16(header[3:])
	payload := make([]byte, length)

	io.ReadFull(conn.c, payload)
	log.Trace(hex.Dump(payload))

}

//EnableSessionSslServerMode wraps the underlying net.Conn in a server tls.Conn using the pairRecord.
func (conn *DeviceConnection) EnableSessionSslServerMode(pairRecord PairRecord) {
	tlsConn, _ := conn.createServerTlsConn(pairRecord)
	conn.unencryptedConn = conn.c
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
	conn.unencryptedConn = conn.c
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
