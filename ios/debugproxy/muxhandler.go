package debugproxy

import (
	"bytes"
	"fmt"
	"io"

	ios "github.com/danielpaulus/go-ios/ios"
	log "github.com/sirupsen/logrus"
	"howett.net/plist"
)

func proxyUsbMuxConnection(p *ProxyConnection, muxOnUnixSocket *ios.UsbMuxConnection, muxToDevice *ios.UsbMuxConnection) {
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
		request, err := muxOnUnixSocket.ReadMessage()
		if err != nil {
			muxOnUnixSocket.ReleaseDeviceConnection().Close()
			muxToDevice.ReleaseDeviceConnection().Close()
			if err == io.EOF {
				p.LogClosed()
				return
			}
			p.log.Info("Failed reading UsbMuxMessage", err)
			return
		}

		var decodedRequest map[string]interface{}
		decoder := plist.NewDecoder(bytes.NewReader(request.Payload))
		err = decoder.Decode(&decodedRequest)
		if err != nil {
			p.log.Info("Failed decoding MuxMessage", request, err)
		}
		p.logJSONMessageToDevice(map[string]interface{}{"header": request.Header, "payload": decodedRequest, "type": "USBMUX"})

		p.log.WithFields(log.Fields{"ID": p.id, "direction": "host->device"}).Trace(decodedRequest)
		if decodedRequest["MessageType"] == "Connect" {
			handleConnect(request, decodedRequest, p, muxOnUnixSocket, muxToDevice)
			return
		}

		err = muxToDevice.SendMuxMessage(request)

		if decodedRequest["MessageType"] == "ReadPairRecord" {
			handleReadPairRecord(p, muxOnUnixSocket, muxToDevice)
			continue
		}
		if err != nil {
			panic(fmt.Sprintf("Failed forwarding message to device: %+v", request))
		}
		if decodedRequest["MessageType"] == "Listen" {
			handleListen(p, muxOnUnixSocket, muxToDevice)
			return
		}

		response, err := muxToDevice.ReadMessage()
		if err != nil {
			p.log.Error("Failed muxToDevice.ReadMessage()", request, err)
		}
		var decodedResponse map[string]interface{}
		decoder = plist.NewDecoder(bytes.NewReader(response.Payload))
		err = decoder.Decode(&decodedResponse)
		if err != nil {
			p.log.Error("Failed decoding MuxMessage", decodedResponse, err)
		}
		p.logJSONMessageFromDevice(map[string]interface{}{"header": response.Header, "payload": decodedResponse, "type": "USBMUX"})
		p.log.WithFields(log.Fields{"ID": p.id, "direction": "device->host"}).Trace(decodedResponse)
		err = muxOnUnixSocket.SendMuxMessage(response)
		if err != nil {
			p.log.Error("Failed muxOnUnixSocket.SendMuxMessage(response)", request, err)
		}
	}
}

func handleReadPairRecord(p *ProxyConnection, muxOnUnixSocket *ios.UsbMuxConnection, muxToDevice *ios.UsbMuxConnection) {
	response, err := muxToDevice.ReadMessage()
	var decodedResponse map[string]interface{}
	decoder := plist.NewDecoder(bytes.NewReader(response.Payload))
	err = decoder.Decode(&decodedResponse)
	if err != nil {
		p.log.Info("Failed decoding MuxMessage", decodedResponse, err)
	}
	pairRecord := ios.PairRecordfromBytes(decodedResponse["PairRecordData"].([]byte))
	pairRecord.DeviceCertificate = pairRecord.HostCertificate
	decodedResponse["PairRecordData"] = []byte(ios.ToPlist(pairRecord))
	newPayload := []byte(ios.ToPlist(decodedResponse))
	response.Payload = newPayload
	response.Header.Length = uint32(len(newPayload) + 16)
	p.logJSONMessageFromDevice(map[string]interface{}{"header": response.Header, "payload": decodedResponse, "type": "USBMUX"})
	p.log.WithFields(log.Fields{"ID": p.id, "direction": "device->host"}).Trace(decodedResponse)
	err = muxOnUnixSocket.SendMuxMessage(response)
}

