package usbmux

import (
	"fmt"
)

type connectMessage struct {
	BundleID            string
	ClientVersionString string
	MessageType         string
	ProgName            string
	LibUSBMuxVersion    uint32 `plist:"kLibUSBMuxVersion"`
	DeviceID            uint32
	PortNumber          uint32
}

func newConnectMessage(deviceID int, portNumber int) *connectMessage {
	data := &connectMessage{
		BundleID:            "go.ios.control",
		ClientVersionString: "go-usbmux-0.0.1",
		MessageType:         "Connect",
		ProgName:            "go-usbmux",
		LibUSBMuxVersion:    3,
		DeviceID:            uint32(deviceID),
		PortNumber:          uint32(portNumber),
	}
	return data
}

//Connect issues a Connect Message to UsbMuxd for the given deviceID on the given port
//enabling the newCodec for it.
//It returns an error containing the UsbMux error code should the connect fail.
func (muxConn *MuxConnection) Connect(deviceID int, port uint16) error {
	msg := newConnectMessage(deviceID, int(Ntohs(port)))
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

//ConnectWithStartServiceResponse issues a Connect Message to UsbMuxd for the given deviceID on the given port
//enabling the newCodec for it. It also enables SSL on the new service connection if requested by StartServiceResponse.
//It returns an error containing the UsbMux error code should the connect fail.
func (muxConn *MuxConnection) ConnectWithStartServiceResponse(deviceID int, startServiceResponse StartServiceResponse, pairRecord PairRecord) error {
	err := muxConn.Connect(deviceID, startServiceResponse.Port)
	if err != nil {
		return err
	}

	if startServiceResponse.EnableServiceSSL {
		err = muxConn.deviceConn.EnableSessionSsl(pairRecord)
		if err != nil {
			return err
		}
	}

	return nil
}

//ConnectLockdown connects this Usbmux connection to the LockDown service that
// always runs on the device on the same port. The connect call needs the deviceID which can be
// retrieved from a DeviceList using the ListDevices function. After this function
// is done, the UsbMuxConnection cannot be used anymore because the same underlying
// network connection is used for talking to Lockdown. Sending usbmux commands would break it.
// It returns a new LockDownConnection.
func (muxConn *MuxConnection) ConnectLockdown(deviceID int) (*LockDownConnection, error) {
	msg := newConnectMessage(deviceID, Lockdownport)
	muxConn.Send(msg)
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
