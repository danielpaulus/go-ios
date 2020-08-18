package debugproxy

import (
	"bytes"

	"github.com/danielpaulus/go-ios/usbmux"
	log "github.com/sirupsen/logrus"
	"howett.net/plist"
)

func proxyUsbMuxConnection(p *ProxyConnection, muxOnUnixSocket *usbmux.MuxConnection, muxToDevice *usbmux.MuxConnection) {
	for {
		request, err := muxOnUnixSocket.ReadMessage()
		if err != nil {
			muxOnUnixSocket.Close().Close()
			muxToDevice.Close().Close()
			log.Info("Failed reading UsbMuxMessage", err)
			return
		}

		var decodedRequest map[string]interface{}
		decoder := plist.NewDecoder(bytes.NewReader(request.Payload))
		err = decoder.Decode(&decodedRequest)
		if err != nil {
			log.Info("Failed decoding MuxMessage", request, err)
		}
		p.logJSONMessageFromDevice(map[string]interface{}{"header": request.Header, "payload": decodedRequest, "type": "USBMUX"})

		log.WithFields(log.Fields{"ID": p.id, "direction": "host->device"}).Trace(decodedRequest)
		if decodedRequest["MessageType"] == "Connect" {
			handleConnect(request, decodedRequest, p, muxOnUnixSocket, muxToDevice)
			return
		}

		err = muxToDevice.SendMuxMessage(*request)

		if decodedRequest["MessageType"] == "ReadPairRecord" {
			handleReadPairRecord(p, muxOnUnixSocket, muxToDevice)
			continue
		}
		if err != nil {
			log.Fatal("Failed forwarding message to device", request)
		}
		if decodedRequest["MessageType"] == "Listen" {
			handleListen(p, muxOnUnixSocket, muxToDevice)
			return
		}

		response, err := muxToDevice.ReadMessage()
		var decodedResponse map[string]interface{}
		decoder = plist.NewDecoder(bytes.NewReader(response.Payload))
		err = decoder.Decode(&decodedResponse)
		if err != nil {
			log.Info("Failed decoding MuxMessage", decodedResponse, err)
		}
		p.logJSONMessageToDevice(map[string]interface{}{"header": request.Header, "payload": decodedRequest, "type": "USBMUX"})
		log.WithFields(log.Fields{"ID": p.id, "direction": "device->host"}).Trace(decodedResponse)
		err = muxOnUnixSocket.SendMuxMessage(*response)
	}
}

func handleReadPairRecord(p *ProxyConnection, muxOnUnixSocket *usbmux.MuxConnection, muxToDevice *usbmux.MuxConnection) {
	response, err := muxToDevice.ReadMessage()
	var decodedResponse map[string]interface{}
	decoder := plist.NewDecoder(bytes.NewReader(response.Payload))
	err = decoder.Decode(&decodedResponse)
	if err != nil {
		log.Info("Failed decoding MuxMessage", decodedResponse, err)
	}
	pairRecord := usbmux.PairRecordfromBytes(decodedResponse["PairRecordData"].([]byte))
	pairRecord.DeviceCertificate = pairRecord.HostCertificate
	decodedResponse["PairRecordData"] = []byte(usbmux.ToPlist(pairRecord))
	newPayload := []byte(usbmux.ToPlist(decodedResponse))
	response.Payload = newPayload
	response.Header.Length = uint32(len(newPayload) + 16)
	log.WithFields(log.Fields{"ID": p.id, "direction": "device->host"}).Info(decodedResponse)
	err = muxOnUnixSocket.SendMuxMessage(*response)
}

func handleConnect(connectRequest *usbmux.MuxMessage, decodedConnectRequest map[string]interface{}, p *ProxyConnection, muxOnUnixSocket *usbmux.MuxConnection, muxToDevice *usbmux.MuxConnection) {
	log.Info("Detected Connect Message")

	var port int
	portFromPlist := decodedConnectRequest["PortNumber"]
	switch portFromPlist.(type) {
	case uint64:
		port = int(portFromPlist.(uint64))

	case int64:
		port = int(portFromPlist.(int64))
	}

	if int(port) == usbmux.Lockdownport {
		log.Info("Upgrading to Lockdown")
		handleConnectToLockdown(connectRequest, decodedConnectRequest, p, muxOnUnixSocket, muxToDevice)
	} else {
		info, err := p.debugProxy.retrieveServiceInfoByPort(usbmux.Ntohs(uint16(port)))
		if err != nil {
			log.Fatal("ServiceInfo for port not found, this is a bug :-)")
		}
		log.Info("Connection to service detected", info)
		handleConnectToService(connectRequest, decodedConnectRequest, p, muxOnUnixSocket, muxToDevice, info)
	}
}

func handleConnectToLockdown(connectRequest *usbmux.MuxMessage, decodedConnectRequest map[string]interface{}, p *ProxyConnection, muxOnUnixSocket *usbmux.MuxConnection, muxToDevice *usbmux.MuxConnection) {
	err := muxToDevice.SendMuxMessage(*connectRequest)
	if err != nil {
		log.Fatal("Failed sending muxmessage to device")
	}
	connectResponse, err := muxToDevice.ReadMessage()
	muxOnUnixSocket.SendMuxMessage(*connectResponse)

	lockdownToDevice := usbmux.NewLockDownConnection(muxToDevice.Close())
	lockdownOnUnixSocket := usbmux.NewLockDownConnection(muxOnUnixSocket.Close())
	proxyLockDownConnection(p, lockdownOnUnixSocket, lockdownToDevice)
}

func handleListen(p *ProxyConnection, muxOnUnixSocket *usbmux.MuxConnection, muxToDevice *usbmux.MuxConnection) {
	//TODO: we need a way to detect if the host closes the connection, otherwise this will stay open longer than needed
	for {
		response, err := muxToDevice.ReadMessage()
		var decodedResponse map[string]interface{}
		decoder := plist.NewDecoder(bytes.NewReader(response.Payload))
		err = decoder.Decode(&decodedResponse)
		if err != nil {
			log.Info("Failed decoding MuxMessage", decodedResponse, err)
		}
		log.WithFields(log.Fields{"ID": p.id, "direction": "device->host"}).Info(decodedResponse)
		err = muxOnUnixSocket.SendMuxMessage(*response)
	}
}
