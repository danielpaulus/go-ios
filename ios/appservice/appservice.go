package appservice

import (
	"bytes"
	"fmt"
	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/xpc"
	"github.com/google/uuid"
	"howett.net/plist"
)

type Connection struct {
	conn *xpc.Connection
}

func New(deviceEntry ios.DeviceEntry) (*Connection, error) {
	xpcConn, err := ios.ConnectToServiceTunnelIface(deviceEntry, "com.apple.coredevice.appservice")
	if err != nil {
		return nil, err
	}

	return &Connection{conn: xpcConn}, nil
}

func (c *Connection) LaunchApp(deviceId string, bundleId string, args []interface{}, env map[string]interface{}) (int64, error) {
	msg := buildAppLaunchPayload(deviceId, bundleId, args, env)
	err := c.conn.Send(msg, xpc.HeartbeatRequestFlag)
	m, err := c.conn.Receive()
	if err != nil {
		return 0, err
	}
	return pidFromResponse(m)
}

func (c *Connection) Close() error {
	return c.conn.Close()
}

func buildAppLaunchPayload(deviceId string, bundleId string, args []interface{}, env map[string]interface{}) map[string]interface{} {
	u := uuid.New()
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
				"startStopped":                  false,
				"terminateExisting":             false,
				"user": map[string]interface{}{
					"active": true,
				},
				"workingDirectory": nil,
			},
			"standardIOIdentifiers": map[string]interface{}{},
		},
		"CoreDevice.invocationIdentifier": u.String(),
	}
}

func pidFromResponse(response map[string]interface{}) (int64, error) {
	if output, ok := response["CoreDevice.output"].(map[string]interface{}); ok {
		if processToken, ok := output["processToken"].(map[string]interface{}); ok {
			if pid, ok := processToken["processIdentifier"].(int64); ok {
				return pid, nil
			}
		}
	}
	return 0, fmt.Errorf("could not get pid from response")
}
