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

	ios "github.com/danielpaulus/go-ios/ios"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

const connectionJSONFileName = "connections.json"

//DebugProxy can be used to dump and modify communication between mac and host
type DebugProxy struct {
	mux               sync.Mutex
	serviceList       []PhoneServiceInformation
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
	pairRecord ios.PairRecord
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
	p.log.Trace("Connection closed")
}

func (d *DebugProxy) storeServiceInformation(serviceInfo PhoneServiceInformation) {
	d.mux.Lock()
	defer d.mux.Unlock()
	d.serviceList = append(d.serviceList, serviceInfo)
}

func (d *DebugProxy) retrieveServiceInfoByPort(port uint16) (PhoneServiceInformation, error) {
	d.mux.Lock()
	defer d.mux.Unlock()
	for _, element := range d.serviceList {
		if element.ServicePort == port {
			return element, nil
		}
	}
	return PhoneServiceInformation{}, fmt.Errorf("No Service found for port %d", port)
}

//NewDebugProxy creates a new Default proxy
func NewDebugProxy() *DebugProxy {
	return &DebugProxy{mux: sync.Mutex{}, serviceList: []PhoneServiceInformation{}}
}

//Launch moves the original /var/run/usbmuxd to /var/run/usbmuxd.real and starts the server at /var/run/usbmuxd
func (d *DebugProxy) Launch(device ios.DeviceEntry, binaryMode bool) error {
	if binaryMode {
		log.Info("Lauching proxy in full binary mode")
	}
	var pairRecord ios.PairRecord
	if !binaryMode {
		var err error
		pairRecord, err = ios.ReadPairRecord(device.Properties.SerialNumber)
		if err != nil {
			return err
		}
		log.Infof("Successfully retrieved pairrecord: %s for device %s", pairRecord.HostID, device.Properties.SerialNumber)
	}
	originalSocket, err := MoveSock(ios.DefaultUsbmuxdSocket)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "socket": ios.DefaultUsbmuxdSocket}).Error("Unable to move, lacking permissions?")
		return err
	}
	d.setupDirectory()
	listener, err := net.Listen("unix", ios.DefaultUsbmuxdSocket)
	if err != nil {
		log.Error("Could not listen on usbmuxd socket, do I have access permissions?", err)
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

		bindumpHostProxyFile := filepath.Join(connectionPath, "bindump-hostservice-to-proxy.txt")

		if !binaryMode {
			//if the proxy is in full binary mode, there is no point in creating another binary dump
			log.Infof("Creating binary dump of all communication between MAC OS and debugproxy at: %s", bindumpHostProxyFile)
			conn = NewDumpingConn(bindumpHostProxyFile, conn)
		}

		startProxyConnection(conn, originalSocket, pairRecord, d, info, binaryMode)
	}
}

func startProxyConnection(conn net.Conn, originalSocket string, pairRecord ios.PairRecord, debugProxy *DebugProxy, info ConnectionInfo, binaryMode bool) {

	devConn, err := ios.NewDeviceConnection(originalSocket)
	if err != nil {
		log.Error(err)
		return
	}

	logger := log.WithFields(log.Fields{"id": info.ID})
	p := ProxyConnection{info.ID, pairRecord, debugProxy, info, logger, sync.Mutex{}, false}

	if binaryMode {
		binOnUnixSocket := BinaryForwardingProxy{ios.NewDeviceConnectionWithConn(conn), NewBinDumpOnly("does not matter", filepath.Join(info.ConnectionPath, "rawbindump-from-host-service.bin"), logger)}
		binToDevice := BinaryForwardingProxy{devConn, NewBinDumpOnly("does not matter", filepath.Join(info.ConnectionPath, "rawbindump-from-device.bin"), logger)}
		go proxyBinDumpConnection(&p, binOnUnixSocket, binToDevice)
		return
	}
	connListeningOnUnixSocket := ios.NewUsbMuxConnection(ios.NewDeviceConnectionWithConn(conn))
	connectionToDevice := ios.NewUsbMuxConnection(devConn)
	go proxyUsbMuxConnection(&p, connListeningOnUnixSocket, connectionToDevice)

}

//Close moves /var/run/usbmuxd.real back to /var/run/usbmuxd and disconnects all active proxy connections
func (d *DebugProxy) Close() {
	log.Info("Moving back original socket")
	err := MoveBack(ios.DefaultUsbmuxdSocket)
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

func (p ProxyConnection) logJSONMessageFromDevice(msg map[string]interface{}) {
	const outPath = "jsondump.json"
	msg["direction"] = "device->host"
	writeJSON(filepath.Join(p.info.ConnectionPath, outPath), msg)
}
func (p ProxyConnection) logJSONMessageToDevice(msg map[string]interface{}) {
	const outPath = "jsondump.json"
	msg["direction"] = "host->device"
	writeJSON(filepath.Join(p.info.ConnectionPath, outPath), msg)
}

func writeJSON(filePath string, JSON interface{}) {
	file, err := os.OpenFile(filePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Sprintf("Could not write to file err: %v filepath:'%s'", err, filePath))
	}
	jsonmsg, err := json.Marshal(JSON)
	if err != nil {
		log.Warnf("Error encoding '%s' to json: %s", JSON, err)
	}
	file.Write(jsonmsg)
	io.WriteString(file, "\n")
	file.Close()
}
