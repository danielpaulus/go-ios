package ios

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/danielpaulus/go-ios/ios/http"

	"github.com/danielpaulus/go-ios/ios/xpc"
)

type connectMessage struct {
	BundleID            string
	ClientVersionString string
	MessageType         string
	ProgName            string
	LibUSBMuxVersion    uint32 `plist:"kLibUSBMuxVersion"`
	DeviceID            uint32
	PortNumber          uint16
}

func newConnectMessage(deviceID int, portNumber uint16) connectMessage {
	data := connectMessage{
		BundleID:            "go.ios.control",
		ClientVersionString: "go-usbmux-0.0.1",
		MessageType:         "Connect",
		ProgName:            "go-usbmux",
		LibUSBMuxVersion:    3,
		DeviceID:            uint32(deviceID),
		PortNumber:          portNumber,
	}
	return data
}

// Connect issues a Connect Message to UsbMuxd for the given deviceID on the given port
// enabling the newCodec for it.
// It returns an error containing the UsbMux error code should the connect fail.
func (muxConn *UsbMuxConnection) Connect(deviceID int, port uint16) error {
	msg := newConnectMessage(deviceID, Ntohs(port))
	muxConn.Send(msg)
	resp, err := muxConn.ReadMessage()
	if err != nil {
		return err
	}
	response := MuxResponsefromBytes(resp.Payload)
	if response.IsSuccessFull() {
		return nil
	}
	return fmt.Errorf("Failed connecting to service, error code:%d", response.Number)
}

// serviceConfigurations stores info about which DTX based services only execute a SSL Handshake
// and then go back to sending unencrypted data right after the handshake.
var serviceConfigurations = map[string]bool{
	"com.apple.instruments.remoteserver":                 true,
	"com.apple.accessibility.axAuditDaemon.remoteserver": true,
	"com.apple.testmanagerd.lockdown":                    true,
	"com.apple.debugserver":                              true,
}

// ConnectLockdown connects this Usbmux connection to the LockDown service that
// always runs on the device on the same port. The connect call needs the deviceID which can be
// retrieved from a DeviceList using the ListDevices function. After this function
// is done, the UsbMuxConnection cannot be used anymore because the same underlying
// network connection is used for talking to Lockdown. Sending usbmux commands would break it.
// It returns a new LockDownConnection.
func (muxConn *UsbMuxConnection) ConnectLockdown(deviceID int) (*LockDownConnection, error) {
	msg := newConnectMessage(deviceID, Lockdownport)
	err := muxConn.Send(msg)
	if err != nil {
		return &LockDownConnection{}, err
	}
	resp, err := muxConn.ReadMessage()
	if err != nil {
		return &LockDownConnection{}, err
	}
	response := MuxResponsefromBytes(resp.Payload)
	if response.IsSuccessFull() {
		return &LockDownConnection{muxConn.deviceConn, "", NewPlistCodec()}, nil
	}

	return nil, fmt.Errorf("Failed connecting to Lockdown with error code:%d", response.Number)
}

func ConnectToService(device DeviceEntry, serviceName string) (DeviceConnectionInterface, error) {
	startServiceResponse, err := StartService(device, serviceName)
	if err != nil {
		return nil, err
	}
	pairRecord, err := ReadPairRecord(device.Properties.SerialNumber)
	if err != nil {
		return nil, err
	}

	muxConn, err := NewUsbMuxConnectionSimple()
	if err != nil {
		return nil, fmt.Errorf("Could not connect to usbmuxd socket, is it running? %w", err)
	}
	err = muxConn.connectWithStartServiceResponse(device.DeviceID, startServiceResponse, pairRecord)
	if err != nil {
		return nil, err
	}
	return muxConn.ReleaseDeviceConnection(), nil
}

// ConnectToShimService opens a new connection of the tunnel interface of the provided device
// to the provided service.
// The 'RSDCheckin' required by shim services is also executed before returning the connection to the caller
func ConnectToShimService(device DeviceEntry, service string) (DeviceConnectionInterface, error) {
	if !device.SupportsRsd() {
		return nil, fmt.Errorf("ConnectToShimService: Cannot connect to %s, missing tunnel address and RSD port.  To start the tunnel, run `ios tunnel start`", service)
	}
	port := device.Rsd.GetPort(service)
	conn, err := ConnectTUNDevice(device.Address, port, device)
	if err != nil {
		return nil, err
	}
	err = RsdCheckin(conn)
	if err != nil {
		return nil, err
	}
	return NewDeviceConnectionWithRWC(conn), nil
}

