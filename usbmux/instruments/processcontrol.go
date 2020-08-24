package instruments

import (
	dtx "github.com/danielpaulus/go-ios/usbmux/dtx_codec"
	log "github.com/sirupsen/logrus"
)

const channelName = "com.apple.instruments.server.services.processcontrol"

type ProcessControl struct {
	processControlChannel dtx.DtxChannel
}

type ProcessControlDispatcher struct{}

func (p ProcessControlDispatcher) Dispatch(m dtx.DtxMessage) {
	log.Info(m)
}

func NewProcessControl(dtxConnection *dtx.DtxConnection) ProcessControl {
	processControlChannel := dtxConnection.RequestChannelIdentifier(channelName, ProcessControlDispatcher{})
	return ProcessControl{processControlChannel: processControlChannel}
}

func startProcess(path string, bundleID string, envVars map[string]string, arguments []string, options map[string]interface{}) {
	const objcMethodName = "launchSuspendedProcessWithDevicePath:bundleIdentifier:environment:arguments:options:"

}
