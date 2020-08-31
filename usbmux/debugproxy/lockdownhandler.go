package debugproxy

import (
	"bytes"
	"io"

	"github.com/danielpaulus/go-ios/usbmux"
	log "github.com/sirupsen/logrus"
	"howett.net/plist"
)

func proxyLockDownConnection(p *ProxyConnection, lockdownOnUnixSocket *usbmux.LockDownConnection, lockdownToDevice *usbmux.LockDownConnection) {
	for {
		request, err := lockdownOnUnixSocket.ReadMessage()
		if err != nil {
			lockdownOnUnixSocket.Close().Close()
			lockdownToDevice.Close().Close()
			if err == io.EOF {
				p.LogClosed()
				return
			}
			p.log.Info("Failed reading LockdownMessage", err)
			return
		}

		var decodedRequest map[string]interface{}
		decoder := plist.NewDecoder(bytes.NewReader(request))
		err = decoder.Decode(&decodedRequest)
		if err != nil {
			p.log.Info("Failed decoding LockdownMessage", request, err)
		}
		p.logJSONMessageToDevice(map[string]interface{}{"payload": decodedRequest, "type": "LOCKDOWN"})
		p.log.WithFields(log.Fields{"ID": p.id, "direction": "host2device"}).Trace(decodedRequest)

		err = lockdownToDevice.Send(decodedRequest)
		if err != nil {
			p.log.Fatal("Failed forwarding message to device", request)
		}

		response, err := lockdownToDevice.ReadMessage()
		var decodedResponse map[string]interface{}
		decoder = plist.NewDecoder(bytes.NewReader(response))
		err = decoder.Decode(&decodedResponse)
		if err != nil {
			p.log.Info("Failed decoding LockdownMessage", decodedResponse, err)
		}
		p.logJSONMessageFromDevice(map[string]interface{}{"payload": decodedResponse, "type": "LOCKDOWN"})
		p.log.WithFields(log.Fields{"ID": p.id, "direction": "device2host"}).Trace(decodedResponse)

		err = lockdownOnUnixSocket.Send(decodedResponse)
		if decodedResponse["EnableSessionSSL"] == true {
			lockdownToDevice.EnableSessionSsl(p.pairRecord)
			lockdownOnUnixSocket.EnableSessionSslServerMode(p.pairRecord)
		}
		if decodedResponse["Request"] == "StartService" {

			useSSL := false
			if decodedResponse["EnableServiceSSL"] != nil {
				useSSL = decodedResponse["EnableServiceSSL"].(bool)
			}
			info := PhoneServiceInformation{
				ServicePort: uint16(decodedResponse["Port"].(uint64)),
				ServiceName: decodedResponse["Service"].(string),
				UseSSL:      useSSL}

			p.log.Debugf("Detected Service Start:%s", info)
			p.debugProxy.storeServiceInformation(info)

		}

		if decodedResponse["Request"] == "StopSession" {
			lockdownOnUnixSocket.DisableSessionSSL()
			lockdownToDevice.DisableSessionSSL()
		}
	}
}
