// Package appservice provides functions to Launch and Kill apps on an iOS devices for iOS17+.
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
		return nil, fmt.Errorf("new: %w", err)
	}

	return &Connection{conn: xpcConn, deviceId: uuid.New().String()}, nil
}

type AppLaunch struct {
	Pid int64
}

type Process struct {
	Pid  uint64
	Path string
}

// LaunchApp launches an app on the device with the given bundleId and arguments for iOS17+.
func (c *Connection) LaunchApp(bundleId string, args []interface{}, env map[string]interface{}, options map[string]interface{}, terminateExisting bool) (AppLaunch, error) {
	msg := buildAppLaunchPayload(c.deviceId, bundleId, args, env, options, terminateExisting)
	err := c.conn.Send(msg, xpc.HeartbeatRequestFlag)
	if err != nil {
		return AppLaunch{}, fmt.Errorf("launchApp: %w", err)
	}
	m, err := c.conn.ReceiveOnServerClientStream()
	if err != nil {
		return AppLaunch{}, fmt.Errorf("launchApp2: %w", err)
	}
	pid, err := pidFromResponse(m)
	if err != nil {
		return AppLaunch{}, fmt.Errorf("launchApp3: %w", err)
	}
	return AppLaunch{Pid: pid}, nil
}

// Close closes the connection to the appservice
func (c *Connection) Close() error {
	return c.conn.Close()
}

// ListProcesses returns a list of processes with PID and Path running on the device for iOS17+.
func (c *Connection) ListProcesses() ([]Process, error) {
	req := buildListProcessesPayload(c.deviceId)
	err := c.conn.Send(req, xpc.HeartbeatRequestFlag)
	if err != nil {
		return nil, fmt.Errorf("listProcesses send: %w", err)
	}
	res, err := c.conn.ReceiveOnServerClientStream()
	if err != nil {
		return nil, fmt.Errorf("listProcesses receive: %w", err)
	}

	output, err := ios.ValueFromMap[map[string]interface{}](res, "CoreDevice.output")
	if err != nil {
		return nil, fmt.Errorf("listProcesses: %w", err)
	}
	tokens, err := ios.ValueFromMap[[]interface{}](output, "processTokens")
	if err != nil {
		return nil, fmt.Errorf("listProcesses: %w", err)
	}
	processes := make([]Process, len(tokens))
	tokensTyped, err := ios.GenericSliceToType[map[string]interface{}](tokens)
	if err != nil {
		return nil, fmt.Errorf("listProcesses: %w", err)
	}
	for i, processMap := range tokensTyped {
		var p Process
		pid, err := ios.ValueFromMap[int64](processMap, "processIdentifier")
		if err != nil {
			return nil, fmt.Errorf("listProcesses: %w", err)
		}
		processPathMap, err := ios.ValueFromMap[map[string]interface{}](processMap, "executableURL")
		if err != nil {
			return nil, fmt.Errorf("listProcesses: %w", err)
		}
		processPath, err := ios.ValueFromMap[string](processPathMap, "relative")
		if err != nil {
			return nil, fmt.Errorf("listProcesses: %w", err)
		}

		p.Pid = uint64(pid)
		p.Path = processPath

		processes[i] = p
	}

	return processes, nil
}

// KillProcess kills the process with the given PID for iOS17+.
func (c *Connection) KillProcess(pid uint64) error {
	req := buildSendSignalPayload(c.deviceId, pid, syscall.SIGKILL)
	err := c.conn.Send(req, xpc.HeartbeatRequestFlag)
	if err != nil {
		return fmt.Errorf("killProcess send: %w", err)
	}
	m, err := c.conn.ReceiveOnServerClientStream()
	if err != nil {
		return fmt.Errorf("killProcess receive: %w", err)
	}
	err = getError(m)
	if err != nil {
		return fmt.Errorf("killProcess: %w", err)
	}
	return nil
}

// Reboot performs a full reboot of the device for iOS17+.
// Just calls RebootWithStyle with RebootFull.
func (c *Connection) Reboot() error {
	return c.RebootWithStyle(RebootFull)
}

// RebootWithStyle performs a reboot of the device with the given style for iOS17+. For style use RebootFull or RebootUserSpace.
func (c *Connection) RebootWithStyle(style string) error {
	err := c.conn.Send(buildRebootPayload(c.deviceId, style))
	if err != nil {
		return fmt.Errorf("reboot send: %w", err)
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
		return fmt.Errorf("reboot receive: %w", err)
	}
	err = getError(m)
	if err != nil {
		return fmt.Errorf("reboot: %w", err)
	}
	return nil
}

// ExecutableName returns the executable name for a process by removing the path.
func (p Process) ExecutableName() string {
	_, file := path.Split(p.Path)
	return file
}

func buildAppLaunchPayload(deviceId string, bundleId string, args []interface{}, env map[string]interface{}, options map[string]interface{}, terminateExisting bool) map[string]interface{} {
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
			"terminateExisting":             terminateExisting,
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
	output, err := ios.ValueFromMap[map[string]interface{}](response, "CoreDevice.output")
	if err != nil {
		return 0, fmt.Errorf("pidFromResponse: could not get pid from response")
	}
	processToken, err := ios.ValueFromMap[map[string]interface{}](output, "processToken")
	if err != nil {
		return 0, fmt.Errorf("pidFromResponse: could not get processToken from response")
	}
	pid, err := ios.ValueFromMap[int64](processToken, "processIdentifier")
	if err != nil {
		return 0, fmt.Errorf("pidFromResponse: could not get pid from processToken")
	}
	return pid, nil
}

func getError(response map[string]interface{}) error {
	if e, ok := response["CoreDevice.error"].(map[string]interface{}); ok {
		return fmt.Errorf("device returned error: %+v", e)
	}
	return nil
}
