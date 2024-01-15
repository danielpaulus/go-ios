package appservice

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"path"
	"syscall"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/coredevice"
	"github.com/danielpaulus/go-ios/ios/xpc"
	"github.com/google/uuid"
	"howett.net/plist"
)

type Connection struct {
	conn     *xpc.Connection
	deviceId string
}

const (
	RebootFull      = "full"
	RebootUserspace = "userspace"
)

func New(deviceEntry ios.DeviceEntry) (*Connection, error) {
	xpcConn, err := ios.ConnectToXpcServiceTunnelIface(deviceEntry, "com.apple.coredevice.appservice")
	if err != nil {
		return nil, err
	}

	return &Connection{conn: xpcConn, deviceId: uuid.New().String()}, nil
}

type AppLaunch struct {
	Pid int64
}

type Process struct {
	Pid  int
	Path string
}

func (c *Connection) LaunchApp(deviceId string, bundleId string, args []interface{}, env map[string]interface{}, options map[string]interface{}) (AppLaunch, error) {
	msg := buildAppLaunchPayload(c.deviceId, bundleId, args, env, options)
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

func (c *Connection) Close() error {
	return c.conn.Close()
}

func (c *Connection) ListProcesses() ([]Process, error) {
	req := buildListProcessesPayload(c.deviceId)
	err := c.conn.Send(req, xpc.HeartbeatRequestFlag)
	if err != nil {
		return nil, err
	}
	res, err := c.conn.ReceiveOnServerClientStream()
	if err != nil {
		return nil, err
	}

	if output, ok := res["CoreDevice.output"].(map[string]interface{}); ok {
		if tokens, ok := output["processTokens"].([]interface{}); ok {
			processes := make([]Process, len(tokens), len(tokens))
			for i, t := range tokens {
				var p Process

				if processMap, ok := t.(map[string]interface{}); ok {
					if pid, ok := processMap["processIdentifier"].(int64); ok {
						p.Pid = int(pid)
					} else {
						return nil, fmt.Errorf("could not parse pid (type: %T)", processMap["processIdentifier"])
					}
					if processPath, ok := processMap["executableURL"].(map[string]interface{})["relative"].(string); ok {
						p.Path = processPath
					} else {
						return nil, fmt.Errorf("could not parse process path (type: %T)", processMap["executableURL"])
					}
				} else {
					return nil, errors.New("could not get process info")
				}

				processes[i] = p
			}
			return processes, nil
		}
	}

	return nil, fmt.Errorf("could not parse response")
}

func (c *Connection) KillProcess(pid uint64) error {
	req := buildSendSignalPayload(c.deviceId, pid, syscall.SIGKILL)
	err := c.conn.Send(req, xpc.HeartbeatRequestFlag)
	if err != nil {
		return err
	}
	m, err := c.conn.ReceiveOnServerClientStream()
	if err != nil {
		return err
	}
	err = getError(m)
	if err != nil {
		return err
	}
	return nil
}

// Reboot performs a full reboot of the device
func (c *Connection) Reboot() error {
	return c.RebootWithStyle(RebootFull)
}

func (c *Connection) RebootWithStyle(style string) error {
	err := c.conn.Send(buildRebootPayload(c.deviceId, style))
	if err != nil {
		return err
	}
	m, err := c.conn.ReceiveOnServerClientStream()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		var opErr *net.OpError
		if errors.As(err, &opErr) && opErr.Timeout() {
			return nil
		}
		return err
	}
	return getError(m)
}

func buildAppLaunchPayload(deviceId string, bundleId string, args []interface{}, env map[string]interface{}, options map[string]interface{}) map[string]interface{} {
	platformSpecificOptions := bytes.NewBuffer(nil)
	plistEncoder := plist.NewBinaryEncoder(platformSpecificOptions)
	err := plistEncoder.Encode(options)
	if err != nil {
		panic(err)
	}

	return coredevice.BuildRequest(deviceId, "com.apple.coredevice.feature.launchapplication", map[string]interface{}{
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
	})
}

func buildListProcessesPayload(deviceId string) map[string]interface{} {
	return coredevice.BuildRequest(deviceId, "com.apple.coredevice.feature.listprocesses", nil)
}

func buildRebootPayload(deviceId string, style string) map[string]interface{} {
	return coredevice.BuildRequest(deviceId, "com.apple.coredevice.feature.rebootdevice", map[string]interface{}{
		"rebootStyle": map[string]interface{}{
			style: map[string]interface{}{},
		},
	})
}

func buildSendSignalPayload(deviceId string, pid uint64, signal syscall.Signal) map[string]interface{} {
	return coredevice.BuildRequest(deviceId, "com.apple.coredevice.feature.sendsignaltoprocess", map[string]interface{}{
		"process": map[string]interface{}{
			"processIdentifier": int64(pid),
		},
		"signal": int64(signal),
	})
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

func getError(response map[string]interface{}) error {
	if e, ok := response["CoreDevice.error"].(map[string]interface{}); ok {
		return fmt.Errorf("device returned error: %+v", e)
	}
	return nil
}

func (p Process) ExecutableName() string {
	_, file := path.Split(p.Path)
	return file
}
