package ios

import (
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

//DeviceConnection wraps the net.Conn to the ios Device and has support for
//switching Codecs and enabling SSL
type DeviceConnection struct {
	c           net.Conn
	jsSocket    *js.Object
	jsTLSSocket *js.Object
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
	ws := js.Global.Call("require", "net").Call("createConnection", socketAddress)
	conn.jsSocket = ws
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
	jsTLSSocket := createClientTLSConn(conn.jsSocket, pairRecord)
	conn.jsTLSSocket = jsTLSSocket
	conn.c = newWSConn(jsTLSSocket)
	return nil
}

//EnableSessionSslHandshakeOnly enables SSL only for the Handshake and then falls back to plaintext
//DTX based services do that currently
func (conn *DeviceConnection) EnableSessionSslHandshakeOnly(pairRecord PairRecord) error {
	log.Errorf("handshake only not supported yet for gopherjs")
	return nil
}

func createClientTLSConn(jsSocket *js.Object, pairRecord PairRecord) *js.Object {
	jsTLS := js.Global.Call("require", "tls")
	tlsVersion := "TLSv1_method"

	tlsOpts := map[string]interface{}{}
	tlsOpts["secureProtocol"] = tlsVersion
	tlsOpts["key"] = pairRecord.HostPrivateKey
	tlsOpts["cert"] = pairRecord.HostCertificate
	secureContext := jsTLS.Call("createSecureContext", tlsOpts)

	opts := map[string]interface{}{}
	opts["rejectUnauthorized"] = false
	opts["secureContext"] = secureContext
	jsTLSSocket := jsTLS.Get("TLSSocket").New(jsSocket, opts)
	return jsTLSSocket
	/**


	import tls from 'tls';


	const TLS_VERSION = 'TLSv1_method';

	function upgradeToSSL (socket, key, cert) {
	  return new tls.TLSSocket(socket, {
	    rejectUnauthorized: false,
	    secureContext: tls.createSecureContext({
	      key,
	      cert,
	      secureProtocol: TLS_VERSION
	    })
	  });
	}
	export { upgradeToSSL };
	*/
}

func (conn *DeviceConnection) Conn() net.Conn {
	return conn.c
}
