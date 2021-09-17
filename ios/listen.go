package ios

import (
	"bytes"
	"fmt"

	"howett.net/plist"
)

//ListenType contains infos for creating a LISTEN message for USBMUX
type ListenType struct {
	MessageType         string
	ProgName            string
	ClientVersionString string
	ConnType            int
	kLibUSBMuxVersion   int
}

//AttachedMessage contains some info about when iOS devices are connected or disconnected from the host
type AttachedMessage struct {
	MessageType string
	DeviceID    int
	Properties  DeviceProperties
}

func attachedFromBytes(plistBytes []byte) (AttachedMessage, error) {
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))
	var obj AttachedMessage
	err := decoder.Decode(&obj)
	if err != nil {
		return obj, err
	}
	return obj, nil
}

//DeviceAttached  checks if the attached message is about a newly added device
func (msg AttachedMessage) DeviceAttached() bool {
	return "Attached" == msg.MessageType
}

//DeviceDetached checks if the attachedMessage is about a disconnected device
func (msg AttachedMessage) DeviceDetached() bool {
	return "Detached" == msg.MessageType
}

//NewListen creates a new Listen Message for USBMUX
func NewListen() ListenType {
	data := ListenType{
		MessageType:         "Listen",
		ProgName:            "go-usbmux",
		ClientVersionString: "usbmuxd-471.8.1",
		//Seems like Conntype is not really needed
		ConnType:          1,
		kLibUSBMuxVersion: 3,
	}
	return data
}

//Listen will send a listen command to usbmuxd which will cause this connection to stay open indefinitely and receive
// messages whenever devices are connected or disconnected
func (muxConn *UsbMuxConnection) Listen() (func() (AttachedMessage, error), error) {
	msg := NewListen()
	err := muxConn.Send(msg)
	if err != nil {
		return nil, err
	}
	response, err := muxConn.ReadMessage()
	if err != nil {
		return nil, err
	}
	if !MuxResponsefromBytes(response.Payload).IsSuccessFull() {
		return nil, fmt.Errorf("Listen command to usbmuxd failed: %x", response.Payload)
	}

	return func() (AttachedMessage, error) {
		mux, err := muxConn.ReadMessage()
		if err != nil {
			return AttachedMessage{}, err
		}
		return attachedFromBytes(mux.Payload)
	}, nil

}
