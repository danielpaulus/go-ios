package afc

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/danielpaulus/go-ios/ios"

	"howett.net/plist"
)

const serviceName = "com.apple.mobile.house_arrest"

type Connection struct {
	deviceConn    ios.DeviceConnectionInterface
	packageNumber uint64
}

//NewFromConn allows to use AFC on a DeviceConnectionInterface, see crashreport for an example
func NewFromConn(deviceConn ios.DeviceConnectionInterface) *Connection {
	return &Connection{deviceConn: deviceConn}
}

//NewWithHouseArrest gives an AFC connection to an app container using the house arrest service
func NewWithHouseArrest(device ios.DeviceEntry, bundleID string) (*Connection, error) {
	deviceConn, err := ios.ConnectToService(device, serviceName)
	if err != nil {
		return &Connection{}, err
	}
	err = vendContainer(deviceConn, bundleID)
	if err != nil {
		return &Connection{}, err
	}
	return &Connection{deviceConn: deviceConn}, nil
}

func vendContainer(deviceConn ios.DeviceConnectionInterface, bundleID string) error {
	plistCodec := ios.NewPlistCodec()
	vendContainer := map[string]interface{}{"Command": "VendContainer", "Identifier": bundleID}
	msg, err := plistCodec.Encode(vendContainer)
	if err != nil {
		return fmt.Errorf("VendContainer Encoding cannot fail unless the encoder is broken: %v", err)
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

func (conn Connection) Close() error {
	return conn.deviceConn.Close()
}
