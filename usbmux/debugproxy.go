package usbmux

/*
import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/danielpaulus/go-ios/usbmux/proxy_utils"
	log "github.com/sirupsen/logrus"
	"howett.net/plist"
)

//DebugProxy can be used to dump and modify communication between mac and host
type DebugProxy struct {
	mux        sync.Mutex
	serviceMap map[string]PhoneServiceInformation
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
	pairRecord := ReadPairRecord("b89227a71e1a97c00bcc297d33c3f58b789dbc8a")
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

		startProxyConnection(conn, originalSocket, pairRecord, d)

	}

}

type PhoneServiceInformation struct {
	ServicePort uint16
	ServiceName string
	UseSSL      bool
}

type ProxyConnection struct {
	connListeningOnUnixSocket DeviceConnectionInterface
	unixSocketChannel         chan []byte
	connectionToDevice        DeviceConnectionInterface
	deviceChannel             chan []byte
	pairRecord                PairRecord
	WaitingForProtocolChange  bool
	debugProxy                *DebugProxy
}

type BinDumpCodec struct {
	received chan []byte
}

func NewBinDumpCodec(channel chan []byte) *BinDumpCodec {
	return &BinDumpCodec{channel}
}

func (b BinDumpCodec) Encode(msg interface{}) ([]byte, error) {
	return msg.([]byte), nil
}

func (b *BinDumpCodec) Decode(r io.Reader) error {
	buffer := make([]byte, 1024)
	n, err := r.Read(buffer)
	if err != nil {
		return err
	}
	b.received <- buffer[0:n]
	return nil
}

func (p *ProxyConnection) handleConnect(connectMessage interface{}, u *MuxConnection, serviceInfo PhoneServiceInformation) {
	p.WaitingForProtocolChange = true
	p.deviceChannel <- nil

	newDeviceChannel := make(chan []byte)
	newUnixSocketChannel := make(chan []byte)
	p.connectionToDevice.StopReadingAfterNextMessage()
	p.connectionToDevice.Send(connectMessage)
	response := <-p.deviceChannel
	p.WaitingForProtocolChange = false
	var decoded map[string]interface{}
	decoder := plist.NewDecoder(bytes.NewReader(response))
	decoder.Decode(&decoded)

	p.connectionToDevice.ResumeReadingWithNewCodec(NewBinDumpCodec(newDeviceChannel))

	p.connListeningOnUnixSocket.Send(decoded)
	unixSocketCodec := NewBinDumpCodec(newUnixSocketChannel)
	p.connListeningOnUnixSocket.SetCodec(unixSocketCodec)
	p.unixSocketChannel = newUnixSocketChannel
	p.deviceChannel = newDeviceChannel
	if u != nil {
		u.StopDecoding()
	}
	log.Info("Added BinDump Codec")
	go readOnUnixDomainSocketAndForwardToDeviceGeneric(p)
	go readOnDeviceConnectionAndForwardToUnixDomainConnectionGeneric(p)
}

func (p *ProxyConnection) handleConnectToLockdown(connectMessage interface{}, u *MuxConnection) {
	p.WaitingForProtocolChange = true
	p.deviceChannel <- nil

	newDeviceChannel := make(chan []byte)
	newUnixSocketChannel := make(chan []byte)
	p.connectionToDevice.StopReadingAfterNextMessage()
	p.connectionToDevice.Send(connectMessage)
	response := <-p.deviceChannel
	log.Info(string(response))
	p.WaitingForProtocolChange = false
	var decoded map[string]interface{}
	decoder := plist.NewDecoder(bytes.NewReader(response))
	decoder.Decode(&decoded)
	plistCodec := NewPlistCodec(newDeviceChannel)
	p.connectionToDevice.ResumeReadingWithNewCodec(plistCodec)

	p.connListeningOnUnixSocket.Send(decoded)
	singleDecodePlistCodec := NewPlistCodecSingleDecode(newUnixSocketChannel)
	p.connListeningOnUnixSocket.SetCodec(singleDecodePlistCodec)
	p.unixSocketChannel = newUnixSocketChannel
	p.deviceChannel = newDeviceChannel
	if u != nil {
		u.StopDecoding()
	}
	log.Info("Upgrade to lockdown complete")
	go readOnUnixDomainSocketAndForwardToDeviceLockdownSingleDecode(p, singleDecodePlistCodec)
	go readOnDeviceConnectionAndForwardToUnixDomainConnection(p)
}

func (p *ProxyConnection) handleSSLUpgrade(startSessionMessage interface{}, plistCodec *PlistCodec) {
	p.WaitingForProtocolChange = true
	p.deviceChannel <- nil
	p.connectionToDevice.StopReadingAfterNextMessage()
	p.connectionToDevice.Send(startSessionMessage)
	response := <-p.deviceChannel
	log.Info(string(response))
	p.WaitingForProtocolChange = false
	var decoded map[string]interface{}
	decoder := plist.NewDecoder(bytes.NewReader(response))
	decoder.Decode(&decoded)

	log.Info(decoded)

	if decoded["EnableSessionSSL"] == true {
		log.Info("should enable ssl")
		p.connectionToDevice.EnableSessionSsl(p.pairRecord)
		p.connListeningOnUnixSocket.StopReadingAfterNextMessage()
		plistCodec.StopDecoding()
		p.connListeningOnUnixSocket.Send(decoded)
		p.connListeningOnUnixSocket.EnableSessionSslServerMode(p.pairRecord)
		p.connectionToDevice.ResumeReading()
		p.connListeningOnUnixSocket.ResumeReading()
		go readOnDeviceConnectionAndForwardToUnixDomainConnection(p)
		go readOnUnixDomainSocketAndForwardToDeviceLockdownSingleDecode(p, plistCodec)
	} else {
		log.Fatal("lockdown without ssl should not exist")
	}
}

func startProxyConnection(conn net.Conn, originalSocket string, pairRecord PairRecord, debugProxy *DebugProxy) {
	connListeningOnUnixSocket := NewUsbMuxServerConnection(conn)
	connectionToDevice := NewUsbMuxConnectionToSocket(originalSocket)
	p := ProxyConnection{connListeningOnUnixSocket.deviceConn, connListeningOnUnixSocket.ResponseChannel, connectionToDevice.deviceConn, connectionToDevice.ResponseChannel, pairRecord, false, debugProxy}

	go readOnUnixDomainSocketAndForwardToDeviceUsbMuxSingleDecode(&p, connListeningOnUnixSocket)
	go readOnDeviceConnectionAndForwardToUnixDomainConnection(&p)
}

func readOnUnixDomainSocketAndForwardToDeviceLockdownSingleDecode(p *ProxyConnection, plistCodec *PlistCodec) {
	for {
		plistCodec.StartDecode()
		msg := <-p.unixSocketChannel

		if msg == nil {
			log.Info("service on host disconnected")
			p.connectionToDevice.Close()
			return
		}
		var decoded map[string]interface{}
		decoder := plist.NewDecoder(bytes.NewReader(msg))
		err := decoder.Decode(&decoded)
		if err != nil {
			log.Info(err)
		}

		log.Info(decoded)
		if decoded["Request"] == "StartSession" {
			log.Info("Lockdown StartSession detected, kicking of SSL check")

			p.handleSSLUpgrade(decoded, plistCodec)
			return
		}
		p.connectionToDevice.Send(decoded)

	}
}

func readOnUnixDomainSocketAndForwardToDeviceUsbMuxSingleDecode(p *ProxyConnection, u *MuxConnection) {
	for {
		u.StartDecode()
		msg := <-p.unixSocketChannel

		if msg == nil {
			log.Info("service on host disconnected")
			p.connectionToDevice.Close()
			return
		}
		var decoded map[string]interface{}
		decoder := plist.NewDecoder(bytes.NewReader(msg))
		err := decoder.Decode(&decoded)
		if err != nil {
			log.Info(err)
		}

		log.Info(decoded)
		if decoded["MessageType"] == "Connect" {
			log.Info("Detected Connect Message")
			port := decoded["PortNumber"].(uint64)
			if int(port) == lockdownport {
				log.Info("Upgrading to Lockdown")
				p.handleConnectToLockdown(decoded, u)
			} else {
				info, err := p.debugProxy.retrieveServiceInfoByPort(ntohs(uint16(port)))
				if err != nil {
					log.Fatal("ServiceInfo for port not found, this is a bug :-)")
				}
				log.Info("Connection to service detected", info)
				p.handleConnect(decoded, u, info)
			}
			return
		}
		p.connectionToDevice.Send(decoded)

	}
}

func readOnDeviceConnectionAndForwardToUnixDomainConnectionGeneric(p *ProxyConnection) {
	for {
		msg := <-p.deviceChannel

		if msg == nil {
			log.Info("device disconnected")
			p.connListeningOnUnixSocket.Close()
			return
		}

		//log.Info(hex.Dump(msg))
		p.connListeningOnUnixSocket.Send(msg)
	}
}

func readOnUnixDomainSocketAndForwardToDeviceGeneric(p *ProxyConnection) {

	for {
		msg := <-p.unixSocketChannel

		if msg == nil {
			log.Info("service on host disconnected")
			p.connectionToDevice.Close()
			return
		}
		//log.Info(hex.Dump(msg))
		p.connectionToDevice.Send(msg)

	}

}

func readOnDeviceConnectionAndForwardToUnixDomainConnection(p *ProxyConnection) {
	for {
		msg := <-p.deviceChannel
		if p.WaitingForProtocolChange {
			log.Info("Protocol Change, killing proxy read loop")
			return
		}

		if msg == nil {
			log.Info("device disconnected")
			p.connListeningOnUnixSocket.Close()
			return
		}

		var decoded map[string]interface{}
		decoder := plist.NewDecoder(bytes.NewReader(msg))
		err := decoder.Decode(&decoded)
		if err != nil {
			log.Info(err)
		}

		log.Info(decoded)
		if decoded["Request"] == "StartService" {

			info := PhoneServiceInformation{ServicePort: uint16(decoded["Port"].(uint64)), ServiceName: decoded["Service"].(string), UseSSL: decoded["EnableServiceSSL"].(bool)}

			log.Info("Detected Service Start", (info))
			p.debugProxy.storeServiceInformation(info)

		}
		p.connListeningOnUnixSocket.Send(decoded)
	}
}

//hier muss der single step lockdown read rein und nach startSession den enableSSL call machen
func readOnUnixDomainSocketAndForwardToDevice(p *ProxyConnection) {

	for {
		msg := <-p.unixSocketChannel

		if msg == nil {
			log.Info("service on host disconnected")
			p.connectionToDevice.Close()
			return
		}
		var decoded map[string]interface{}
		decoder := plist.NewDecoder(bytes.NewReader(msg))
		err := decoder.Decode(&decoded)
		if err != nil {
			log.Info(err)
		}

		if decoded["MessageType"] == "Connect" {
			log.Info("Upgrading to Lockdown")
			p.handleConnectToLockdown(decoded, nil)
			return
		}

		p.connectionToDevice.Send(decoded)
		log.Info("RCV ONHOST AND SEND TO DEVICE:")
		log.Info(decoded)
		log.Info("END SEND")
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
*/
