package screenshotr

/*
import (
	"errors"

	"github.com/danielpaulus/go-ios/usbmux"
)

const serviceName string = "com.apple.mobile.screenshotr"

//Connection exposes the LogReader channel which send the LogMessages as strings.
type Connection struct {
	deviceConn usbmux.DeviceConnectionInterface
	plistCodec *usbmux.PlistCodec
	version    versionInfo
}

//New returns a new SysLog Connection for the given DeviceID and Udid
//It will create LogReader as a buffered Channel because Syslog is very verbose.
func New(deviceID int, udid string, pairRecord usbmux.PairRecord) (*Connection, error) {
	startServiceResponse, err := usbmux.StartService(deviceID, udid, serviceName)

	muxConn := usbmux.NewUsbMuxConnection()

	err = muxConn.ConnectWithStartServiceResponse(deviceID, *startServiceResponse, pairRecord)
	if err != nil {
		return &Connection{}, err
	}

	var screenShotrConn Connection
	screenShotrConn.plistCodec = usbmux.NewPlistCodec()
	screenShotrConn.readVersion()
	bytes, err := screenShotrConn.plistCodec.Encode(newVersionExchangeRequest(screenShotrConn.version.major))
	screenShotrConn.deviceConn.Send(bytes)
	screenShotrConn.readExchangeResponse()
	return &screenShotrConn, nil
}

func (screenshotrConn *Connection) readExchangeResponse() error {
	reader := screenshotrConn.deviceConn.Reader()
	response, err := screenshotrConn.plistCodec.Decode(reader)

	response2 := getArrayFromBytes(response)
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
	screenShotrConn.deviceConn.Send(newScreenShotRequest())
	responseBytes := <-screenShotrConn.plistCodec.ResponseChannel
	response := getArrayFromBytes(responseBytes)
	responseMap := response[1].(map[string]interface{})
	bytes := responseMap["ScreenShotData"].([]uint8)
	return bytes
}

func (screenShotrConn *Connection) Close() {
	screenShotrConn.deviceConn.Close()
}
*/
