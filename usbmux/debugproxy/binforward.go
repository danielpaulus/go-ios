package debugproxy

import (
	"encoding/hex"
	"io"

	"github.com/danielpaulus/go-ios/usbmux"
	log "github.com/sirupsen/logrus"
)

var servicesWithHandshakeOnlySSL = map[string]bool{"com.apple.instruments.remoteserver": true}

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
		p.log.Fatal("Failed sending muxmessage to device")
	}
	connectResponse, err := muxToDevice.ReadMessage()
	muxOnUnixSocket.SendMuxMessage(*connectResponse)

	binToDevice := BinDumpCodec{muxToDevice.Close()}
	binOnUnixSocket := BinDumpCodec{muxOnUnixSocket.Close()}
	if serviceInfo.UseSSL {
		if servicesWithHandshakeOnlySSL[serviceInfo.ServiceName] {
			binToDevice.deviceConn.EnableSessionSslHandshakeOnly(p.pairRecord)
			binOnUnixSocket.deviceConn.EnableSessionSslServerModeHandshakeOnly(p.pairRecord)
		} else {
			binToDevice.deviceConn.EnableSessionSsl(p.pairRecord)
			binOnUnixSocket.deviceConn.EnableSessionSslServerMode(p.pairRecord)
		}
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
			if err == io.EOF {
				p.LogClosed()
				return
			}
			p.log.Info("Failed reading bytes", err)
			return
		}
		p.log.WithFields(log.Fields{"direction": "host2device"}).Trace(hex.Dump(bytes))
		p.logBinaryMessageToDevice(bytes)
		binToDevice.Send(bytes)

	}
}

func proxyBinFromDeviceToHost(p *ProxyConnection, binOnUnixSocket BinDumpCodec, binToDevice BinDumpCodec) {
	byteCounter := 0
	for {
		bytes, err := binToDevice.ReadMessage()
		byteCounter += len(bytes)
		if byteCounter > 1024*500 {
			p.log.WithFields(log.Fields{"bytes": byteCounter}).Info("bytes transferred")
			byteCounter = 0
		}
		if err != nil {
			binOnUnixSocket.Close()
			binToDevice.Close()
			if err == io.EOF {
				p.LogClosed()
				return
			}
			p.log.Info("Failed reading bytes", err)
			return
		}
		p.log.WithFields(log.Fields{"direction": "device2host"}).Trace(hex.Dump(bytes))
		p.logBinaryMessageFromDevice(bytes)
		binOnUnixSocket.Send(bytes)
	}
}
