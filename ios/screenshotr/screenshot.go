package screenshotr

import (
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"

	ios "github.com/danielpaulus/go-ios/ios"
)

const serviceName string = "com.apple.mobile.screenshotr"

// Connection exposes the LogReader channel which send the LogMessages as strings.
type Connection struct {
	deviceConn ios.DeviceConnectionInterface
	plistCodec ios.PlistCodec
	version    versionInfo
}

// New returns a new SysLog Connection for the given DeviceID and Udid
// It will create LogReader as a buffered Channel because Syslog is very verbose.
func New(device ios.DeviceEntry) (*Connection, error) {
	deviceConn, err := ios.ConnectToService(device, serviceName)
	if err != nil {
		return &Connection{}, err
	}

	var screenShotrConn Connection
	screenShotrConn.deviceConn = deviceConn
	screenShotrConn.plistCodec = ios.NewPlistCodec()
	reader := screenShotrConn.deviceConn.Reader()
	err = screenShotrConn.readVersion(reader)
	if err != nil {
		return &screenShotrConn, fmt.Errorf("failed reading version from screenshotr with err: %w", err)
	}

	bytes, err := screenShotrConn.plistCodec.Encode(newVersionExchangeRequest(screenShotrConn.version.major))
	if err != nil {
		return &screenShotrConn, fmt.Errorf("failed plist encoding version for screenshotr with err: %w", err)
	}

	err = screenShotrConn.deviceConn.Send(bytes)
	if err != nil {
		return &screenShotrConn, fmt.Errorf("failed sending version exchange message to screenshotr with err: %w", err)
	}

	err = screenShotrConn.readExchangeResponse(reader)
	if err != nil {
		return &screenShotrConn, fmt.Errorf("failed reading version exchange response from screenshotr with err: %w", err)
	}

	return &screenShotrConn, nil
}

func (screenShotrConn *Connection) readExchangeResponse(reader io.Reader) error {
	responseBytes, err := screenShotrConn.plistCodec.Decode(reader)
	if err != nil {
		return err
	}

	response, err := getArrayFromBytes(responseBytes)
	if err != nil {
		return fmt.Errorf("could not decode %x to an array", responseBytes)
	}
	readyMessage, ok := response[0].(string)
	if !ok || readyMessage != "DLMessageDeviceReady" {
		return fmt.Errorf("wrong message received: '%s'", readyMessage)
	}
	return nil
}

func (screenShotrConn *Connection) readVersion(reader io.Reader) error {
	versionBytes, err := screenShotrConn.plistCodec.Decode(reader)
	if err != nil {
		return err
	}
	screenShotrConn.version, err = getVersionfromBytes(versionBytes)
	log.Debugf("screenshotr version: %v", screenShotrConn.version)
	return err
}

// TakeScreenshot uses Screenshotr to get a screenshot as a byteslice
func (screenShotrConn *Connection) TakeScreenshot() ([]uint8, error) {
	reader := screenShotrConn.deviceConn.Reader()
	bytes, err := screenShotrConn.plistCodec.Encode(newScreenShotRequest())
	if err != nil {
		return make([]uint8, 0), err
	}
	err = screenShotrConn.deviceConn.Send(bytes)
	if err != nil {
		return make([]uint8, 0), err
	}
	responseBytes, err := screenShotrConn.plistCodec.Decode(reader)
	if err != nil {
		return make([]uint8, 0), err
	}
	response, err := getArrayFromBytes(responseBytes)
	if err != nil {
		return make([]uint8, 0), err
	}

	if len(response) < 2 {
		return make([]uint8, 0), fmt.Errorf("response only contained one field %+v", response)
	}
	msg, ok := response[0].(string)
	if !ok || msg != dlMessageProcessMessage {
		log.Warnf("expected DLMessageProcessMessage but got '%s'", msg)
	}
	responseMap, ok := response[1].(map[string]interface{})
	if !ok {
		return make([]uint8, 0), fmt.Errorf("could not decode screenshot, expected map %+v", response)
	}
	screenshotData, ok := responseMap["ScreenShotData"]
	if !ok {
		return make([]uint8, 0), fmt.Errorf("could not find ScreenShotData: %+v", responseMap)
	}
	screenshotBytes, ok := screenshotData.([]uint8)
	if !ok {
		return make([]uint8, 0), fmt.Errorf("ScreenShotData not []uint8 but was %+v", screenshotData)
	}
	return screenshotBytes, nil
}

// Close closes the underlying DeviceConnection
func (screenShotrConn *Connection) Close() error {
	return screenShotrConn.deviceConn.Close()
}
