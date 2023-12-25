package amfi

import (
	"fmt"

	"github.com/danielpaulus/go-ios/ios"
)

const serviceName string = "com.apple.amfi.lockdown"

type Connection struct {
	deviceConn ios.DeviceConnectionInterface
	plistCodec ios.PlistCodec
}

func New(device ios.DeviceEntry) (*Connection, error) {
	deviceConn, err := ios.ConnectToService(device, serviceName)
	if err != nil {
		return &Connection{}, err
	}

	var devModeConn Connection
	devModeConn.deviceConn = deviceConn
	devModeConn.plistCodec = ios.NewPlistCodec()

	return &devModeConn, nil
}

func (devModeConn *Connection) Close() error {
	return devModeConn.deviceConn.Close()
}

func (devModeConn *Connection) EnableDevMode() error {
	reader := devModeConn.deviceConn.Reader()

	request := map[string]interface{}{"action": 1}

	bytes, err := devModeConn.plistCodec.Encode(request)
	if err != nil {
		return err
	}

	err = devModeConn.deviceConn.Send(bytes)
	if err != nil {
		return err
	}

	responseBytes, err := devModeConn.plistCodec.Decode(reader)
	if err != nil {
		return err
	}

	plist, err := ios.ParsePlist(responseBytes)
	if err != nil {
		return err
	}

	// Check if we have an error returned by the service
	if _, ok := plist["Error"]; ok {
		return fmt.Errorf("amfi could not enable developer mode")
	}

	if _, ok := plist["success"]; ok {
		return nil
	}

	return fmt.Errorf("amfi could not enable developer mode but no error or success was reported")
}

func (devModeConn *Connection) EnableDevModePostRestart() error {
	reader := devModeConn.deviceConn.Reader()

	request := map[string]interface{}{"action": 2}

	bytes, err := devModeConn.plistCodec.Encode(request)
	if err != nil {
		return err
	}

	err = devModeConn.deviceConn.Send(bytes)
	if err != nil {
		return err
	}

	responseBytes, err := devModeConn.plistCodec.Decode(reader)
	if err != nil {
		return err
	}

	plist, err := ios.ParsePlist(responseBytes)
	if err != nil {
		return err
	}

	if _, ok := plist["success"]; ok {
		return nil
	}

	return fmt.Errorf("amfi could not enable developer mode post restart")
}
