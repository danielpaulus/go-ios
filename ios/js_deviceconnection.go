package ios

import "io"

/**
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
*/

type JSDeviceConnection struct {
}

func (connection JSDeviceConnection) Send(message []byte) error {
	_, err := connection.Writer().Write(message)
	return err
}

func (connection JSDeviceConnection) Writer() io.Writer {

}
