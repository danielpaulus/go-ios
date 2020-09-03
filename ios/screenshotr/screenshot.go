package screenshotr

import (
	"errors"
	"io"

	ios "github.com/danielpaulus/go-ios/ios"
)

const serviceName string = "com.apple.mobile.screenshotr"

//Connection exposes the LogReader channel which send the LogMessages as strings.
type Connection struct {
	deviceConn ios.DeviceConnectionInterface
	plistCodec ios.PlistCodec
	version    versionInfo
}

//New returns a new SysLog Connection for the given DeviceID and Udid
//It will create LogReader as a buffered Channel because Syslog is very verbose.
func New(deviceID int, udid string, pairRecord ios.PairRecord) (*Connection, error) {
	startServiceResponse, err := ios.StartService(deviceID, udid, serviceName)

	muxConn := ios.NewUsbMuxConnection()

	err = muxConn.ConnectWithStartServiceResponse(deviceID, *startServiceResponse, pairRecord)
	if err != nil {
		return &Connection{}, err
	}

	var screenShotrConn Connection
	screenShotrConn.deviceConn = muxConn.Close()
	screenShotrConn.plistCodec = ios.NewPlistCodec()
	reader := screenShotrConn.deviceConn.Reader()
	screenShotrConn.readVersion(reader)
	bytes, err := screenShotrConn.plistCodec.Encode(newVersionExchangeRequest(screenShotrConn.version.major))
	screenShotrConn.deviceConn.Send(bytes)
	screenShotrConn.readExchangeResponse(reader)
	return &screenShotrConn, nil
}

func (screenShotrConn *Connection) readExchangeResponse(reader io.Reader) error {

	responseBytes, err := screenShotrConn.plistCodec.Decode(reader)
	if err != nil {
		return err
	}

	response := getArrayFromBytes(responseBytes)
	readyMessage, ok := response[0].(string)
	if !ok || readyMessage != "DLMessageDeviceReady" {
		return errors.New("wrong message received")
	}
	return nil
}

func (screenShotrConn *Connection) readVersion(reader io.Reader) error {

	versionBytes, err := screenShotrConn.plistCodec.Decode(reader)
	if err != nil {
		return err
	}
	screenShotrConn.version = getVersionfromBytes(versionBytes)
	return nil
}

func (screenShotrConn *Connection) TakeScreenshot() ([]uint8, error) {
	reader := screenShotrConn.deviceConn.Reader()
	bytes, err := screenShotrConn.plistCodec.Encode(newScreenShotRequest())
	if err != nil {
		return make([]uint8, 0), err
	}
	screenShotrConn.deviceConn.Send(bytes)
	responseBytes, err := screenShotrConn.plistCodec.Decode(reader)
	if err != nil {
		return make([]uint8, 0), err
	}
	response := getArrayFromBytes(responseBytes)
	responseMap := response[1].(map[string]interface{})
	screenshotBytes := responseMap["ScreenShotData"].([]uint8)
	return screenshotBytes, nil
}

func (screenShotrConn *Connection) Close() {
	screenShotrConn.deviceConn.Close()
}
