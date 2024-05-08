package debugproxy

import (
	"encoding/hex"
	"io"
	"path"

	ios "github.com/danielpaulus/go-ios/ios"
	log "github.com/sirupsen/logrus"
)

type serviceConfig struct {
	codec            func(string, string, *log.Entry) decoder
	handshakeOnlySSL bool
}

// serviceConfigurations stores info about which codec to use for which service by name.
// In addition, DTX based services only execute a SSL Handshake
// and then go back to sending unencrypted data right after the handshake.
var serviceConfigurations = map[string]serviceConfig{
	"com.apple.instruments.remoteserver":                      {NewDtxDecoder, true},
	"com.apple.accessibility.axAuditDaemon.remoteserver":      {NewDtxDecoder, true},
	"com.apple.testmanagerd.lockdown":                         {NewDtxDecoder, true},
	"com.apple.debugserver":                                   {NewBinDumpOnly, true},
	"com.apple.instruments.dtservicehub":                      {NewDtxDecoder, false},
	"com.apple.instruments.remoteserver.DVTSecureSocketProxy": {NewDtxDecoder, false},
	"com.apple.testmanagerd.lockdown.secure":                  {NewDtxDecoder, false},
	"bindumper":                                               {NewBinDumpOnly, false},
}

func getServiceConfigForName(serviceName string) serviceConfig {
	if val, ok := serviceConfigurations[serviceName]; ok {
		return val
	}
	return serviceConfigurations["bindumper"]
}

type BinaryForwardingProxy struct {
	deviceConn ios.DeviceConnectionInterface
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
	buffer := make([]byte, 50000)
	n, err := r.Read(buffer)
	if err != nil {
		return buffer[0:n], err
	}
	return buffer[0:n], nil
}

func handleConnectToService(connectRequest ios.UsbMuxMessage,
	decodedConnectRequest map[string]interface{},
	p *ProxyConnection,
	muxOnUnixSocket *ios.UsbMuxConnection,
	muxToDevice *ios.UsbMuxConnection,
	serviceInfo PhoneServiceInformation,
) {
	err := muxToDevice.SendMuxMessage(connectRequest)
	if err != nil {
		panic("Failed sending muxmessage to device")
	}
	connectResponse, err := muxToDevice.ReadMessage()
	if err != nil {
		panic("Failed reading muxmessage to device")
	}
	err = muxOnUnixSocket.SendMuxMessage(connectResponse)
	if err != nil {
		panic("Failed sending muxmessage to device")
	}

	serviceConfig := getServiceConfigForName(serviceInfo.ServiceName)
	binToDevice := BinaryForwardingProxy{muxToDevice.ReleaseDeviceConnection(), serviceConfig.codec(
		path.Join(p.info.ConnectionPath, "from-device.json"),
		path.Join(p.info.ConnectionPath, "from-device.bin"),
		p.log,
	)}
	binOnUnixSocket := BinaryForwardingProxy{muxOnUnixSocket.ReleaseDeviceConnection(), serviceConfig.codec(
		path.Join(p.info.ConnectionPath, "to-device.json"),
		path.Join(p.info.ConnectionPath, "to-device.bin"),
		p.log,
	)}

	if serviceInfo.UseSSL {
		if serviceConfig.handshakeOnlySSL {
			err = binToDevice.deviceConn.EnableSessionSslHandshakeOnly(p.pairRecord)
			if err != nil {
				panic("Failed enabling ssl EnableSessionSslHandshakeOnly")
			}
			binOnUnixSocket.deviceConn.EnableSessionSslServerModeHandshakeOnly(p.pairRecord)
		} else {
			err = binToDevice.deviceConn.EnableSessionSsl(p.pairRecord)
			if err != nil {
				panic("Failed EnableSessionSsl")
			}
			binOnUnixSocket.deviceConn.EnableSessionSslServerMode(p.pairRecord)
		}
	}
	proxyBinDumpConnection(p, binOnUnixSocket, binToDevice)
}

func proxyBinDumpConnection(p *ProxyConnection, binOnUnixSocket BinaryForwardingProxy, binToDevice BinaryForwardingProxy) {
	defer func() {
		log.Println("done") // Println executes normally even if there is a panic
		if x := recover(); x != nil {
			log.Printf("run time panic, moving back socket %v", x)
			err := MoveBack(ios.GetUsbmuxdSocket())
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Error("Failed moving back socket")
			}
			panic(x)
		}
	}()
	go proxyBinFromDeviceToHost(p, binOnUnixSocket, binToDevice)
	for {
		bytes, err := binOnUnixSocket.ReadMessage()
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("Failed readmessage bin unix sock")
		}
		binOnUnixSocket.decoder.decode(bytes)
		if err != nil && len(bytes) == 0 {
			binOnUnixSocket.Close()
			binToDevice.Close()
			if err == io.EOF {
				p.LogClosed()
				return
			}
			p.log.Errorf("Failed reading bytes %v", err)
			return
		}

		err = binToDevice.Send(bytes)
		if err != nil {
			log.Errorf("failed binforward sending to device: %v", err)
		}
	}
}

func proxyBinFromDeviceToHost(p *ProxyConnection, binOnUnixSocket BinaryForwardingProxy, binToDevice BinaryForwardingProxy) {
	defer func() {
		log.Println("done") // Println executes normally even if there is a panic
		if x := recover(); x != nil {
			log.Printf("run time panic, moving back socket %v", x)
			err := MoveBack(ios.ToUnixSocketPath(ios.GetUsbmuxdSocket()))
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Error("Failed moving back socket")
			}
			panic(x)
		}
	}()
	for {
		bytes, err := binToDevice.ReadMessage()
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Errorf("Failed binToDevice.ReadMessage b: %d", len(bytes))
		}
		binToDevice.decoder.decode(bytes)

		if err != nil && len(bytes) == 0 {
			binOnUnixSocket.Close()
			binToDevice.Close()
			if err == io.EOF {
				p.LogClosed()
				return
			}
			p.log.Errorf("Failed reading bytes %v", err)
			return
		}
		p.log.WithFields(log.Fields{"direction": "device2host"}).Trace(hex.Dump(bytes))
		err = binOnUnixSocket.Send(bytes)
		if err != nil {
			log.Errorf("failed binforward sending to host: %v", err)
		}
	}
}
