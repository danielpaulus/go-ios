package usbmux

import (
	"bytes"

	log "github.com/sirupsen/logrus"
	plist "howett.net/plist"
)

const Lockdownport int = 32498

//LockDownConnection allows you to interact with the Lockdown service on the phone.
//The UsbMuxConnection used to create this LockDownConnection cannot be used anymore.
type LockDownConnection struct {
	deviceConnection DeviceConnectionInterface
	sessionID        string
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

//Send takes a go struct, converts it to a PLIST and sends it with a 4 byte length field.
func (lockDownConn LockDownConnection) Send(msg interface{}) error {
	bytes, err := lockDownConn.plistCodec.Encode(msg)
	if err != nil {
		log.Error("failed lockdown send")
		return err
	}
	return lockDownConn.deviceConnection.Send(bytes)
}

//ReadMessage grabs the next LockDown Message using the PlistDecoder from the underlying
//DeviceConnection and returns the Plist as a byte slice.
func (lockDownConn *LockDownConnection) ReadMessage() ([]byte, error) {
	reader := lockDownConn.deviceConnection.Reader()
	resp, err := lockDownConn.plistCodec.Decode(reader)
	if err != nil {
		return make([]byte, 0), err
	}
	return resp, err
}

//GetValues retrieves a GetAllValuesResponse containing all values lockdown returns
func (lockDownConn *LockDownConnection) GetValues() (GetAllValuesResponse, error) {
	lockDownConn.Send(newGetValue(""))
	resp, err := lockDownConn.ReadMessage()

	response := getAllValuesResponseFromBytes(resp)
	return response, err
}

//GetProductVersion returns the ProductVersion of the device f.ex. "10.3"
func (lockDownConn *LockDownConnection) GetProductVersion() string {
	msg, err := lockDownConn.GetValue("ProductVersion")
	if err != nil {
		log.Fatal("Failed getting ProductVersion", err)
	}
	return msg.(string)
}

//GetValue gets and returns the string value for the lockdown key
func (lockDownConn *LockDownConnection) GetValue(key string) (interface{}, error) {
	lockDownConn.Send(newGetValue(key))
	resp, err := lockDownConn.ReadMessage()
	response := getValueResponsefromBytes(resp)
	return response.Value, err
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
