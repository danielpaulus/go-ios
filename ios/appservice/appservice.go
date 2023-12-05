package appservice

import (
	"bytes"
	"errors"
	"strings"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/xpc"
	"github.com/google/uuid"

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

func (c *Connection) LaunchApp(deviceId string, bundleId string, args []interface{}, env map[string]interface{}) (uint64, error) {
	msg := buildAppLaunchPayload(deviceId, bundleId, args, env)
	result, err := c.conn.SendReceive(msg)
	if err != nil {
		return 0, err
	}

	output, exists := result["CoreDevice.output"].(map[string]interface{})
	if !exists {
		return 0, errors.New("Process not launched")
	}
	processToken, exists := output["processToken"].(map[string]interface{})
	if !exists {
		return 0, errors.New("Process not launched")
	}
	pid, exists := processToken["processIdentifier"].(int64)
	if !exists {
		return 0, errors.New("Process not launched")
	}

	return uint64(pid), nil
}

func (c *Connection) ListProcesses(deviceId string) error {
	msg := buildCoreDevicePayload(deviceId, "com.apple.coredevice.feature.listprocesses", map[string]interface{}{})
	lol, err := c.conn.SendReceive(msg)
	lol = lol
	return err
}

func (c *Connection) Close() error {
	return c.conn.Close()
}

func buildAppLaunchPayload(deviceId string, bundleId string, args []interface{}, env map[string]interface{}) map[string]interface{} {
	platformSpecificOptions := bytes.NewBuffer(nil)
	plistEncoder := plist.NewBinaryEncoder(platformSpecificOptions)
	err := plistEncoder.Encode(map[string]interface{}{
		"ActivateSuspended": uint64(1),
	})
	if err != nil {
		panic(err)
	}

	return buildCoreDevicePayload(deviceId, "com.apple.coredevice.feature.launchapplication", map[string]interface{}{
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
	})
}

func buildCoreDevicePayload(deviceId string, feature string, input map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"CoreDevice.CoreDeviceDDIProtocolVersion": int64(0),
		"CoreDevice.action":                       map[string]interface{}{},
		"CoreDevice.coreDeviceVersion": map[string]interface{}{
			"components":              []interface{}{uint64(0x15c), uint64(0x1), uint64(0x0), uint64(0x0), uint64(0x0)},
			"originalComponentsCount": int64(2),
			"stringValue":             "348.1",
		},
		"CoreDevice.deviceIdentifier":     deviceId,
		"CoreDevice.featureIdentifier":    feature,
		"CoreDevice.input":                input,
		"CoreDevice.invocationIdentifier": strings.ToUpper(uuid.New().String()),
	}
}
