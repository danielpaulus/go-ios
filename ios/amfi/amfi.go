package amfi

import (
	"errors"
	"fmt"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/danielpaulus/go-ios/ios/imagemounter"
)

const serviceName string = "com.apple.amfi.lockdown"

const logModule = "go-ios/amfi"

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

// RevealDevMode sends action 0 to the AMFI service, which makes the
// Developer Mode toggle visible in Settings → Privacy & Security.
// This replicates what Xcode does when it detects Developer Mode is disabled.
func (devModeConn *Connection) RevealDevMode() error {
	reader := devModeConn.deviceConn.Reader()

	request := map[string]interface{}{"action": 0}

	bytes, err := devModeConn.plistCodec.Encode(request)
	if err != nil {
		return fmt.Errorf("RevealDevMode: failed encoding request to service with err: %w", err)
	}

	err = devModeConn.deviceConn.Send(bytes)
	if err != nil {
		return fmt.Errorf("RevealDevMode: failed sending request bytes to service with err: %w", err)
	}

	responseBytes, err := devModeConn.plistCodec.Decode(reader)
	if err != nil {
		return fmt.Errorf("RevealDevMode: failed decoding response from service with err: %w", err)
	}

	plist, err := ios.ParsePlist(responseBytes)
	if err != nil {
		return fmt.Errorf("RevealDevMode: failed parsing response plist with err: %w", err)
	}

	if errorMsg, ok := plist["Error"]; ok {
		return fmt.Errorf("RevealDevMode: could not reveal developer mode menu: %s", errorMsg)
	}

	if _, ok := plist["success"]; ok {
		return nil
	}

	return fmt.Errorf("RevealDevMode: no error or success was reported")
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
	if errorMsg, ok := plist["Error"]; ok {
		return fmt.Errorf("EnableDevMode: could not enable developer mode through amfi service with error: %s", errorMsg)
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
	// Developer Mode was introduced in iOS 16; older devices have no such
	// toggle, so fail with a clear explanation instead of an opaque service error.
	if v, err := ios.GetProductVersion(device); err == nil && v != nil && v.LessThan(ios.IOS16()) {
		return fmt.Errorf("EnableDeveloperMode: Developer Mode was introduced in iOS 16, but this device runs iOS %s — there is no Developer Mode to enable", v.String())
	}

	// Don't try to enable if it already is
	devModeEnabled, err := imagemounter.IsDevModeEnabled(device)
	if err != nil {
		return fmt.Errorf("EnableDeveloperMode: failed checking developer mode status with err: %w", err)
	}

	if devModeEnabled {
		golog.Info("Developer mode is already enabled on the device", "module", logModule, "udid", device.Properties.SerialNumber)
		return nil
	}

	// Perform the first step of developer mode enablement and wait for the device to restart
	conn, err := New(device)
	if err != nil {
		return fmt.Errorf("EnableDeveloperMode: failed connecting to amfi service with err: %w", err)
	}

	err = conn.EnableDevMode()
	if err != nil {
		conn.Close()
		// If ARM fails (e.g. passcode set), at least reveal the menu
		revealConn, revealErr := New(device)
		if revealErr == nil {
			_ = revealConn.RevealDevMode()
			revealConn.Close()
		}
		return fmt.Errorf("EnableDeveloperMode: failed enabling developer mode with err: %w (Developer Mode menu has been revealed in Settings)", err)
	}
	golog.Info("Successfully enabled developer mode on device, device will restart", "module", logModule, "udid", device.Properties.SerialNumber)

	udid := device.Properties.SerialNumber
	golog.Info("Waiting for device to restart after enabling developer mode", "module", logModule, "udid", udid)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Loop trying to reinit the device to find out if it restarted
WaitLoop:
	for {
		select {
		case <-ticker.C:
			device, err = ios.GetDevice(udid)
			if err != nil {
				golog.Info("Device is not yet available", "module", logModule, "udid", udid)
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
	golog.Info("Device was successfully restarted after enabling developer mode", "module", logModule, "udid", udid)

	// Try to also enable dev mode after the device restarts - skips the system popup that asks you to finalize dev mode enablement
	if enablePostRestart {
		golog.Info("Will attempt to enable developer mode post restart", "module", logModule, "udid", udid)
		conn, err = New(device)
		if err != nil {
			return fmt.Errorf("EnableDeveloperMode: failed connecting to amfi service post restart with err: %w", err)
		}
		defer conn.Close()
		err = conn.EnableDevModePostRestart()
		if err != nil {
			return fmt.Errorf("EnableDeveloperMode: failed enabling developer mode post restart, you need to finish the set up manually through the popup on the device, err: %w", err)
		}
		golog.Info("Successfully enabled developer mode on device post restart", "module", logModule, "udid", udid)
	}

	return nil
}
