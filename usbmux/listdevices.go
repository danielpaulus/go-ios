package usbmux

import (
	"bytes"

	plist "howett.net/plist"
)

//ReadDevicesType contains all the data necessary to request a DeviceList from
//usbmuxd. Can be created with newReadDevices
type ReadDevicesType struct {
	MessageType         string
	ProgName            string
	ClientVersionString string
}

func deviceListfromBytes(plistBytes []byte) DeviceList {
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))
	var deviceList DeviceList
	_ = decoder.Decode(&deviceList)
	return deviceList
}

//Print all udids to the console
func (deviceList DeviceList) Print() {
	for _, element := range deviceList.DeviceList {
		println(element.Properties.SerialNumber)
	}
}

//DeviceList is a simple wrapper for a
//array of  DeviceEntry
type DeviceList struct {
	DeviceList []DeviceEntry
}

//DeviceEntry contains the DeviceID with is sometimes needed
//f.ex. to enable LockdownSSL. More importantly it contains
// DeviceProperties where the udid is stored.
type DeviceEntry struct {
	DeviceID    int
	MessageType string
	Properties  DeviceProperties
}

//DeviceProperties contains important device related info like the udid which is named SerialNumber
//here
type DeviceProperties struct {
	ConnectionSpeed int
	ConnectionType  string
	DeviceID        int
	LocationID      int
	ProductID       int
	SerialNumber    string
}

func newReadDevices() *ReadDevicesType {
	data := &ReadDevicesType{
		MessageType:         "ListDevices",
		ProgName:            "go-usbmux",
		ClientVersionString: "go-usbmux-0.0.1",
	}
	return data
}

//ListDevices returns a DeviceList containing data about all
//currently connected iOS devices
func (muxConn *MuxConnection) ListDevices() DeviceList {
	msg := newReadDevices()
	muxConn.deviceConn.Send(msg)
	response := <-muxConn.ResponseChannel
	return deviceListfromBytes(response)
}

//ListDevices returns a DeviceList containing data about all
//currently connected iOS devices using a new UsbMuxConnection
func ListDevices() DeviceList {
	muxConnection := NewUsbMuxConnection()
	defer muxConnection.Close()
	return muxConnection.ListDevices()
}
