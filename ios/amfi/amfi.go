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

// Enable developer mode on a device, e.g. after content reset
func (devModeConn *Connection) EnableDevMode() error {
	reader := devModeConn.deviceConn.Reader()

	request := map[string]interface{}{"action": 1}

	bytes, err := devModeConn.plistCodec.Encode(request)
	if err != nil {
		return fmt.Errorf("EnableDevMode: failed encoding request to service with err: %w", err)
	}

	err = devModeConn.deviceConn.Send(bytes)
	if err != nil {
		return fmt.Errorf("EnableDevMode: failed sending request bytes to service with err: %w", err)
	}

	responseBytes, err := devModeConn.plistCodec.Decode(reader)
	if err != nil {
		return fmt.Errorf("EnableDevMode: failed decoding response from service with err: %w", err)
	}

	plist, err := ios.ParsePlist(responseBytes)
	if err != nil {
		return fmt.Errorf("EnableDevMode: failed parsing response plist with err: %w", err)
	}

	// Check if we have an error returned by the service
	if _, ok := plist["Error"]; ok {
		return fmt.Errorf("EnableDevMode: could not enable developer mode through amfi service")
	}

	if _, ok := plist["success"]; ok {
		return nil
	}

	return fmt.Errorf("EnableDevMode: could not enable developer mode through amfi service but no error or success was reported")
}

// When you enable developer mode and device is rebooted, you get a popup on the device to finish enabling developer mode
// This function "accepts" that popup
func (devModeConn *Connection) EnableDevModePostRestart() error {
	reader := devModeConn.deviceConn.Reader()

	request := map[string]interface{}{"action": 2}

	bytes, err := devModeConn.plistCodec.Encode(request)
	if err != nil {
		return fmt.Errorf("EnableDevModePostRestart: failed encoding request to service with err: %w", err)
	}

	err = devModeConn.deviceConn.Send(bytes)
	if err != nil {
		return fmt.Errorf("EnableDevModePostRestart: failed sending request bytes to service with err: %w", err)
	}

	responseBytes, err := devModeConn.plistCodec.Decode(reader)
	if err != nil {
		return fmt.Errorf("EnableDevModePostRestart: failed decoding response from service with err: %w", err)
	}

	plist, err := ios.ParsePlist(responseBytes)
	if err != nil {
		return fmt.Errorf("EnableDevModePostRestart: failed parsing response plist with err: %w", err)
	}

	if _, ok := plist["success"]; ok {
		return nil
	}

	return fmt.Errorf("EnableDevModePostRestart: could not enable developer mode post restart through amfi service")
}
