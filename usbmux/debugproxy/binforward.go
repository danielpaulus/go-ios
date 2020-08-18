package debugproxy

import (
	"encoding/hex"

	"github.com/danielpaulus/go-ios/usbmux"
	log "github.com/sirupsen/logrus"
)

type BinDumpCodec struct {
	deviceConn usbmux.DeviceConnectionInterface
}

func (b BinDumpCodec) Close() {
	b.deviceConn.Close()
}

func (b BinDumpCodec) Send(msg []byte) error {
	return b.deviceConn.Send(msg)
}

func (b *BinDumpCodec) ReadMessage() ([]byte, error) {
	r := b.deviceConn.Reader()
	buffer := make([]byte, 1024)
	n, err := r.Read(buffer)
	if err != nil {
		return make([]byte, 0), err
	}
	return buffer[0:n], nil
}
func handleConnectToService(connectRequest *usbmux.MuxMessage,
	decodedConnectRequest map[string]interface{},
	p *ProxyConnection,
	muxOnUnixSocket *usbmux.MuxConnection,
	muxToDevice *usbmux.MuxConnection,
	serviceInfo PhoneServiceInformation) {
	err := muxToDevice.SendMuxMessage(*connectRequest)
	if err != nil {
		log.Fatal("Failed sending muxmessage to device")
	}
	connectResponse, err := muxToDevice.ReadMessage()
	muxOnUnixSocket.SendMuxMessage(*connectResponse)

	binToDevice := BinDumpCodec{muxToDevice.Close()}
	binOnUnixSocket := BinDumpCodec{muxOnUnixSocket.Close()}
	if serviceInfo.UseSSL {
		binToDevice.deviceConn.EnableSessionSsl(p.pairRecord)
		binOnUnixSocket.deviceConn.EnableSessionSslServerMode(p.pairRecord)
	}
	proxyBinDumpConnection(p, binOnUnixSocket, binToDevice)
}

func proxyBinDumpConnection(p *ProxyConnection, binOnUnixSocket BinDumpCodec, binToDevice BinDumpCodec) {
	go proxyBinFromDeviceToHost(p, binOnUnixSocket, binToDevice)
	for {
		bytes, err := binOnUnixSocket.ReadMessage()

		if err != nil {
			binOnUnixSocket.Close()
			binToDevice.Close()
			log.Info("Failed reading bytes", err)
			return
		}
		log.WithFields(log.Fields{"ID": p.id, "direction": "host2device"}).Trace(hex.Dump(bytes))
		p.logBinaryMessageToDevice(bytes)
		binToDevice.Send(bytes)

	}
}

func proxyBinFromDeviceToHost(p *ProxyConnection, binOnUnixSocket BinDumpCodec, binToDevice BinDumpCodec) {
	for {
		bytes, err := binToDevice.ReadMessage()

		if err != nil {
			binOnUnixSocket.Close()
			binToDevice.Close()
			log.Info("Failed reading bytes", err)
			return
		}
		log.WithFields(log.Fields{"ID": p.id, "direction": "device2host"}).Trace(hex.Dump(bytes))
		p.logBinaryMessageFromDevice(bytes)
		binOnUnixSocket.Send(bytes)

	}
}
