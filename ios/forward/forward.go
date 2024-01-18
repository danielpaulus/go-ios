package forward

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/danielpaulus/go-ios/ios"
	log "github.com/sirupsen/logrus"
)

type iosproxy struct {
	tcpConn    net.Conn
	deviceConn ios.DeviceConnectionInterface
}

type ConnListener struct {
	listener net.Listener
	quit     chan interface{}
}

// Forward forwards every connection made to the hostPort to whatever service runs inside an app on the device on phonePort.
func Forward(device ios.DeviceEntry, hostPort uint16, phonePort uint16) (*ConnListener, error) {
	log.Infof("Start listening on port %d forwarding to port %d on device", hostPort, phonePort)
	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", hostPort))
	if err != nil {
		return nil, fmt.Errorf("forward: failed listener with err: %w", err)
	}
	cl := &ConnListener{
		listener: l,
		quit:     make(chan interface{}),
	}

	go connectionAccept(cl, device.DeviceID, phonePort)

	return cl, nil
}

// Close stops listening on the host port for the forwarded connection
func (cl *ConnListener) Close() error {
	close(cl.quit)

	err := cl.listener.Close()
	if err != nil {
		return fmt.Errorf("forward: failed closing listener with err: %w", err)
	}

	return nil
}

func connectionAccept(cl *ConnListener, deviceID int, phonePort uint16) {
	for {
		select {
		case <-cl.quit:
			log.WithFields(log.Fields{"phonePort": phonePort}).Info("closed listener successfully")
			return
		default:
			clientConn, err := cl.listener.Accept()
			if err != nil {
				log.Errorf("Error accepting new connection %v", err)
				continue
			}
			log.WithFields(log.Fields{"conn": fmt.Sprintf("%#v", cl)}).Info("new client connected")
			go StartNewProxyConnection(context.TODO(), clientConn, deviceID, phonePort)
		}
	}
}

func StartNewProxyConnection(ctx context.Context, clientConn io.ReadWriteCloser, deviceID int, phonePort uint16) error {
	usbmuxConn, err := ios.NewUsbMuxConnectionSimple()
	if err != nil {
		log.Errorf("could not connect to usbmuxd: %+v", err)
		clientConn.Close()
		return fmt.Errorf("could not connect to usbmuxd: %v", err)
	}
	muxError := usbmuxConn.Connect(deviceID, phonePort)
	if muxError != nil {
		log.WithFields(log.Fields{"conn": fmt.Sprintf("%#v", clientConn), "err": muxError, "phonePort": phonePort}).Infof("could not connect to phone")
		clientConn.Close()
		return fmt.Errorf("could not connect to port:%d on iOS: %v", phonePort, err)
	}
	log.WithFields(log.Fields{"conn": fmt.Sprintf("%#v", clientConn), "phonePort": phonePort}).Infof("Connected to port")
	deviceConn := usbmuxConn.ReleaseDeviceConnection()

	// proxyConn := iosproxy{clientConn, deviceConn}
	ctx2, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup
	wg.Add(1)

	closed := false
	go func() {
		io.Copy(clientConn, deviceConn.Reader())
		if ctx2.Err() == nil {
			cancel()
			clientConn.Close()
			deviceConn.Close()
			closed = true
		}

		log.Errorf("forward: close clientConn <-- deviceConn")
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		io.Copy(deviceConn.Writer(), clientConn)
		if ctx2.Err() == nil {
			cancel()
			clientConn.Close()
			deviceConn.Close()
			closed = true
		}

		log.Errorf("forward: close clientConn --> deviceConn")
		wg.Done()
	}()

	<-ctx2.Done()
	if !closed {
		clientConn.Close()
		deviceConn.Close()
	}

	wg.Wait()
	return nil
}

func (proxyConn *iosproxy) Close() {
	proxyConn.tcpConn.Close()
}
