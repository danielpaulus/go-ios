package debugproxy

import (
	"encoding/hex"
	"io"
	"path"

	"github.com/danielpaulus/go-ios/usbmux"
	log "github.com/sirupsen/logrus"
)

type serviceConfig struct {
	codec            func(string, string) decoder
	handshakeOnlySSL bool
}

var serviceConfigurations = map[string]serviceConfig{
	"com.apple.instruments.remoteserver": {NewDtxDecoder, true},
	"bindumper":                          {NewBinDumpOnly, false},
}

func getServiceConfigForName(serviceName string) serviceConfig {
	if val, ok := serviceConfigurations[serviceName]; ok {
		return val
	}
	return serviceConfigurations["bindumper"]
}

type BinaryForwardingProxy struct {
	deviceConn usbmux.DeviceConnectionInterface
	decoder    decoder
}

func (b BinaryForwardingProxy) Close() {
	b.deviceConn.Close()
}

func (b BinaryForwardingProxy) Send(msg []byte) error {
	return b.deviceConn.Send(msg)
}

func (b *BinaryForwardingProxy) ReadMessage() ([]byte, error) {
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

	serviceConfig := getServiceConfigForName(serviceInfo.ServiceName)
	binToDevice := BinaryForwardingProxy{muxToDevice.Close(), serviceConfig.codec(
		path.Join(p.info.ConnectionPath, "from-device.json"),
		path.Join(p.info.ConnectionPath, "from-device.bin"),
	)}
	binOnUnixSocket := BinaryForwardingProxy{muxOnUnixSocket.Close(), serviceConfig.codec(
		path.Join(p.info.ConnectionPath, "to-device.json"),
		path.Join(p.info.ConnectionPath, "to-device.bin"),
	)}

	if serviceInfo.UseSSL {
		if serviceConfig.handshakeOnlySSL {
			binToDevice.deviceConn.EnableSessionSslHandshakeOnly(p.pairRecord)
			binOnUnixSocket.deviceConn.EnableSessionSslServerModeHandshakeOnly(p.pairRecord)
		} else {
			binToDevice.deviceConn.EnableSessionSsl(p.pairRecord)
			binOnUnixSocket.deviceConn.EnableSessionSslServerMode(p.pairRecord)
		}
	}
	proxyBinDumpConnection(p, binOnUnixSocket, binToDevice)
}

func proxyBinDumpConnection(p *ProxyConnection, binOnUnixSocket BinaryForwardingProxy, binToDevice BinaryForwardingProxy) {
	go proxyBinFromDeviceToHost(p, binOnUnixSocket, binToDevice)
	for {
		bytes, err := binOnUnixSocket.ReadMessage()
		binOnUnixSocket.decoder.decode(bytes)
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

		binToDevice.Send(bytes)
	}
}

func proxyBinFromDeviceToHost(p *ProxyConnection, binOnUnixSocket BinaryForwardingProxy, binToDevice BinaryForwardingProxy) {
	for {
		bytes, err := binToDevice.ReadMessage()
		binToDevice.decoder.decode(bytes)

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
		binOnUnixSocket.Send(bytes)
	}
}
