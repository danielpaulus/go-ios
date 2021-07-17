package imagemounter

import (
	"fmt"
	"github.com/danielpaulus/go-ios/ios"
)

const serviceName string = "com.apple.mobile.mobile_image_mounter"

//Connection to mobile image mounter
type Connection struct {
	deviceConn ios.DeviceConnectionInterface
	plistCodec ios.PlistCodec
}

//New returns a new mobile image mounter Connection for the given DeviceID and Udid
func New(device ios.DeviceEntry) (*Connection, error) {
	deviceConn, err := ios.ConnectToService(device, serviceName)
	if err != nil {
		return &Connection{}, err
	}
	return &Connection{
		deviceConn: deviceConn,
		plistCodec: ios.NewPlistCodec(),
	}, nil
}

//ListImages returns a list with signatures of installed developer images
func (conn *Connection) ListImages() ([][]byte, error) {
	req := map[string]interface{}{
		"Command":   "LookupImage",
		"ImageType": "Developer",
	}
	bytes, err := conn.plistCodec.Encode(req)
	if err != nil {
		return nil, err
	}

	err = conn.deviceConn.Send(bytes)
	if err != nil {
		return nil, err
	}

	bytes, err = conn.plistCodec.Decode(conn.deviceConn.Reader())

	resp, err := ios.ParsePlist(bytes)
	if err != nil {
		return nil, err
	}
	deviceError, ok := resp["Error"]
	if ok {
		return nil, fmt.Errorf("device error: %v", deviceError)
	}
	signatures, ok := resp["ImageSignature"]
	if !ok {
		return nil, fmt.Errorf("invalid response: %+v", signatures)
	}

	array, ok := signatures.([]interface{})
	result := make([][]byte, len(array))
	for i, intf := range array {
		bytes, ok := intf.([]byte)
		if !ok {
			return nil, fmt.Errorf("could not convert %+v to byte slice", intf)
		}
		result[i] = bytes
	}
	return result, nil
}

//Close closes the underlying UsbMuxConnection
func (conn *Connection) Close() {
	conn.deviceConn.Close()
}
