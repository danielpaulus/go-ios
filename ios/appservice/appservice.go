package appservice

import (
	"encoding/base64"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/xpc"
)

func New(deviceEntry ios.DeviceEntry) (*xpc.Connection, error) {
	xpcConn, err := ios.ConnectToServiceTunnelIface(deviceEntry, "com.apple.coredevice.appservice")
	if err != nil {
		return nil, err
	}

	print("We have a connection: ")
	print(xpcConn)

	return xpcConn, nil
}

func LaunchApp(conn *xpc.Connection, deviceId string, bundleId string, args []interface{}, env map[string]interface{}) {
	msg := map[string]interface{}{
		"CoreDevice.CoreDeviceDDIProtocolVersion": int64(0),
		"CoreDevice.action":                       map[string]interface{}{},
		"CoreDevice.coreDeviceVersion": map[string]interface{}{
			"components":              []interface{}{uint64(0x15c), uint64(0x1), uint64(0x0), uint64(0x0), uint64(0x0)},
			"originalComponentsCount": int64(2),
			"stringValue":             "348.1",
		},
		"CoreDevice.deviceIdentifier":  deviceId,
		"CoreDevice.featureIdentifier": "com.apple.coredevice.feature.launchapplication",
		"CoreDevice.input": map[string]interface{}{
			"applicationSpecifier": map[string]interface{}{
				"bundleIdentifier": map[string]interface{}{
					"_0": bundleId,
				},
			},
			"options": map[string]interface{}{
				"arguments":                     args,
				"environmentVariables":          env,
				"platformSpecificOptions":       base64Decode("YnBsaXN0MDDQCAAAAAAAAAEBAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAJ"),
				"standardIOUsesPseudoterminals": true,
				"startStopped":                  false,
				"terminateExisting":             false,
				"user": map[string]interface{}{
					"active": true,
				},
				"workingDirectory": nil,
			},
			"standardIOIdentifiers": map[string]interface{}{},
		},
		"CoreDevice.invocationIdentifier": "62419FC1-5ABF-4D96-BCA8-7A5F6F9A69EE",
	}

	conn.Send(msg)
}

func base64Decode(s string) []byte {
	dst := make([]byte, base64.StdEncoding.DecodedLen(len(s)))
	_, err := base64.StdEncoding.Decode(dst, []byte(s))
	if err != nil {
		panic(err)
	}
	return dst
}
