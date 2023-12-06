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

type AppLaunch struct {
	Pid int64
}

func (c *Connection) LaunchApp(deviceId string, bundleId string, args []interface{}, env map[string]interface{}) (AppLaunch, error) {
	msg := buildAppLaunchPayload(deviceId, bundleId, args, env)
	err := c.conn.Send(msg, xpc.HeartbeatRequestFlag)
	m, err := c.conn.ReceiveOnServerClientStream()
	if err != nil {
		return AppLaunch{}, err
	}
	pid, err := pidFromResponse(m)
	if err != nil {
		return AppLaunch{}, err
	}
	return AppLaunch{Pid: pid}, nil
}

func (c *Connection) StartProcess(bundleID string, envVars map[string]interface{}, arguments []interface{}, options map[string]interface{}) (uint64, error) {
	l, err := c.LaunchApp("E66A4DED-A888-495F-A701-1C478F94DC8B", bundleID, arguments, envVars)
	if err != nil {
		return 0, err
	}
	return uint64(l.Pid), nil
}

func (c *Connection) Close() error {
	return c.conn.Close()
}

func (c *Connection) KillProcess(pid uint64) error {
	//TODO implement me
	panic("implement me")
}

func buildAppLaunchPayload(deviceId string, bundleId string, args []interface{}, env map[string]interface{}) map[string]interface{} {
	u := uuid.New()
	platformSpecificOptions := bytes.NewBuffer(nil)
	plistEncoder := plist.NewBinaryEncoder(platformSpecificOptions)
	err := plistEncoder.Encode(map[string]interface{}{
		"__ActivateSuspende": 1,
	})
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
				"terminateExisting":             true,
				"user": map[string]interface{}{
					"active": true,
				},
				"workingDirectory": nil,
			},
			//"installationResult": map[string]interface{}{
			//	"_persistentIdentifier":  "AAAAAEwGAAAIAAAAPb97tuaXS8yiRGRoMPN1U0wGAAAAAAAA",
			//	"applicationBundleId":    "com.saucelabs.TestGridWithInjectorUITests.xctrunner",
			//	"databaseSequenceNumber": uint64(1612),
			//	"databaseUUID":           "3dbf7bb6-e697-4bcc-a244-646830f37553",
			//	"installationURL": map[string]interface{}{
			//		"relative": "file:///private/var/containers/Bundle/Application/17B6955A-A4C0-47DC-BD3A-C3B39E1922E9/TestGridWithInjectorUITests-Runner.app/",
			//	},
			//},
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
