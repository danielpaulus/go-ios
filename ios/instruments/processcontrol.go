package instruments

import (
	"fmt"
	"maps"

	"github.com/danielpaulus/go-ios/ios"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	log "github.com/sirupsen/logrus"
)

type ProcessControl struct {
	processControlChannel *dtx.Channel
	conn                  *dtx.Connection
}

// LaunchApp launches the app with the given bundleID on the given device.LaunchApp
// Use LaunchAppWithArgs for passing arguments and envVars. It returns the PID of the created app process.
func (p *ProcessControl) LaunchApp(bundleID string, my_opts map[string]any) (uint64, error) {
	opts := map[string]interface{}{
		"StartSuspendedKey": uint64(0),
		"KillExisting":      uint64(0),
	}
	maps.Copy(opts, my_opts)
	// Xcode sends all these, no idea if we need them for sth. later.
	// "CA_ASSERT_MAIN_THREAD_TRANSACTIONS": "0", "CA_DEBUG_TRANSACTIONS": "0", "LLVM_PROFILE_FILE": "/dev/null", "METAL_DEBUG_ERROR_MODE": "0", "METAL_DEVICE_WRAPPER_TYPE": "1",
	// "OS_ACTIVITY_DT_MODE": "YES", "SQLITE_ENABLE_THREAD_ASSERTIONS": "1", "__XPC_LLVM_PROFILE_FILE": "/dev/null"
	// NSUnbufferedIO seems to make the app send its logs via instruments using the outputReceived:fromProcess:atTime: selector
	// We'll supply per default to get logs
	env := map[string]interface{}{"NSUnbufferedIO": "YES"}
	return p.StartProcess(bundleID, env, []interface{}{}, opts)
}

// LaunchApp launches the app with the given bundleID on the given device.LaunchApp
// It returns the PID of the created app process.
func (p *ProcessControl) LaunchAppWithArgs(bundleID string, my_args []interface{}, my_env map[string]any, my_opts map[string]any) (uint64, error) {
	opts := map[string]interface{}{
		"StartSuspendedKey": uint64(0),
		"KillExisting":      uint64(0),
	}
	maps.Copy(opts, my_opts)
	// Xcode sends all these, no idea if we need them for sth. later.
	// "CA_ASSERT_MAIN_THREAD_TRANSACTIONS": "0", "CA_DEBUG_TRANSACTIONS": "0", "LLVM_PROFILE_FILE": "/dev/null", "METAL_DEBUG_ERROR_MODE": "0", "METAL_DEVICE_WRAPPER_TYPE": "1",
	// "OS_ACTIVITY_DT_MODE": "YES", "SQLITE_ENABLE_THREAD_ASSERTIONS": "1", "__XPC_LLVM_PROFILE_FILE": "/dev/null"
	// NSUnbufferedIO seems to make the app send its logs via instruments using the outputReceived:fromProcess:atTime: selector
	// We'll supply per default to get logs
	env := map[string]interface{}{"NSUnbufferedIO": "YES"}
	maps.Copy(env, my_env)
	return p.StartProcess(bundleID, env, my_args, opts)
}

func (p *ProcessControl) Close() error {
	return p.conn.Close()
}

func NewProcessControl(device ios.DeviceEntry) (*ProcessControl, error) {
	dtxConn, err := connectInstruments(device)
	if err != nil {
		return nil, err
	}
	processControlChannel := dtxConn.RequestChannelIdentifier(procControlChannel, loggingDispatcher{dtxConn})
	return &ProcessControl{processControlChannel: processControlChannel, conn: dtxConn}, nil
}

// DisableMemoryLimit disables the memory limit of a process.
func (p ProcessControl) DisableMemoryLimit(pid uint64) (bool, error) {
	aux := dtx.NewPrimitiveDictionary()
	aux.AddInt32(int(pid))
	msg, err := p.processControlChannel.MethodCallWithAuxiliary("requestDisableMemoryLimitsForPid:", aux)
	if err != nil {
		return false, err
	}
	if disabled, ok := msg.Payload[0].(bool); ok {
		return disabled, nil
	}
	return false, fmt.Errorf("expected int 0 or 1 as payload of msg: %v", msg)
}

// KillProcess kills the process on the device.
func (p ProcessControl) KillProcess(pid uint64) error {
	_, err := p.processControlChannel.MethodCall("killPid:", pid)
	return err
}

// StartProcess launches an app on the device using the bundleID and optional envvars, arguments and options. It returns the PID.
func (p ProcessControl) StartProcess(bundleID string, envVars map[string]interface{}, arguments []interface{}, options map[string]interface{}) (uint64, error) {
	// seems like the path does not matter
	const path = "/private/"

	log.WithFields(log.Fields{"channel_id": procControlChannel, "bundleID": bundleID}).Info("Launching process")

	msg, err := p.processControlChannel.MethodCall(
		"launchSuspendedProcessWithDevicePath:bundleIdentifier:environment:arguments:options:",
		path,
		bundleID,
		envVars,
		arguments,
		options)
	if err != nil {
		log.WithFields(log.Fields{"channel_id": procControlChannel, "error": err}).Errorln("failed starting process: ", bundleID)
		return 0, err
	}
	if msg.HasError() {
		return 0, fmt.Errorf("Failed starting process: %s, msg:%v", bundleID, msg.Payload[0])
	}
	if pid, ok := msg.Payload[0].(uint64); ok {
		log.WithFields(log.Fields{"channel_id": procControlChannel, "pid": pid}).Info("Process started successfully")
		return pid, nil
	}
	return 0, fmt.Errorf("pid returned in payload was not of type uint64 for processcontroll.startprocess, instead: %s", msg.Payload)
}
