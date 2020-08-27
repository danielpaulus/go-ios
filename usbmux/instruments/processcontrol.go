package instruments

import (
	"fmt"

	"github.com/danielpaulus/go-ios/usbmux"
	dtx "github.com/danielpaulus/go-ios/usbmux/dtx_codec"
	"github.com/danielpaulus/go-ios/usbmux/nskeyedarchiver"
	log "github.com/sirupsen/logrus"
)

const channelName = "com.apple.instruments.server.services.processcontrol"

type ProcessControl struct {
	processControlChannel dtx.DtxChannel
}

type ProcessControlDispatcher struct{}

func LaunchApp(bundleID string, device usbmux.DeviceEntry) (uint64, error) {
	conn, _ := dtx.NewDtxConnection(device.DeviceID, device.Properties.SerialNumber, "com.apple.instruments.remoteserver")
	defer conn.Close()
	processControl := NewProcessControl(conn)
	options := map[string]interface{}{}
	options["StartSuspendedKey"] = uint64(0)
	return processControl.StartProcess(bundleID, map[string]interface{}{}, []interface{}{}, options)
}

func (p ProcessControlDispatcher) Dispatch(m dtx.DtxMessage) {
	log.Info(m)
}

func NewProcessControl(dtxConnection *dtx.DtxConnection) ProcessControl {
	processControlChannel := dtxConnection.RequestChannelIdentifier(channelName, ProcessControlDispatcher{})
	return ProcessControl{processControlChannel: processControlChannel}
}

func (p ProcessControl) StartProcess(bundleID string, envVars map[string]interface{}, arguments []interface{}, options map[string]interface{}) (uint64, error) {
	const objcMethodName = "launchSuspendedProcessWithDevicePath:bundleIdentifier:environment:arguments:options:"
	//seems like the path does not matter
	const path = "/private/"

	payload, _ := nskeyedarchiver.ArchiveBin(objcMethodName)
	auxiliary := dtx.NewDtxPrimitiveDictionary()

	auxiliary.AddNsKeyedArchivedObject(path)
	auxiliary.AddNsKeyedArchivedObject(bundleID)
	auxiliary.AddNsKeyedArchivedObject(envVars)
	auxiliary.AddNsKeyedArchivedObject(arguments)
	auxiliary.AddNsKeyedArchivedObject(options)

	log.WithFields(log.Fields{"channel_id": channelName, "bundleID": bundleID}).Info("Launching process")

	msg, err := p.processControlChannel.SendAndAwaitReply(true, dtx.MethodinvocationWithoutExpectedReply, payload, auxiliary)
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
