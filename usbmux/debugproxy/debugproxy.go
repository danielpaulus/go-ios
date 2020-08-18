package debugproxy

import (
	"fmt"
	"net"
	"sync"

	"github.com/danielpaulus/go-ios/usbmux"
	"github.com/danielpaulus/go-ios/usbmux/proxy_utils"
	log "github.com/sirupsen/logrus"
)

//DebugProxy can be used to dump and modify communication between mac and host
type DebugProxy struct {
	mux               sync.Mutex
	serviceMap        map[string]PhoneServiceInformation
	connectionCounter int
}

//PhoneServiceInformation contains info about a service started on the phone via lockdown.
type PhoneServiceInformation struct {
	ServicePort uint16
	ServiceName string
	UseSSL      bool
}

//ProxyConnection keeps track of the pairRecord and uses an ID to identify connections.
type ProxyConnection struct {
	id         string
	pairRecord usbmux.PairRecord
	debugProxy *DebugProxy
}

func (d *DebugProxy) storeServiceInformation(serviceInfo PhoneServiceInformation) {
	d.mux.Lock()
	defer d.mux.Unlock()
	d.serviceMap[serviceInfo.ServiceName] = serviceInfo
}

func (d *DebugProxy) retrieveServiceInfoByName(serviceName string) PhoneServiceInformation {
	d.mux.Lock()
	defer d.mux.Unlock()
	return d.serviceMap[serviceName]
}

func (d *DebugProxy) retrieveServiceInfoByPort(port uint16) (PhoneServiceInformation, error) {
	d.mux.Lock()
	defer d.mux.Unlock()
	for _, element := range d.serviceMap {
		if element.ServicePort == port {
			return element, nil
		}
	}
	return PhoneServiceInformation{}, fmt.Errorf("No Service found for port %d", port)
}

//NewDebugProxy creates a new Default proxy
func NewDebugProxy() *DebugProxy {
	return &DebugProxy{mux: sync.Mutex{}, serviceMap: make(map[string]PhoneServiceInformation)}
}

//Launch moves the original /var/run/usbmuxd to /var/run/usbmuxd.real and starts the server at /var/run/usbmuxd
func (d *DebugProxy) Launch() error {
	pairRecord := usbmux.ReadPairRecord("b89227a71e1a97c00bcc297d33c3f58b789dbc8a")
	originalSocket, err := proxy_utils.MoveSock(usbmux.DefaultUsbmuxdSocket)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "socket": usbmux.DefaultUsbmuxdSocket}).Error("Unable to move, lacking permissions?")
		return err
	}

	listener, err := net.Listen("unix", usbmux.DefaultUsbmuxdSocket)
	if err != nil {
		log.Fatal("Could not listen on usbmuxd socket, do I have access permissions?", err)
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Errorf("error with connection: %e", err)
		}
		d.connectionCounter++
		startProxyConnection(conn, originalSocket, pairRecord, d)

	}
}

func startProxyConnection(conn net.Conn, originalSocket string, pairRecord usbmux.PairRecord, debugProxy *DebugProxy) {
	connListeningOnUnixSocket := usbmux.NewUsbMuxConnectionWithConn(conn)
	connectionToDevice := usbmux.NewUsbMuxConnectionToSocket(originalSocket)
	p := ProxyConnection{fmt.Sprintf("#%d", debugProxy.connectionCounter), pairRecord, debugProxy}

	go proxyUsbMuxConnection(&p, connListeningOnUnixSocket, connectionToDevice)
}

//Close moves /var/run/usbmuxd.real back to /var/run/usbmuxd and disconnects all active proxy connections
func (d *DebugProxy) Close() {
	log.Info("Moving back original socket")
	err := proxy_utils.MoveBack(usbmux.DefaultUsbmuxdSocket)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed moving back socket")
	}
}
