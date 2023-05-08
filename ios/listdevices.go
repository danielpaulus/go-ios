package ios

import (
	"bytes"
	"fmt"
	"strings"

	plist "howett.net/plist"
)

//ReadDevicesType contains all the data necessary to request a DeviceList from
//usbmuxd. Can be created with newReadDevices
type ReadDevicesType struct {
	MessageType         string
	ProgName            string
	ClientVersionString string
}

//DeviceListfromBytes parses a DeviceList from a byte array
func DeviceListfromBytes(plistBytes []byte) DeviceList {
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))
	var deviceList DeviceList
	_ = decoder.Decode(&deviceList)
	return deviceList
}

//String returns a list of all udids in a formatted string
func (deviceList DeviceList) String() string {
	var sb strings.Builder
	for _, element := range deviceList.DeviceList {
		sb.WriteString(element.Properties.SerialNumber)
		sb.WriteString("\n")
	}
	return sb.String()
}

//CreateMapForJSONConverter creates a simple json ready map containing all UDIDs
func (deviceList DeviceList) CreateMapForJSONConverter() map[string]interface{} {
	devices := make([]string, len(deviceList.DeviceList))
	for i, element := range deviceList.DeviceList {
		devices[i] = element.Properties.SerialNumber
	}
	return map[string]interface{}{"deviceList": devices}
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

// NewReadDevices creates a struct containing a request for a device list that can be sent
// to UsbMuxD.
func NewReadDevices() ReadDevicesType {
	data := ReadDevicesType{
		MessageType:         "ListDevices",
		ProgName:            "go-usbmux",
		ClientVersionString: "go-usbmux-0.0.1",
	}
	return data
}

//ListDevices returns a DeviceList containing data about all
//currently connected iOS devices
func (muxConn *UsbMuxConnection) ListDevices() (DeviceList, error) {
	err := muxConn.Send(NewReadDevices())
	if err != nil {
		return DeviceList{}, fmt.Errorf("Failed sending to usbmux requesting devicelist: %v", err)
	}
	response, err := muxConn.ReadMessage()
	if err != nil {
		return DeviceList{}, fmt.Errorf("Failed getting devicelist: %v", err)
	}
	return DeviceListfromBytes(response.Payload), nil
}

//ListDevices returns a DeviceList containing data about all
//currently connected iOS devices using a new UsbMuxConnection
func ListDevices() (DeviceList, error) {
	muxConnection, err := NewUsbMuxConnectionSimple()
	if err != nil {
		return DeviceList{}, err
	}
	defer muxConnection.Close()
	return muxConnection.ListDevices()
}
