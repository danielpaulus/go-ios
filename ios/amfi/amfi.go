package amfi

import (
	"errors"
	"fmt"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/imagemounter"
	log "github.com/sirupsen/logrus"
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

func EnableDeveloperMode(device ios.DeviceEntry, enablePostRestart bool) error {
	// Don't try to enable if it already is
	devModeEnabled, err := imagemounter.IsDevModeEnabled(device)
	if err != nil {
		return fmt.Errorf("EnableDeveloperMode: failed checking developer mode status with err: %w", err)
	}

	if devModeEnabled {
		log.Info("Developer mode is already enabled on the device")
		return nil
	}

	// Perform the first step of developer mode enablement and wait for the device to restart
	conn, err := New(device)
	if err != nil {
		return fmt.Errorf("EnableDeveloperMode: failed connecting to amfi service with err: %w", err)
	}

	err = conn.EnableDevMode()
	if err != nil {
		return fmt.Errorf("EnableDeveloperMode: failed enabling developer mode with err: %w", err)
	}
	log.Infof("Successfully enabled developer mode on device `%s`, device will restart", device.Properties.SerialNumber)

	udid := device.Properties.SerialNumber
	log.Info("Waiting for device to restart after enabling developer mode")
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Loop trying to reinit the device to find out if it restarted
WaitLoop:
	for {
		select {
		case <-ticker.C:
			device, err = ios.GetDevice(udid)
			if err != nil {
				log.Info("Device is not yet available")
				continue WaitLoop
			}
			break WaitLoop
		case <-time.After(60 * time.Second):
			ticker.Stop()
			if err != nil {
				return errors.New("Device was not restarted in 60 seconds")
			}
		}
	}
	log.Info("Device was successfully restarted after enabling developer mode")

	// Try to also enable dev mode after the device restarts - skips the system popup that asks you to finalize dev mode enablement
	if enablePostRestart {
		log.Info("Will attempt to enable developer mode post restart")
		conn, err = New(device)
		if err != nil {
			return fmt.Errorf("EnableDeveloperMode: failed connecting to amfi service post restart with err: %w", err)
		}
		defer conn.Close()
		err = conn.EnableDevModePostRestart()
		if err != nil {
			return fmt.Errorf("EnableDeveloperMode: failed enabling developer mode post restart, you need to finish the set up manually through the popup on the device, err: %w", err)
		}
		log.Info("Successfully enabled developer mode on device post restart")
	}

	return nil
}
