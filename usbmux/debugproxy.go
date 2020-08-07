package usbmux

import (
	"github.com/danielpaulus/go-ios/usbmux/proxy_utils"
	log "github.com/sirupsen/logrus"
)

type DebugProxy struct{}

func NewDebugProxy() *DebugProxy {
	return &DebugProxy{}
}

func (d *DebugProxy) Launch() error {
	originalSocket, err := proxy_utils.MoveSock(DefaultUsbmuxdSocket)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "socket": DefaultUsbmuxdSocket}).Error("Unable to move, lacking permissions?")
		return err
	}
	serverConn := NewUsbMuxServerConnection(DefaultUsbmuxdSocket)
	clientConn := NewUsbMuxConnectionToSocket(originalSocket)
	for {
		msg := <-serverConn.ResponseChannel
		log.Info(msg)
		clientConn.Send(msg)
		//just one way
	}

}

func (d *DebugProxy) Close() {
	log.Info("Moving back original socket")
	err := proxy_utils.MoveBack(DefaultUsbmuxdSocket)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed moving back socket")
	}
}
