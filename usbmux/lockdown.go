package usbmux

import (
	"bytes"

	plist "howett.net/plist"
)

const lockdownport int = 32498

//LockDownConnection allows you to interact with the Lockdown service on the phone.
//The UsbMuxConnection used to create this LockDownConnection cannot be used anymore.
type LockDownConnection struct {
	deviceConnection *DeviceConnection
	sessionID        string
	ResponseChannel  chan []byte
	plistCodec       *PlistCodec
}

type getValue struct {
	Label   string
	Key     string `plist:"Key,omitempty"`
	Request string
}

func newGetValue(key string) *getValue {
	data := &getValue{
		Label:   "go.ios.control",
		Key:     key,
		Request: "GetValue",
	}
	return data
}

//GetValues retrieves a GetAllValuesResponse containing all values lockdown returns
func (lockDownConn *LockDownConnection) GetValues() GetAllValuesResponse {
	lockDownConn.deviceConnection.Send(newGetValue(""))
	resp := <-lockDownConn.ResponseChannel
	response := getAllValuesResponseFromBytes(resp)
	return response
}

//GetProductVersion returns the ProductVersion of the device f.ex. "10.3"
func (lockDownConn *LockDownConnection) GetProductVersion() string {
	return lockDownConn.GetValue("ProductVersion").(string)
}

//GetValue gets and returns the string value for the lockdown key
func (lockDownConn *LockDownConnection) GetValue(key string) interface{} {
	lockDownConn.deviceConnection.Send(newGetValue(key))
	resp := <-lockDownConn.ResponseChannel
	response := getValueResponsefromBytes(resp)
	return response.Value
}

//GetValueResponse contains the response for a GetValue Request
type GetValueResponse struct {
	Key     string
	Request string
	Value   interface{}
}

func getValueResponsefromBytes(plistBytes []byte) GetValueResponse {
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))
	var getValueResponse GetValueResponse
	_ = decoder.Decode(&getValueResponse)
	return getValueResponse
}

func getAllValuesResponseFromBytes(plistBytes []byte) GetAllValuesResponse {
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))
	var getAllValuesResponse GetAllValuesResponse
	_ = decoder.Decode(&getAllValuesResponse)
	return getAllValuesResponse
}
