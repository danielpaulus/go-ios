package ios

import (
	"crypto/tls"
	"github.com/gopherjs/gopherjs/js"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"runtime"
)

// DeviceConnectionInterface contains a physical network connection to a usbmuxd socket.
type DeviceConnectionInterface interface {
	Close() error
	Send(message []byte) error
	Reader() io.Reader
	Writer() io.Writer
	EnableSessionSsl(pairRecord PairRecord) error
	EnableSessionSslServerMode(pairRecord PairRecord)
	EnableSessionSslHandshakeOnly(pairRecord PairRecord) error
	EnableSessionSslServerModeHandshakeOnly(pairRecord PairRecord)
	DisableSessionSSL()
	Conn() net.Conn
}

func traceWS(ws *js.Object) {
	ws.Call("addEventListener", "open", func(evt *js.Object) {
		js.Global.Get("console").Call("log", "open", evt)
	})
	ws.Call("addEventListener", "message", func(evt *js.Object) {
		enc := js.Global.Get("TextDecoder").New()
		msg := enc.Call("decode", evt.Get("data"))

		js.Global.Get("console").Call("log", "message", msg)
	})
	ws.Call("addEventListener", "error", func(evt *js.Object) {
		js.Global.Get("console").Call("log", "error", evt)
	})
	ws.Call("addEventListener", "close", func(evt *js.Object) {
		js.Global.Get("console").Call("log", "close", evt)
	})
}

//DeviceConnection wraps the net.Conn to the ios Device and has support for
//switching Codecs and enabling SSL
type DeviceConnection struct {
	c net.Conn
}

//NewDeviceConnection creates a new DeviceConnection pointing to the given socket waiting for a call to Connect()
func NewDeviceConnection(socketToConnectTo string) (*DeviceConnection, error) {
	conn := &DeviceConnection{}
	return conn, conn.connectToSocketAddress(socketToConnectTo)
}

//NewDeviceConnectionWithConn create a DeviceConnection with a already connected network conn.
func NewDeviceConnectionWithConn(conn net.Conn) *DeviceConnection {
	panic("not supported in JS")
}

//ConnectToSocketAddress connects to the USB multiplexer with a specified socket addres
func (conn *DeviceConnection) connectToSocketAddress(socketAddress string) error {
	var network, address string
	switch runtime.GOOS {
	case "windows":
		network, address = "tcp", "127.0.0.1:27015"
	default:
		network, address = "unix", socketAddress
	}
	log.Info("not using", network, address)
	/*
		var client = net.createConnection("/tmp/mysocket");
	*/

	//ws := js.Global.Get("WebSocket").New("ws://" + host + ":5000/dial/" + port)
	ws := js.Module.Get("net").Call("createConnection", socketAddress)
	traceWS(ws) // OMIT
	conn.c = newWSConn(ws)

	return nil
}

//Close closes the network connection
func (conn *DeviceConnection) Close() error {
	log.Tracef("Closing connection: %v", &conn.c)
	return conn.c.Close()
}

//Send sends a message
func (conn *DeviceConnection) Send(bytes []byte) error {
	n, err := conn.c.Write(bytes)
	if n < len(bytes) {
		log.Errorf("DeviceConnection failed writing %d bytes, only %d sent", len(bytes), n)
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
	log.Errorf("dproxy not supported with gopherjs")
}

//EnableSessionSslServerMode wraps the underlying net.Conn in a server tls.Conn using the pairRecord.
func (conn *DeviceConnection) EnableSessionSslServerMode(pairRecord PairRecord) {
	log.Errorf("dproxy not supported with gopherjs")
}

//EnableSessionSslServerModeHandshakeOnly enables SSL only for the Handshake and then falls back to plaintext
//DTX based services do that currently. Server mode is needed only in the debugproxy.
func (conn *DeviceConnection) EnableSessionSslServerModeHandshakeOnly(pairRecord PairRecord) {
	log.Errorf("dproxy not supported with gopherjs")
}

//EnableSessionSsl wraps the underlying net.Conn in a client tls.Conn using the pairRecord.
func (conn *DeviceConnection) EnableSessionSsl(pairRecord PairRecord) error {
	tlsConn, err := conn.createClientTLSConn(pairRecord)
	if err != nil {
		return err
	}
	conn.c = net.Conn(tlsConn)
	return nil
}

//EnableSessionSslHandshakeOnly enables SSL only for the Handshake and then falls back to plaintext
//DTX based services do that currently
func (conn *DeviceConnection) EnableSessionSslHandshakeOnly(pairRecord PairRecord) error {
	_, err := conn.createClientTLSConn(pairRecord)
	if err != nil {
		return err
	}
	return nil
}

func (conn *DeviceConnection) createClientTLSConn(pairRecord PairRecord) (*tls.Conn, error) {
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

	log.Tracef("enable session ssl on %v and wrap with tlsConn: %v", &conn.c, &tlsConn)
	return tlsConn, nil
}

func (conn *DeviceConnection) Conn() net.Conn {
	return conn.c
}
