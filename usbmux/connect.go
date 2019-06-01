package usbmux

import (
	"encoding/binary"
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

//ntohs is a re-implementation of the C function ntohs.
//it means networkorder to host oder and basically swaps
//the endianness of the given int.
//It returns port converted to little endian.
func ntohs(port uint16) uint16 {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, port)
	return binary.LittleEndian.Uint16(buf)
}

//Connect issues a Connect Message to UsbMuxd for the given deviceID on the given port
//enabling the newCodec for it.
//It returns an error containing the UsbMux error code should the connect fail.
func (muxConn *MuxConnection) Connect(deviceID int, port uint16, newCodec Codec) error {
	msg := newConnectMessage(deviceID, int(ntohs(port)))
	responseBytes := muxConn.deviceConn.SendForProtocolUpgrade(muxConn, msg, newCodec)
	response := MuxResponsefromBytes(responseBytes)
	if response.IsSuccessFull() {
		return nil
	}
	return fmt.Errorf("Failed connecting to service, error code:%d", response.Number)
}

//ConnectLockdown connects this Usbmux connection to the LockDown service that
// always runs on the device on the same port. The connect call needs the deviceID which can be
// retrieved from a DeviceList using the ListDevices function. After this function
// is done, the UsbMuxConnection cannot be used anymore because the same underlying
// network connection is used for talking to Lockdown. Sending usbmux commands would break it.
// It returns a new LockDownConnection.
func (muxConn *MuxConnection) ConnectLockdown(deviceID int) (*LockDownConnection, error) {
	msg := newConnectMessage(deviceID, lockdownport)
	responseChannel := make(chan []byte)
	lockdownConn := &LockDownConnection{muxConn.deviceConn, "", responseChannel, NewPlistCodec(muxConn.deviceConn, responseChannel)}

	responseBytes := muxConn.deviceConn.SendForProtocolUpgrade(muxConn, msg, lockdownConn.plistCodec)
	response := MuxResponsefromBytes(responseBytes)
	if response.IsSuccessFull() {
		return lockdownConn, nil
	}
	return nil, fmt.Errorf("Failed connecting to Lockdown with error code:%d", response.Number)
}
