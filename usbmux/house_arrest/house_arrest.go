package house_arrest

import (
	"bytes"
	"errors"
	"log"

	"github.com/danielpaulus/go-ios/usbmux"
	"howett.net/plist"
)

const serviceName = "com.apple.mobile.house_arrest"

type Connection struct {
	deviceConn usbmux.DeviceConnectionInterface
}

func New(deviceID int, udid string, bundleID string) (*Connection, error) {
	deviceConn, err := usbmux.ConnectToService(deviceID, udid, serviceName)
	if err != nil {
		return &Connection{}, err
	}
	err = vendContainer(deviceConn, bundleID)
	if err != nil {
		return &Connection{}, err
	}
	return &Connection{deviceConn: deviceConn}, nil
}

func vendContainer(deviceConn usbmux.DeviceConnectionInterface, bundleID string) error {
	plistCodec := usbmux.NewPlistCodec()
	vendContainer := map[string]interface{}{"Command": "VendContainer", "Identifier": bundleID}
	msg, err := plistCodec.Encode(vendContainer)
	if err != nil {
		log.Fatal("VendContainer Encoding cannot fail unless the encoder is broken")
	}
	err = deviceConn.Send(msg)
	if err != nil {
		return err
	}
	reader := deviceConn.Reader()
	response, err := plistCodec.Decode(reader)
	if err != nil {
		return err
	}
	return checkResponse(response)
}

func checkResponse(vendContainerResponseBytes []byte) error {
	response, err := plistFromBytes(vendContainerResponseBytes)
	if err != nil {
		return err
	}
	if "Complete" == response.Status {
		return nil
	}
	if response.Error != "" {
		return errors.New(response.Error)
	}
	return errors.New("unknown error during vendcontainer")
}

func plistFromBytes(plistBytes []byte) (vendContainerResponse, error) {
	var vendResponse vendContainerResponse
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))

	err := decoder.Decode(&vendResponse)
	if err != nil {
		return vendResponse, err
	}
	return vendResponse, nil
}

type vendContainerResponse struct {
	Status string
	Error  string
}

func (c Connection) Close() {
	c.deviceConn.Close()
}
