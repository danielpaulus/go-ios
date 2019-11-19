package screenshotr

import (
	"errors"

	"github.com/danielpaulus/go-ios/usbmux"
)

const serviceName string = "com.apple.mobile.screenshotr"

//Connection exposes the LogReader channel which send the LogMessages as strings.
type Connection struct {
	muxConn    *usbmux.MuxConnection
	plistCodec *usbmux.PlistCodec
	version    versionInfo
}

//New returns a new SysLog Connection for the given DeviceID and Udid
//It will create LogReader as a buffered Channel because Syslog is very verbose.
func New(deviceID int, udid string, pairRecord usbmux.PairRecord) (*Connection, error) {
	startServiceResponse := usbmux.StartService(deviceID, udid, serviceName)
	var screenShotrConn Connection
	screenShotrConn.muxConn = usbmux.NewUsbMuxConnection()
	responseChannel := make(chan []byte)

	plistCodec := usbmux.NewPlistCodec(responseChannel)
	screenShotrConn.plistCodec = plistCodec

	err := screenShotrConn.muxConn.ConnectWithStartServiceResponse(deviceID, *startServiceResponse, plistCodec, pairRecord)
	if err != nil {
		return &Connection{}, err
	}
	screenShotrConn.readVersion()
	screenShotrConn.muxConn.Send(newVersionExchangeRequest(screenShotrConn.version.major))
	screenShotrConn.readExchangeResponse()
	return &screenShotrConn, nil
}

func (screenshotrConn *Connection) readExchangeResponse() error {
	responseBytes := <-screenshotrConn.plistCodec.ResponseChannel
	response := getArrayFromBytes(responseBytes)
	readyMessage, ok := response[0].(string)
	if !ok || readyMessage != "DLMessageDeviceReady" {
		return errors.New("wrong message received")
	}
	return nil
}

func (screenShotrConn *Connection) readVersion() {
	versionBytes := <-screenShotrConn.plistCodec.ResponseChannel
	screenShotrConn.version = getVersionfromBytes(versionBytes)
}

func (screenShotrConn *Connection) TakeScreenshot() []uint8 {
	screenShotrConn.muxConn.Send(newScreenShotRequest())
	responseBytes := <-screenShotrConn.plistCodec.ResponseChannel
	response := getArrayFromBytes(responseBytes)
	responseMap := response[1].(map[string]interface{})
	bytes := responseMap["ScreenShotData"].([]uint8)
	return bytes
}

func (screenShotrConn *Connection) Close() {
	screenShotrConn.muxConn.Close()
}
