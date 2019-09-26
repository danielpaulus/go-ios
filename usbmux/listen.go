package usbmux

import (
	"bytes"
	"encoding/hex"
	"errors"
	"howett.net/plist"
)

type ListenType struct {
	MessageType         string
	ProgName            string
	ClientVersionString string
	ConnType            int
	kLibUSBMuxVersion   int
}

//Contains some info about when iOS devices are connected or disconnected from the host
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

func (msg AttachedMessage) DeviceAttached() bool {
	return "Attached" == msg.MessageType
}

func (msg AttachedMessage) DeviceDetached() bool {
	return "Detached" == msg.MessageType
}

func NewListen() *ListenType {
	data := &ListenType{
		MessageType:         "Listen",
		ProgName:            "go-usbmux",
		ClientVersionString: "usbmuxd-471.8.1",
		//dunno if conntype is needed
		ConnType:          1,
		kLibUSBMuxVersion: 3,
	}
	return data
}

//Listen will send a listen command to usbmuxd which will cause this connection to stay open indefinitely and receive
// messages whenever devices are connected or disconnected
func (muxConn *MuxConnection) Listen() (func() (AttachedMessage, error), error) {
	msg := NewListen()
	muxConn.Send(msg)
	response := <-muxConn.ResponseChannel
	if !MuxResponsefromBytes(response).IsSuccessFull() {
		return nil, errors.New("Listen command to usbmuxd failed:" + hex.Dump(response))
	}

	return func() (AttachedMessage, error) {
		return attachedFromBytes(<-muxConn.ResponseChannel)
	}, nil
}
