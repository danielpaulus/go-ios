package debugproxy

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/danielpaulus/go-ios/usbmux"
	"github.com/danielpaulus/go-ios/usbmux/proxy_utils"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

const connectionJSONFileName = "connections.json"

//DebugProxy can be used to dump and modify communication between mac and host
type DebugProxy struct {
	mux               sync.Mutex
	serviceMap        map[string]PhoneServiceInformation
	connectionCounter int
	WorkingDir        string
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
	info       ConnectionInfo
	log        *logrus.Entry
	mux        sync.Mutex
	closed     bool
}

type ConnectionInfo struct {
	ConnectionPath string
	CreatedAt      time.Time
	ID             string
}

func (p *ProxyConnection) LogClosed() {
	p.mux.Lock()
	defer p.mux.Unlock()
	if p.closed {
		return
	}
	p.closed = true
	p.log.Info("Connection closed")
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

	d.setupDirectory()
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
		id := fmt.Sprintf("#%d", d.connectionCounter)
		connectionPath := filepath.Join(".", d.WorkingDir, "connection-"+id+"-"+time.Now().UTC().Format("2006.01.02-15.04.05.000"))

		os.MkdirAll(connectionPath, os.ModePerm)

		info := ConnectionInfo{ConnectionPath: connectionPath, CreatedAt: time.Now(), ID: id}
		d.addConnectionInfoToJsonFile(info)

		startProxyConnection(conn, originalSocket, pairRecord, d, info)

	}
}

func startProxyConnection(conn net.Conn, originalSocket string, pairRecord usbmux.PairRecord, debugProxy *DebugProxy, info ConnectionInfo) {
	connListeningOnUnixSocket := usbmux.NewUsbMuxConnectionWithConn(conn)
	connectionToDevice := usbmux.NewUsbMuxConnectionToSocket(originalSocket)

	p := ProxyConnection{info.ID, pairRecord, debugProxy, info, log.WithFields(log.Fields{"id": info.ID}), sync.Mutex{}, false}

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

func (d *DebugProxy) setupDirectory() {
	newpath := filepath.Join(".", "dump-"+time.Now().UTC().Format("2006.01.02-15.04.05.000"))
	d.WorkingDir = newpath
	os.MkdirAll(newpath, os.ModePerm)
}

func (d DebugProxy) addConnectionInfoToJsonFile(connInfo ConnectionInfo) {
	file, err := os.OpenFile(filepath.Join(d.WorkingDir, connectionJSONFileName),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	data, err := json.Marshal(connInfo)
	if err != nil {
		log.Printf("Failed json:%s", err)
	}
	file.Write(data)
	io.WriteString(file, "\n")
	file.Close()
}

func (p ProxyConnection) logJSONMessageFromDevice(msg interface{}) {
	const outPath = "jsondump-fromdevice.bin"
	writeJSON(filepath.Join(p.info.ConnectionPath, outPath), msg)
}
func (p ProxyConnection) logJSONMessageToDevice(msg interface{}) {
	const outPath = "jsondump-todevice.bin"
	writeJSON(filepath.Join(p.info.ConnectionPath, outPath), msg)
}

func writeJSON(filePath string, JSON interface{}) {
	file, err := os.OpenFile(filePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Could not write to file, this should not happen", err, filePath)
	}
	jsonmsg, err := json.Marshal(JSON)
	if err != nil {
		log.Warnf("Error encoding '%s' to json: %s", JSON, err)
	}
	file.Write(jsonmsg)
	io.WriteString(file, "\n")
	file.Close()
}