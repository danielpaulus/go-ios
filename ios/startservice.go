package ios

import (
	"bytes"
	"fmt"

	log "github.com/sirupsen/logrus"
	plist "howett.net/plist"
)

type startServiceRequest struct {
	Label   string
	Request string
	Service string
}

//StartServiceResponse is sent by the phone after starting a service, it contains servicename, port and tells us
//whether we should enable SSL or not.
type StartServiceResponse struct {
	Port             uint16
	Request          string
	Service          string
	EnableServiceSSL bool
	Error            string
}

func getStartServiceResponsefromBytes(plistBytes []byte) StartServiceResponse {
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))
	var data StartServiceResponse
	_ = decoder.Decode(&data)
	return data
}

//StartService sends a StartServiceRequest using the provided serviceName
//and returns the Port of the services in a BigEndian Integer.
//This port cann be used with a new UsbMuxClient and the Connect call.
func (lockDownConn *LockDownConnection) StartService(serviceName string) (StartServiceResponse, error) {
	err := lockDownConn.Send(startServiceRequest{Label: "go.ios.control", Request: "StartService", Service: serviceName})
	if err != nil {
		return StartServiceResponse{}, err
	}
	resp, err := lockDownConn.ReadMessage()
	if err != nil {
		return StartServiceResponse{}, err
	}
	response := getStartServiceResponsefromBytes(resp)
	if response.Error != "" {
		return StartServiceResponse{}, fmt.Errorf("Could not start service:%s with reason:'%s'. Have you mounted the Developer Image?", serviceName, response.Error)
	}
	log.WithFields(log.Fields{"Port": response.Port, "Request": response.Request, "Service": response.Service, "EnableServiceSSL": response.EnableServiceSSL}).Debug("Service started on device")
	return response, nil
}

//StartService conveniently starts a service on a device and cleans up the used UsbMuxconnection.
//It returns the service port as a uint16 in BigEndian byte order.
func StartService(device DeviceEntry, serviceName string) (StartServiceResponse, error) {
	lockdown, err := ConnectLockdownWithSession(device)
	if err != nil {
		return StartServiceResponse{}, err
	}
	defer lockdown.Close()
	response, err := lockdown.StartService(serviceName)
	if err != nil {
		return response, err
	}
	return response, nil
}
