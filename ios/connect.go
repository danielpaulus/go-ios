package ios

import (
	"fmt"
	"net"
	"strings"

	"github.com/danielpaulus/go-ios/ios/xpc"
	"golang.org/x/net/http2"
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

// ConnectToService connects to a service on the phone and returns the ready to use DeviceConnectionInterface
func ConnectToService(device DeviceEntry, serviceName string) (DeviceConnectionInterface, error) {
	v, err := GetProductVersion(device)
	if err != nil {
		return nil, fmt.Errorf("failed to get product version. %w", err)
	}
	if v.Major() < 17 {
		return connectToServiceUsbmuxd(device, serviceName)
	}
	return connectToServiceTunnelIface(device, serviceName)
}

type FramerDataWriter struct {
	framer    http2.Framer
	streamID  uint32
	endStream bool
}

func (writer FramerDataWriter) Write(p []byte) (int, error) {
	err := writer.framer.WriteData(writer.streamID, writer.endStream, p)

	return len(p), err
}

func connectToServiceTunnelIface(device DeviceEntry, serviceName string) (DeviceConnectionInterface, error) {
	port := device.Rsd.GetPort(serviceName)
	conn, err := net.Dial("tcp6", fmt.Sprintf("[%s]:%d", device.Address, port))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to device. %w", err)
	}
	if !strings.Contains(serviceName, "testmanager") {
		err = RsdCheckin(conn)
	}
	if err != nil {
		return nil, err
	}

	deviceInterface := NewDeviceConnectionWithConn(conn)
	// TODO : everything after this line should go into its own method, i.e doHandshake()

	// TODO : send HTTP MAGIC

	framer := http2.NewFramer(deviceInterface.c, deviceInterface.c)

	// TODO : test and then figure out if we should keep reading the frame and somehow act on it
	firstReadFrame, err := framer.ReadFrame()
	if err != nil {
		return nil, err
	} else {
		print(firstReadFrame) // TODO : remove after debugging
	}

	err = framer.WriteSettings(
		http2.Setting{ID: http2.SettingMaxConcurrentStreams, Val: 100},  // TODO : Extract constant
		http2.Setting{ID: http2.SettingInitialWindowSize, Val: 1048576}, // TODO : Extract constant
	)
	if err != nil {
		return nil, err
	}

	err = framer.WriteWindowUpdate(0, 983041)
	if err != nil {
		return nil, err
	}

	err = framer.WriteHeaders(http2.HeadersFrameParam{StreamID: rootChannel, EndHeaders: true})
	if err != nil {
		return nil, err
	}

	err = xpc.EncodeData(FramerDataWriter{
		framer:    *framer,
		streamID:  rootChannel,
		endStream: false,
	}, map[string]interface{}{}, deviceInterface.rootChannelMessageId, false)
	if err != nil {
		return nil, err
	}

	// TODO : send Data frame

	deviceInterface.proceedToNextRootChannelMessage()

	// TODO : send remaining frames (figure out what)

	return deviceInterface, nil
}

func connectToServiceUsbmuxd(device DeviceEntry, serviceName string) (DeviceConnectionInterface, error) {
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
