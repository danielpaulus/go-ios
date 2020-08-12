package usbmux

import (
	"bytes"
	"net"

	"github.com/danielpaulus/go-ios/usbmux/proxy_utils"
	log "github.com/sirupsen/logrus"
	"howett.net/plist"
)

//DebugProxy can be used to dump and modify communication between mac and host
type DebugProxy struct{}

//NewDebugProxy creates a new Default proxy
func NewDebugProxy() *DebugProxy {
	return &DebugProxy{}
}

//Launch moves the original /var/run/usbmuxd to /var/run/usbmuxd.real and starts the server at /var/run/usbmuxd
func (d *DebugProxy) Launch() error {
	originalSocket, err := proxy_utils.MoveSock(DefaultUsbmuxdSocket)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "socket": DefaultUsbmuxdSocket}).Error("Unable to move, lacking permissions?")
		return err
	}

	listener, err := net.Listen("unix", DefaultUsbmuxdSocket)
	if err != nil {
		log.Fatal("Could not listen on usbmuxd socket, do I have access permissions?", err)
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Errorf("error with connection: %e", err)
		}
		serverConn := NewUsbMuxServerConnection(conn)
		clientConn := NewUsbMuxConnectionToSocket(originalSocket)

		go func() {
			for {
				msg := <-serverConn.ResponseChannel
				if msg == nil {
					log.Info("service on host disconnected")
					clientConn.Close()
					return
				}
				var decoded map[string]interface{}
				decoder := plist.NewDecoder(bytes.NewReader(msg))
				err := decoder.Decode(&decoded)
				if err != nil {
					log.Info(err)
				}

				log.Info(decoded)
				clientConn.Send(decoded)
				//just one way
			}
		}()
		go func() {
			for {
				msg := <-clientConn.ResponseChannel
				if msg == nil {
					log.Info("device disconnected")
					serverConn.Close()
					return
				}
				var decoded map[string]interface{}
				decoder := plist.NewDecoder(bytes.NewReader(msg))
				err := decoder.Decode(&decoded)
				if err != nil {
					log.Info(err)
				}

				log.Info(decoded)

				serverConn.Send(decoded)
			}
		}()

	}

}

//Close moves /var/run/usbmuxd.real back to /var/run/usbmuxd and disconnects all active proxy connections
func (d *DebugProxy) Close() {
	log.Info("Moving back original socket")
	err := proxy_utils.MoveBack(DefaultUsbmuxdSocket)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed moving back socket")
	}
}
