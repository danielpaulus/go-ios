package debugproxy

import (
	"bytes"

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
			log.Info("Failed reading LockdownMessage", err)
			return
		}

		var decodedRequest map[string]interface{}
		decoder := plist.NewDecoder(bytes.NewReader(request))
		err = decoder.Decode(&decodedRequest)
		if err != nil {
			log.Info("Failed decoding LockdownMessage", request, err)
		}

		log.WithFields(log.Fields{"ID": p.id, "direction": "host2device"}).Info(decodedRequest)

		err = lockdownToDevice.Send(decodedRequest)
		if err != nil {
			log.Fatal("Failed forwarding message to device", request)
		}

		response, err := lockdownToDevice.ReadMessage()
		var decodedResponse map[string]interface{}
		decoder = plist.NewDecoder(bytes.NewReader(response))
		err = decoder.Decode(&decodedResponse)
		if err != nil {
			log.Info("Failed decoding LockdownMessage", decodedResponse, err)
		}

		log.WithFields(log.Fields{"ID": p.id, "direction": "device2host"}).Info(decodedResponse)
		err = lockdownOnUnixSocket.Send(decodedResponse)
		if decodedResponse["EnableSessionSSL"] == true {
			lockdownToDevice.EnableSessionSsl(p.pairRecord)
			lockdownOnUnixSocket.EnableSessionSslServerMode(p.pairRecord)
		}
	}
}
