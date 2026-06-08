package instruments

import (
	"fmt"
	"maps"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
)

const (
	// deviceProcessControlDomain is the NSError domain the on-device developer
	// tools process-control service reports launch failures under.
	deviceProcessControlDomain = "com.apple.dt.deviceprocesscontrolservice"
	// transientLaunchErrorCode is the "could not launch right now" code that
	// service returns for transient device states (screen locked, a prior
	// instance still terminating, FrontBoard not ready). It clears on retry.
	transientLaunchErrorCode = 2
	// launchMaxAttempts and launchRetryBackoff bound the retry loop. Backoff
	// grows linearly per attempt, so worst case adds ~3s before surfacing a
	// persistent failure.
	launchMaxAttempts  = 4
	launchRetryBackoff = 500 * time.Millisecond
)

// launchRetrySleep is the sleep used between launch retries; a package var so
// tests can stub it out and not actually wait.
var launchRetrySleep = time.Sleep

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
//
// The on-device process-control service intermittently rejects a launch with a
// transient "could not launch right now" error (NSError code 2 in the
// com.apple.dt.deviceprocesscontrolservice domain) when the device is in a
// momentary state that won't accept it — screen locked, a previous instance
// still terminating, FrontBoard not yet ready. That clears on its own, so the
// launch is retried a bounded number of times before the error is surfaced.
func (p ProcessControl) StartProcess(bundleID string, envVars map[string]interface{}, arguments []interface{}, options map[string]interface{}) (uint64, error) {
	// seems like the path does not matter
	const path = "/private/"

	return startProcessWithRetry(bundleID, func() (dtx.Message, error) {
		golog.Info("Launching process", "module", logModule, "channel_id", procControlChannel, "bundleID", bundleID)
		return p.processControlChannel.MethodCall(
			"launchSuspendedProcessWithDevicePath:bundleIdentifier:environment:arguments:options:",
			path,
			bundleID,
			envVars,
			arguments,
			options)
	})
}

// startProcessWithRetry runs launch (a single launchSuspendedProcess call) up to
// launchMaxAttempts times, retrying only when the device reports a transient
// launch failure. It is split out from StartProcess so the retry behaviour can
// be unit-tested without a device by injecting launch.
func startProcessWithRetry(bundleID string, launch func() (dtx.Message, error)) (uint64, error) {
	var lastErr error
	for attempt := 1; attempt <= launchMaxAttempts; attempt++ {
		msg, err := launch()
		pid, perr := parseLaunchReply(bundleID, msg, err)
		if perr == nil {
			golog.Info("Process started successfully", "module", logModule, "channel_id", procControlChannel, "pid", pid, "bundleID", bundleID)
			return pid, nil
		}
		lastErr = perr

		if attempt < launchMaxAttempts && isTransientLaunchError(msg) {
			backoff := time.Duration(attempt) * launchRetryBackoff
			golog.Warn("transient launch failure from device, retrying", "module", logModule, "channel_id", procControlChannel, "bundleID", bundleID, "attempt", attempt, "backoff", backoff.String(), "error", perr)
			launchRetrySleep(backoff)
			continue
		}

		golog.Error("failed starting process", "module", logModule, "channel_id", procControlChannel, "error", perr, "bundleID", bundleID, "attempts", attempt)
		return 0, perr
	}
	return 0, lastErr
}

// parseLaunchReply turns one launch call's (msg, err) into a PID or an error.
func parseLaunchReply(bundleID string, msg dtx.Message, err error) (uint64, error) {
	if err != nil {
		return 0, err
	}
	if msg.HasError() {
		return 0, fmt.Errorf("Failed starting process: %s, msg:%v", bundleID, msg.Payload[0])
	}
	if pid, ok := msg.Payload[0].(uint64); ok {
		return pid, nil
	}
	return 0, fmt.Errorf("pid returned in payload was not of type uint64 for processcontroll.startprocess, instead: %s", msg.Payload)
}

// isTransientLaunchError reports whether msg carries the device's transient
// launch-rejection NSError (code 2, deviceprocesscontrolservice domain), the
// only failure StartProcess retries on. Matching the structured NSError code and
// domain avoids brittle string matching on the localized description.
func isTransientLaunchError(msg dtx.Message) bool {
	if len(msg.Payload) == 0 {
		return false
	}
	nsErr, ok := msg.Payload[0].(nskeyedarchiver.NSError)
	if !ok {
		return false
	}
	return nsErr.Domain == deviceProcessControlDomain && nsErr.ErrorCode == transientLaunchErrorCode
}
