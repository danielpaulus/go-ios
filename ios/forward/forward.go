package forward

import (
	"fmt"
	"io"
	"net"

	"github.com/danielpaulus/go-ios/ios"
	log "github.com/sirupsen/logrus"
)

type iosproxy struct {
	tcpConn    net.Conn
	deviceConn ios.DeviceConnectionInterface
}

//Forward forwards every connection made to the hostPort to whatever service runs inside an app on the device on phonePort.
func Forward(device ios.DeviceEntry, hostPort uint16, phonePort uint16) error {

	log.Infof("Start listening on port %d forwarding to port %d on device", hostPort, phonePort)
	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", hostPort))

	go connectionAccept(l, device.DeviceID, phonePort)

	if err != nil {
		return err
	}

	return nil
}

func connectionAccept(l net.Listener, deviceID int, phonePort uint16) {
	for {
		clientConn, err := l.Accept()
		if err != nil {
			log.Errorf("Error accepting new connection %v", err)
			continue
		}
		log.WithFields(log.Fields{"conn": fmt.Sprintf("%#v", clientConn)}).Info("new client connected")
		go startNewProxyConnection(clientConn, deviceID, phonePort)
	}
}

func startNewProxyConnection(clientConn net.Conn, deviceID int, phonePort uint16) {
	usbmuxConn, err := ios.NewUsbMuxConnectionSimple()
	if err != nil {
		log.Errorf("could not connect to usbmuxd: %+v", err)
		clientConn.Close()
		return
	}
	muxError := usbmuxConn.Connect(deviceID, phonePort)
	if muxError != nil {
		log.WithFields(log.Fields{"conn": fmt.Sprintf("%#v", clientConn), "err": muxError, "phonePort": phonePort}).Infof("could not connect to phone")
		clientConn.Close()
		return
	}
	log.WithFields(log.Fields{"conn": fmt.Sprintf("%#v", clientConn), "phonePort": phonePort}).Infof("Connected to port")
	deviceConn := usbmuxConn.ReleaseDeviceConnection()

	//proxyConn := iosproxy{clientConn, deviceConn}
	go func() {
		io.Copy(clientConn, deviceConn.Reader())
	}()
	go func() {
		io.Copy(deviceConn.Writer(), clientConn)
	}()
}

func (proxyConn *iosproxy) Close() {
	proxyConn.tcpConn.Close()
}
