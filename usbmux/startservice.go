package usbmux

import (
	"bytes"

	log "github.com/sirupsen/logrus"
	plist "howett.net/plist"
)

type startServiceRequest struct {
	Label   string
	Request string
	Service string
}

func newStartServiceRequest(serviceName string) *startServiceRequest {
	var req startServiceRequest
	req.Label = "go.ios.control"
	req.Request = "StartService"
	req.Service = serviceName
	return &req
}

type startServiceResponse struct {
	Port    uint16
	Request string
	Service string
}

func getStartServiceResponsefromBytes(plistBytes []byte) *startServiceResponse {
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))
	var data startServiceResponse
	_ = decoder.Decode(&data)
	return &data
}

//StartService sends a StartServiceRequest using the provided serviceName
//and returns the Port of the services in a BigEndian Integer.
//This port cann be used with a new UsbMuxClient and the Connect call.
func (lockDownConn *LockDownConnection) StartService(serviceName string) uint16 {
	lockDownConn.Send(newStartServiceRequest(serviceName))
	resp := <-lockDownConn.ResponseChannel
	response := getStartServiceResponsefromBytes(resp)
	return response.Port
}

//StartService conveniently starts a service on a device and cleans up the used UsbMuxconnection.
//It returns the service port as a uint16 in BigEndian byte order.
func StartService(deviceID int, udid string, serviceName string) uint16 {
	muxConnection := NewUsbMuxConnection()
	defer muxConnection.Close()
	pairRecord := muxConnection.ReadPair(udid)
	lockdown, err := muxConnection.ConnectLockdown(deviceID)
	if err != nil {
		log.Fatal(err)
	}
	lockdown.StartSession(pairRecord)
	port := lockdown.StartService(serviceName)
	lockdown.StopSession()
	return port
}
