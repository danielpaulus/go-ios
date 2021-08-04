package instruments

import (
	"fmt"

	"github.com/danielpaulus/go-ios/ios"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	log "github.com/sirupsen/logrus"
)

const processControlChannelName = "com.apple.instruments.server.services.processcontrol"

type ProcessControl struct {
	processControlChannel *dtx.Channel
	conn                  *dtx.Connection
}

//LaunchApp launches the app with the given bundleID on the given device.LaunchApp
//Use LaunchAppWithArgs for passing arguments and envVars. It returns the PID of the created app process.
func (p *ProcessControl) LaunchApp(bundleID string) (uint64, error) {
	options := map[string]interface{}{}
	options["StartSuspendedKey"] = uint64(0)
	return p.StartProcess(bundleID, map[string]interface{}{}, []interface{}{}, options)
}

func (p *ProcessControl) Close() {
	p.conn.Close()
}

func NewProcessControl(device ios.DeviceEntry) (*ProcessControl, error) {
	dtxConn, err := connectInstruments(device)
	if err != nil {
		return nil, err
	}
	processControlChannel := dtxConn.RequestChannelIdentifier(processControlChannelName, loggingDispatcher{dtxConn})
	return &ProcessControl{processControlChannel: processControlChannel, conn: dtxConn}, nil
}

//KillProcess kills the process on the device.
func (p ProcessControl) KillProcess(pid uint64) error {
	_, err := p.processControlChannel.MethodCall("killPid:", pid)
	return err
}

//StartProcess launches an app on the device using the bundleID and optional envvars, arguments and options. It returns the PID.
func (p ProcessControl) StartProcess(bundleID string, envVars map[string]interface{}, arguments []interface{}, options map[string]interface{}) (uint64, error) {
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