func handleConnect(connectRequest ios.UsbMuxMessage, decodedConnectRequest map[string]interface{}, p *ProxyConnection, muxOnUnixSocket *ios.UsbMuxConnection, muxToDevice *ios.UsbMuxConnection) {
	var port uint16
	portFromPlist := decodedConnectRequest["PortNumber"]
	switch portFromPlist.(type) {
	case uint64:
		port = uint16(portFromPlist.(uint64))

	case int64:
		port = uint16(portFromPlist.(int64))
	}

	if port == ios.Lockdownport {
		p.log.Trace("Connect to Lockdown")
		handleConnectToLockdown(connectRequest, decodedConnectRequest, p, muxOnUnixSocket, muxToDevice)
	} else {
		info, err := p.debugProxy.retrieveServiceInfoByPort(ios.Ntohs(uint16(port)))
		if err != nil {
			panic(fmt.Sprintf("ServiceInfo for port: %d not found, this is a bug :-)reqheader: %+v repayload: %x", port, connectRequest.Header, connectRequest.Payload))
		}
		p.log.Infof("Connection to service '%s' detected on port %d", info.ServiceName, info.ServicePort)
		handleConnectToService(connectRequest, decodedConnectRequest, p, muxOnUnixSocket, muxToDevice, info)
	}
}

func handleConnectToLockdown(connectRequest ios.UsbMuxMessage, decodedConnectRequest map[string]interface{}, p *ProxyConnection, muxOnUnixSocket *ios.UsbMuxConnection, muxToDevice *ios.UsbMuxConnection) {
	err := muxToDevice.SendMuxMessage(connectRequest)
	if err != nil {
		panic("Failed sending muxmessage to device")
	}
	connectResponse, err := muxToDevice.ReadMessage()
	muxOnUnixSocket.SendMuxMessage(connectResponse)

	lockdownToDevice := ios.NewLockDownConnection(muxToDevice.ReleaseDeviceConnection())
	lockdownOnUnixSocket := ios.NewLockDownConnection(muxOnUnixSocket.ReleaseDeviceConnection())
	proxyLockDownConnection(p, lockdownOnUnixSocket, lockdownToDevice)
}

func handleListen(p *ProxyConnection, muxOnUnixSocket *ios.UsbMuxConnection, muxToDevice *ios.UsbMuxConnection) {
	go func() {
		// use this to detect when the conn is closed. There shouldn't be any messages received ever.
		_, err := muxOnUnixSocket.ReadMessage()
		if err == io.EOF {
			muxOnUnixSocket.ReleaseDeviceConnection().Close()
			muxToDevice.ReleaseDeviceConnection().Close()
			p.LogClosed()
			return
		}
		p.log.WithFields(log.Fields{"error": err}).Error("Unexpected error on read for LISTEN connection")
	}()

	for {
		response, err := muxToDevice.ReadMessage()
		if err != nil {
			// TODO: ugly, improve
			d := muxOnUnixSocket.ReleaseDeviceConnection()
			d1 := muxToDevice.ReleaseDeviceConnection()
			if d != nil {
				d.Close()
			}
			if d1 != nil {
				d1.Close()
			}

			p.LogClosed()
			return
		}
		var decodedResponse map[string]interface{}
		decoder := plist.NewDecoder(bytes.NewReader(response.Payload))
		err = decoder.Decode(&decodedResponse)
		if err != nil {
			p.log.Info("Failed decoding MuxMessage", decodedResponse, err)
		}
		p.logJSONMessageFromDevice(map[string]interface{}{"header": response.Header, "payload": decodedResponse, "type": "USBMUX"})
		p.log.WithFields(log.Fields{"ID": p.id, "direction": "device->host"}).Trace(decodedResponse)
		err = muxOnUnixSocket.SendMuxMessage(response)
		if err != nil {
			p.log.Info("Failed muxOnUnixSocket.SendMuxMessage(response)", decodedResponse, err)
		}
	}
}
