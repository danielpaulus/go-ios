package instruments

import (
	"fmt"

	ios "github.com/danielpaulus/go-ios/usbmux"
	dtx "github.com/danielpaulus/go-ios/usbmux/dtx_codec"
	log "github.com/sirupsen/logrus"
)

const serviceName string = "com.apple.instruments.remoteserver"
const processControlChannelName = "com.apple.instruments.server.services.processcontrol"

type processControl struct {
	processControlChannel *dtx.Channel
}

type processControlDispatcher struct{}

//KillApp kills an app with the given PID on the given device.
func KillApp(pid uint64, device ios.DeviceEntry) error {
	conn, _ := dtx.NewConnection(device, serviceName)
	//defer conn.Close()
	processControl := newProcessControl(conn)
	return processControl.KillProcess(pid)
}

//LaunchApp launches the app with the given bundleID on the given device.LaunchApp
//Use LaunchAppWithArgs for passing arguments and envVars. It returns the PID of the created app process.
func LaunchApp(bundleID string, device ios.DeviceEntry) (uint64, error) {
	conn, _ := dtx.NewConnection(device, serviceName)
	//defer conn.Close()const
	processControl := newProcessControl(conn)
	options := map[string]interface{}{}
	options["StartSuspendedKey"] = uint64(0)
	return processControl.StartProcess(bundleID, map[string]interface{}{}, []interface{}{}, options)
}

//LaunchAppWithArgs same as LaunchApp but passes arguments, envVars and options.
func LaunchAppWithArgs(bundleID string, device ios.DeviceEntry, args []interface{}, env map[string]interface{}, opts map[string]interface{}) (uint64, error) {
	conn, _ := dtx.NewConnection(device, serviceName)
	//defer conn.Close()
	return newProcessControl(conn).StartProcess(bundleID, env, args, opts)
}

func (p processControlDispatcher) Dispatch(m dtx.Message) {
	log.Debug(m)
}

func newProcessControl(dtxConnection *dtx.Connection) processControl {
	processControlChannel := dtxConnection.RequestChannelIdentifier(processControlChannelName, processControlDispatcher{})
	return processControl{processControlChannel: processControlChannel}
}

//KillProcess kills the process on the device.
func (p processControl) KillProcess(pid uint64) error {
	_, err := p.processControlChannel.MethodCall("killPid:", pid)
	return err
}

//StartProcess launches an app on the device using the bundleID and optional envvars, arguments and options. It returns the PID.
func (p processControl) StartProcess(bundleID string, envVars map[string]interface{}, arguments []interface{}, options map[string]interface{}) (uint64, error) {
	//seems like the path does not matter
	const path = "/private/"

	log.WithFields(log.Fields{"channel_id": processControlChannelName, "bundleID": bundleID}).Info("Launching process")

	msg, err := p.processControlChannel.MethodCall(
		"launchSuspendedProcessWithDevicePath:bundleIdentifier:environment:arguments:options:",
		path,
		bundleID,
		envVars,
		arguments,
		options)
	if err != nil {
		log.WithFields(log.Fields{"channel_id": processControlChannelName, "error": err}).Info("failed starting process")
	}
	if msg.HasError() {
		return 0, fmt.Errorf("Failed starting process: %s", msg.Payload[0])
	}
	if pid, ok := msg.Payload[0].(uint64); ok {
		log.WithFields(log.Fields{"channel_id": processControlChannelName, "pid": pid}).Info("Process started successfully")
		return pid, nil
	}
	return 0, fmt.Errorf("pid returned in payload was not of type uint64 for processcontroll.startprocess, instead: %s", msg.Payload)

}
