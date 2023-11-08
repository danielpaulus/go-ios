package appservice

import (
	"bytes"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/xpc"

	plist "howett.net/plist"
)

type Connection struct {
	conn *xpc.Connection
}

func New(deviceEntry ios.DeviceEntry) (*Connection, error) {
	xpcConn, err := ios.ConnectToXpcServiceTunnelIface(deviceEntry, "com.apple.coredevice.appservice")
	if err != nil {
		return nil, err
	}

	print("We have a connection: ")
	print(xpcConn)

	return &Connection{conn: xpcConn}, nil
}

func (c *Connection) LaunchApp(deviceId string, bundleId string, args []interface{}, env map[string]interface{}) error {
	msg := buildAppLaunchPayload(deviceId, bundleId, args, env)
	return c.conn.Send(msg)
}

func (c *Connection) Close() error {
	return c.conn.Close()
}

func buildAppLaunchPayload(deviceId string, bundleId string, args []interface{}, env map[string]interface{}) map[string]interface{} {
	platformSpecificOptions := bytes.NewBuffer(nil)
	plistEncoder := plist.NewBinaryEncoder(platformSpecificOptions)
	err := plistEncoder.Encode(map[string]interface{}{})
	if err != nil {
		panic(err)
	}

	return map[string]interface{}{
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
				"platformSpecificOptions":       platformSpecificOptions.Bytes(),
				"standardIOUsesPseudoterminals": true,
				"startStopped":                  true,
				"terminateExisting":             true,
				"user": map[string]interface{}{
					"active": true,
				},
				"workingDirectory": nil,
			},
			"standardIOIdentifiers": map[string]interface{}{},
		},
		"CoreDevice.invocationIdentifier": "62419FC1-5ABF-4D96-BCA8-7A5F6F9A69EE",
	}
}