// ConnectToServiceTunnelIface connects to a service on an iOS17+ device using a XPC over HTTP2 connection
// It returns a new xpc.Connection
func ConnectToXpcServiceTunnelIface(device DeviceEntry, serviceName string) (*xpc.Connection, error) {
	if !device.SupportsRsd() {
		return nil, fmt.Errorf("ConnectToXpcServiceTunnelIface: Cannot connect to %s, missing tunnel address and RSD port. To start the tunnel, run `ios tunnel start`", serviceName)
	}
	port := device.Rsd.GetPort(serviceName)

	conn, err := ConnectTUNDevice(device.Address, port, device)
	if err != nil {
		return nil, fmt.Errorf("ConnectToHttp2: failed to dial: %w", err)
	}

	h, err := http.NewHttpConnection(conn)
	if err != nil {
		return nil, fmt.Errorf("ConnectToXpcServiceTunnelIface: failed to connect to http2: %w", err)
	}
	return CreateXpcConnection(h)
}

func ConnectToServiceTunnelIface(device DeviceEntry, serviceName string) (DeviceConnectionInterface, error) {
	if !device.SupportsRsd() {
		return nil, fmt.Errorf("ConnectToServiceTunnelIface: Cannot connect to %s, missing tunnel address and RSD port", serviceName)
	}
	port := device.Rsd.GetPort(serviceName)

	conn, err := ConnectTUNDevice(device.Address, port, device)
	if err != nil {
		return nil, fmt.Errorf("ConnectToServiceTunnelIface: failed to connect to tunnel: %w", err)
	}

	return NewDeviceConnectionWithRWC(conn), nil
}

func CreateXpcConnection(h *http.HttpConnection) (*xpc.Connection, error) {
	err := initializeXpcConnection(h)
	if err != nil {
		return nil, fmt.Errorf("CreateXpcConnection: failed to initialize xpc connection: %w", err)
	}

	clientServerChannel := http.NewStreamReadWriter(h, http.ClientServer)
	serverClientChannel := http.NewStreamReadWriter(h, http.ServerClient)

	xpcConn, err := xpc.New(clientServerChannel, serverClientChannel, h)
	if err != nil {
		return nil, fmt.Errorf("CreateXpcConnection: failed to create xpc connection: %w", err)
	}

	return xpcConn, nil
}

// connectWithStartServiceResponse issues a Connect Message to UsbMuxd for the given deviceID on the given port
// enabling the newCodec for it. It also enables SSL on the new service connection if requested by StartServiceResponse.
// It returns an error containing the UsbMux error code should the connect fail.
func (muxConn *UsbMuxConnection) connectWithStartServiceResponse(deviceID int, startServiceResponse StartServiceResponse, pairRecord PairRecord) error {
	err := muxConn.Connect(deviceID, startServiceResponse.Port)
	if err != nil {
		return err
	}

	var sslerr error
	if startServiceResponse.EnableServiceSSL {
		if _, ok := serviceConfigurations[startServiceResponse.Service]; ok {
			sslerr = muxConn.deviceConn.EnableSessionSslHandshakeOnly(pairRecord)
		} else {
			sslerr = muxConn.deviceConn.EnableSessionSsl(pairRecord)
		}
		if sslerr != nil {
			return sslerr
		}
	}

	return nil
}

func ConnectLockdownWithSession(device DeviceEntry) (*LockDownConnection, error) {
	muxConnection, err := NewUsbMuxConnectionSimple()
	if err != nil {
		return nil, fmt.Errorf("USBMuxConnection failed with: %v", err)
	}
	defer muxConnection.ReleaseDeviceConnection()

	pairRecord, err := muxConnection.ReadPair(device.Properties.SerialNumber)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve PairRecord with error: %v", err)
	}

	lockdownConnection, err := muxConnection.ConnectLockdown(device.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("Lockdown connection failed with: %v", err)
	}
	resp, err := lockdownConnection.StartSession(pairRecord)
	if err != nil {
		return nil, fmt.Errorf("StartSession failed: %+v error: %v", resp, err)
	}
	return lockdownConnection, nil
}

