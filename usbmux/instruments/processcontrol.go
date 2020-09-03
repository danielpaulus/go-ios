package instruments

import (
	"fmt"

	"github.com/danielpaulus/go-ios/usbmux"
	dtx "github.com/danielpaulus/go-ios/usbmux/dtx_codec"
	log "github.com/sirupsen/logrus"
)

const channelName = "com.apple.instruments.server.services.processcontrol"

type ProcessControl struct {
	processControlChannel *dtx.Channel
}

type ProcessControlDispatcher struct{}

func KillApp(pid uint64, device usbmux.DeviceEntry) error {
	conn, _ := dtx.NewConnection(device, "com.apple.instruments.remoteserver")
	//defer conn.Close()
	processControl := NewProcessControl(conn)
	return processControl.KillProcess(pid)
}

func LaunchApp(bundleID string, device usbmux.DeviceEntry) (uint64, error) {
	conn, _ := dtx.NewConnection(device, "com.apple.instruments.remoteserver")
	//defer conn.Close()
	processControl := NewProcessControl(conn)
	options := map[string]interface{}{}
	options["StartSuspendedKey"] = uint64(0)
	return processControl.StartProcess(bundleID, map[string]interface{}{}, []interface{}{}, options)
}

func LaunchAppWithArgs(bundleID string, device usbmux.DeviceEntry, args []interface{}, env map[string]interface{}, opts map[string]interface{}) (uint64, error) {
	conn, _ := dtx.NewConnection(device, "com.apple.instruments.remoteserver")
	//defer conn.Close()
	return NewProcessControl(conn).StartProcess(bundleID, env, args, opts)
}

func (p ProcessControlDispatcher) Dispatch(m dtx.Message) {
	log.Debug(m)
}

func NewProcessControl(dtxConnection *dtx.Connection) ProcessControl {
	processControlChannel := dtxConnection.RequestChannelIdentifier(channelName, ProcessControlDispatcher{})
	return ProcessControl{processControlChannel: processControlChannel}
}

func (p ProcessControl) KillProcess(pid uint64) error {
	_, err := p.processControlChannel.MethodCall("killPid:", pid)
	return err
}

func (p ProcessControl) StartProcess(bundleID string, envVars map[string]interface{}, arguments []interface{}, options map[string]interface{}) (uint64, error) {
	//seems like the path does not matter
	const path = "/private/"

	log.WithFields(log.Fields{"channel_id": channelName, "bundleID": bundleID}).Info("Launching process")

	msg, err := p.processControlChannel.MethodCall(
		"launchSuspendedProcessWithDevicePath:bundleIdentifier:environment:arguments:options:",
		path,
		bundleID,
		envVars,
		arguments,
		options)
	if err != nil {
		log.WithFields(log.Fields{"channel_id": channelName, "error": err}).Info("failed starting process")
	}
	if msg.HasError() {
		return 0, fmt.Errorf("Failed starting process: %s", msg.Payload[0])
	}
	if pid, ok := msg.Payload[0].(uint64); ok {
		log.WithFields(log.Fields{"channel_id": channelName, "pid": pid}).Info("Process started successfully")
		return pid, nil
	}
	return 0, fmt.Errorf("pid returned in payload was not of type uint64 for processcontroll.startprocess, instead: %s", msg.Payload)

}
