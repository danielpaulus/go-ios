package forward

/*
import (
	"fmt"
	"io"
	"net"

	"github.com/danielpaulus/go-ios/usbmux"
	log "github.com/sirupsen/logrus"
)

type iosproxy struct {
	tcpWriter net.Conn

	readBuffer []byte
}

//Forward forwards every connection made to the hostPort to whatever service runs inside an app on the device on phonePort.
func Forward(device usbmux.DeviceEntry, hostPort uint16, phonePort uint16) error {

	log.Infof("Start listening on port %d forwarding to port %d on device", hostPort, phonePort)
	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", hostPort))

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
			log.Fatal("Error accepting new connections.")
		}
		log.WithFields(log.Fields{"conn": fmt.Sprintf("%#v", clientConn)}).Info("new client connected")
		go startNewProxyConnection(clientConn, deviceID, phonePort)
	}
}

func startNewProxyConnection(clientConn net.Conn, deviceID int, phonePort uint16) {
	usbmuxConn := usbmux.NewUsbMuxConnection()
	//defer usbmuxConn.Close()
	var proxyConnection iosproxy

	buf := make([]byte, 4096)
	proxyConnection.readBuffer = buf

	proxyConnection.tcpWriter = clientConn
	muxError := usbmuxConn.Connect(deviceID, phonePort, &proxyConnection)
	if muxError != nil {
		log.WithFields(log.Fields{"conn": fmt.Sprintf("%#v", clientConn), "err": muxError, "phonePort": phonePort}).Infof("could not connect to phone")
		clientConn.Close()
		return
	}
	log.WithFields(log.Fields{"conn": fmt.Sprintf("%#v", clientConn), "phonePort": phonePort}).Infof("Connected to port")

	tcpbuf := make([]byte, 4096)

	go func() {
		for {
			n, err := clientConn.Read(tcpbuf)
			if err != nil {
				log.WithFields(log.Fields{"conn": fmt.Sprintf("%#v", clientConn)}).Info("Read Error, closing")
				clientConn.Close()
				return
			}
			usbmuxConn.Send(tcpbuf[:n])
		}
	}()
}

func (proxyConn *iosproxy) Close() {

}

func (proxyConn *iosproxy) Encode(message interface{}) ([]byte, error) {
	return message.([]byte), nil
}
func (proxyConn *iosproxy) Decode(r io.Reader) error {
	tcpbuf := make([]byte, 4096)
	readBytes, err := r.Read(tcpbuf)

	if err != nil {
		log.WithFields(log.Fields{"conn": fmt.Sprintf("%#v", proxyConn.tcpWriter), "err": err}).Info("failed reading from device, closing client")
		proxyConn.tcpWriter.Close()
		return err
	}
	bytesToSend := tcpbuf[:readBytes]
	remainingBytes := readBytes
	for remainingBytes > 0 {
		sentBytes, writerErr := proxyConn.tcpWriter.Write(bytesToSend)
		if writerErr != nil {
			log.WithFields(log.Fields{"conn": fmt.Sprintf("%#v", proxyConn.tcpWriter), "err": writerErr}).Info("failed writing to host tcp connection, closing client")
			proxyConn.tcpWriter.Close()
			return writerErr
		}
		remainingBytes -= sentBytes
		bytesToSend = bytesToSend[sentBytes:]
	}

	return nil
}
*/
