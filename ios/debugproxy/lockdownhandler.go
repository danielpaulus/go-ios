package debugproxy

import (
	"bytes"
	"io"

	ios "github.com/danielpaulus/go-ios/ios"
	log "github.com/sirupsen/logrus"
	"howett.net/plist"
)

func proxyLockDownConnection(p *ProxyConnection, lockdownOnUnixSocket *ios.LockDownConnection, lockdownToDevice *ios.LockDownConnection) {
	for {
		request, err := lockdownOnUnixSocket.ReadMessage()
		if err != nil {
			lockdownOnUnixSocket.Close()
			lockdownToDevice.Close()
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
		p.log.WithFields(log.Fields{"ID": p.id, "direction": "host2device"}).Info(decodedRequest)

		err = lockdownToDevice.Send(decodedRequest)
		if err != nil {
			p.log.Errorf("Failed forwarding message to device: %x", request)
		}
		p.log.Info("done sending to device")
		response, err := lockdownToDevice.ReadMessage()
		if err != nil {
			log.Errorf("error reading from device: %+v", err)
			response, err = lockdownToDevice.ReadMessage()
			log.Infof("second read: %+v %+v", response, err)
		}

		var decodedResponse map[string]interface{}
		decoder = plist.NewDecoder(bytes.NewReader(response))
		err = decoder.Decode(&decodedResponse)
		if err != nil {
			p.log.Info("Failed decoding LockdownMessage", decodedResponse, err)
		}
		p.logJSONMessageFromDevice(map[string]interface{}{"payload": decodedResponse, "type": "LOCKDOWN"})
		p.log.WithFields(log.Fields{"ID": p.id, "direction": "device2host"}).Info(decodedResponse)

		err = lockdownOnUnixSocket.Send(decodedResponse)
		if err != nil {
			p.log.Info("Failed sending LockdownMessage from device to host service", decodedResponse, err)
		}
		if decodedResponse["EnableSessionSSL"] == true {
			lockdownToDevice.EnableSessionSsl(p.pairRecord)
			lockdownOnUnixSocket.EnableSessionSslServerMode(p.pairRecord)
		}
		if decodedResponse["Request"] == "StartService" && decodedResponse["Error"] == nil {

			useSSL := false
			if decodedResponse["EnableServiceSSL"] != nil {
				useSSL = decodedResponse["EnableServiceSSL"].(bool)
			}
			info := PhoneServiceInformation{
				ServicePort: uint16(decodedResponse["Port"].(uint64)),
				ServiceName: decodedResponse["Service"].(string),
				UseSSL:      useSSL,
			}

			p.log.Debugf("Detected Service Start:%+v", info)
			p.debugProxy.storeServiceInformation(info)

		}

		if decodedResponse["Request"] == "StopSession" {
			p.log.Info("Stop Session detected, disabling SSL")
			lockdownOnUnixSocket.DisableSessionSSL()
			lockdownToDevice.DisableSessionSSL()
		}
	}
}
