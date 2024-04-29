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
	"github.com/danielpaulus/go-ios/ios/openstdio"
	"github.com/danielpaulus/go-ios/ios/xpc"
	"github.com/google/uuid"
	"howett.net/plist"
)

// Connection represents a connection to the appservice on an iOS device for iOS17+.
// It is used to launch and kill apps and to list processes.
type Connection struct {
	conn     *xpc.Connection
	deviceId string
	device   ios.DeviceEntry
}

const (
	// RebootFull is the style for a full reboot of the device.
	RebootFull = "full"
	// RebootUserspace is the style for a reboot of the userspace of the device.
	RebootUserspace = "userspace"
)

// New creates a new connection to the appservice on the device for iOS17+.
// It returns an error if the connection could not be established.
func New(deviceEntry ios.DeviceEntry) (*Connection, error) {
	xpcConn, err := ios.ConnectToXpcServiceTunnelIface(deviceEntry, "com.apple.coredevice.appservice")
	if err != nil {
		return nil, fmt.Errorf("new: %w", err)
	}

	return &Connection{conn: xpcConn, deviceId: uuid.New().String(), device: deviceEntry}, nil
}

// LaunchedAppWithStdIo is the launched app with a connection to the stdio-socket
type LaunchedAppWithStdIo struct {
	stdIoConnection openstdio.Connection
	Pid             int
}

// Read reads from the stdio socket of the launched app
func (a LaunchedAppWithStdIo) Read(p []byte) (n int, err error) {
	return a.stdIoConnection.Read(p)
}

// Write reads from the stdio socket of the launched app
func (a LaunchedAppWithStdIo) Write(p []byte) (n int, err error) {
	return a.stdIoConnection.Write(p)
}

// Close closes the connection to stdio-socket of the launched app
func (a LaunchedAppWithStdIo) Close() error {
	return a.stdIoConnection.Close()
}

// Process represents a process running on the device for iOS17+.
// It contains the PID and the path of the process.
type Process struct {
	Pid  int
	Path string
}

// LaunchApp launches an app on the device with the given bundleId and arguments for iOS17+.
// On a successful launch it returns the PID of the launched process.
func (c *Connection) LaunchApp(bundleId string, args []interface{}, env map[string]interface{}, options map[string]interface{}, terminateExisting bool) (int, error) {
	pid, err := c.launchApp(bundleId, args, env, options, terminateExisting, map[string]any{})
	if err != nil {
		return 0, fmt.Errorf("LaunchApp: failed to launch app: %w", err)
	}
	return pid, nil
}

// LaunchAppWithStdIo launches an app and connects to the stdio-socket
// the returned value implements the io.ReadWriteCloser interface and needs to be closed when finished using the stdio-socket
func (c *Connection) LaunchAppWithStdIo(bundleId string, args []interface{}, env map[string]interface{}, options map[string]interface{}, terminateExisting bool) (LaunchedAppWithStdIo, error) {
	stdio, err := openstdio.NewOpenStdIoSocket(c.device)
	if err != nil {
		return LaunchedAppWithStdIo{}, fmt.Errorf("LaunchAppWithStdIo: failed to open stdio socket: %w", err)
	}

	// this is also how Xcode handles it. It uses the same socket for stdOut/stdErr/stdIn
	stdIoConfig := map[string]any{
		"standardInput":  stdio.ID,
		"standardOutput": stdio.ID,
		"standardError":  stdio.ID,
	}

	pid, err := c.launchApp(bundleId, args, env, options, terminateExisting, stdIoConfig)
	if err != nil {
		return LaunchedAppWithStdIo{}, fmt.Errorf("LaunchAppWithStdIo: failed to launch app: %w", err)
	}
	return LaunchedAppWithStdIo{
		stdIoConnection: stdio,
		Pid:             pid,
	}, nil
}

func (c *Connection) launchApp(bundleId string, args []interface{}, env map[string]interface{}, options map[string]interface{}, terminateExisting bool, stdio map[string]any) (int, error) {
	msg := buildAppLaunchPayload(c.deviceId, bundleId, args, env, options, terminateExisting, stdio)
	err := c.conn.Send(msg, xpc.HeartbeatRequestFlag)
	if err != nil {
		return 0, fmt.Errorf("launchApp: failed to send launch-app request: %w", err)
	}
	m, err := c.conn.ReceiveOnServerClientStream()
	if err != nil {
		return 0, fmt.Errorf("launchApp: failed to read response: %w", err)
	}
	pid, err := pidFromResponse(m)
	if err != nil {
		return 0, fmt.Errorf("launchApp: failed to get PID: %w", err)
	}
	return int(pid), nil
}

// Close closes the connection to the appservice
func (c *Connection) Close() error {
	return c.conn.Close()
}

// ListProcesses returns a list of processes with their PID and executable path running on the device for iOS17+.
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

	output, ok := res["CoreDevice.output"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("listProcesses output")
	}
	tokens, ok := output["processTokens"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("listProcesses processTokens")
	}
	processes := make([]Process, len(tokens))
	tokensTyped, err := ios.GenericSliceToType[map[string]interface{}](tokens)
	if err != nil {
		return nil, fmt.Errorf("listProcesses: %w", err)
	}
	for i, processMap := range tokensTyped {
		var p Process
		pid, ok := processMap["processIdentifier"].(int64)
		if !ok {
			return nil, fmt.Errorf("listProcesses processIdentifier")
		}
		processPathMap, ok := processMap["executableURL"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("listProcesses executableURL")
		}
		processPath, ok := processPathMap["relative"].(string)
		if !ok {
			return nil, fmt.Errorf("listProcesses relative")
		}

		p.Pid = int(pid)
		p.Path = processPath

		processes[i] = p
	}

	return processes, nil
}

// KillProcess kills the process with the given PID for iOS17+.
func (c *Connection) KillProcess(pid int) error {
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

func buildAppLaunchPayload(deviceId string, bundleId string, args []interface{}, env map[string]interface{}, options map[string]interface{}, terminateExisting bool, stdIo map[string]any) map[string]interface{} {
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
		"standardIOIdentifiers": stdIo,
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

func buildSendSignalPayload(deviceId string, pid int, signal syscall.Signal) map[string]interface{} {
	return coredevice.BuildRequest(deviceId, "com.apple.coredevice.feature.sendsignaltoprocess", map[string]interface{}{
		"process": map[string]interface{}{
			"processIdentifier": int64(pid),
		},
		"signal": int64(signal),
	})
}

func pidFromResponse(response map[string]interface{}) (int64, error) {
	output, ok := response["CoreDevice.output"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("pidFromResponse: could not get pid from response")
	}
	processToken, ok := output["processToken"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("pidFromResponse: could not get processToken from response")
	}
	pid, ok := processToken["processIdentifier"].(int64)
	if !ok {
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