func initializeXpcConnection(h *http.HttpConnection) error {
	csWriter := http.NewStreamReadWriter(h, http.ClientServer)
	ssWriter := http.NewStreamReadWriter(h, http.ServerClient)

	err := xpc.EncodeMessage(csWriter, xpc.Message{
		Flags: xpc.AlwaysSetFlag,
		Body:  map[string]interface{}{},
		Id:    0,
	})
	if err != nil {
		return fmt.Errorf("initializeXpcConnection: failed to encode message: %w", err)
	}

	_, err = xpc.DecodeMessage(csWriter) // TODO : figure out if need to act on this frame
	if err != nil {
		return fmt.Errorf("initializeXpcConnection: failed to decode message: %w", err)
	}

	err = xpc.EncodeMessage(ssWriter, xpc.Message{
		Flags: xpc.InitHandshakeFlag | xpc.AlwaysSetFlag,
		Body:  nil,
		Id:    0,
	})
	if err != nil {
		return fmt.Errorf("initializeXpcConnection: failed to encode message 2: %w", err)
	}

	_, err = xpc.DecodeMessage(ssWriter) // TODO : figure out if need to act on this frame
	if err != nil {
		return fmt.Errorf("initializeXpcConnection: failed to decode message 2: %w", err)
	}

	err = xpc.EncodeMessage(csWriter, xpc.Message{
		Flags: 0x201, // alwaysSetFlag | 0x200
		Body:  nil,
		Id:    0,
	})
	if err != nil {
		return fmt.Errorf("initializeXpcConnection: failed to encode message 3: %w", err)
	}

	_, err = xpc.DecodeMessage(csWriter) // TODO : figure out if need to act on this frame
	if err != nil {
		return fmt.Errorf("initializeXpcConnection: failed to decode message 3: %w", err)
	}

	return nil
}

// ConnectTUNDevice creates a *net.TCPConn to the device at the given address and port.
// If the device is a userspaceTUN device provided by go-ios agent, it will connect to this
// automatically. Otherwise it will try a operating system level TUN device.
func ConnectTUNDevice(remoteIp string, port int, d DeviceEntry) (*net.TCPConn, error) {
	if !d.UserspaceTUN {
		return connectTUN(remoteIp, port)
	}

	addr, _ := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", d.UserspaceTUNHost, d.UserspaceTUNPort))
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("ConnectUserSpaceTunnel: failed to dial: %w", err)
	}
	err = conn.SetKeepAlive(true)
	if err != nil {
		return nil, fmt.Errorf("ConnectUserSpaceTunnel: failed to set keepalive: %w", err)
	}
	err = conn.SetKeepAlivePeriod(1 * time.Second)
	if err != nil {
		return nil, fmt.Errorf("ConnectUserSpaceTunnel: failed to set keepalive period: %w", err)
	}
	_, err = conn.Write(net.ParseIP(remoteIp).To16())
	portBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(portBytes, uint32(port))
	_, err1 := conn.Write(portBytes)
	return conn, errors.Join(err, err1)
}

// connect to a operating system level TUN device
func connectTUN(address string, port int) (*net.TCPConn, error) {
	addr, err := net.ResolveTCPAddr("tcp6", fmt.Sprintf("[%s]:%d", address, port))
	if err != nil {
		return nil, fmt.Errorf("ConnectToHttp2WithAddr: failed to resolve address: %w", err)
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("ConnectToHttp2WithAddr: failed to dial: %w", err)
	}
	err = conn.SetKeepAlive(true)
	if err != nil {
		return nil, fmt.Errorf("ConnectUserSpaceTunnel: failed to set keepalive: %w", err)
	}
	err = conn.SetKeepAlivePeriod(1 * time.Second)
	if err != nil {
		return nil, fmt.Errorf("ConnectUserSpaceTunnel: failed to set keepalive period: %w", err)
	}

	return conn, nil
}

// defaultHttpApiPort is the port on which we start the HTTP-Server for exposing started tunnels
// 60-105 is leetspeek for go-ios :-D
const defaultHttpApiPort = 60105

// defaultHttpApiHost is the host on which the HTTP-Server runs, by default it is 127.0.0.1
const defaultHttpApiHost = "127.0.0.1"

// DefaultHttpApiPort is the port on which we start the HTTP-Server for exposing started tunnels
// if GO_IOS_AGENT_PORT is set, we use that port. Otherwise we use the default port 60106.
// 60-105 is leetspeek for go-ios :-D
func HttpApiPort() int {
	port, err := strconv.Atoi(os.Getenv("GO_IOS_AGENT_PORT"))
	if err != nil {
		return defaultHttpApiPort
	}
	return port
}

// DefaultHttpApiHost is the host on which the HTTP-Server runs, by default it is 127.0.0.1
// if GO_IOS_AGENT_HOST is set, we use that host. Otherwise we use the default host
func HttpApiHost() string {
	host := os.Getenv("GO_IOS_AGENT_HOST")
	if host == "" {
		return defaultHttpApiHost
	}
	return host
}
