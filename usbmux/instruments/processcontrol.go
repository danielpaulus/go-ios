package instruments

import (
	"fmt"

	dtx "github.com/danielpaulus/go-ios/usbmux/dtx_codec"
	"github.com/danielpaulus/go-ios/usbmux/nskeyedarchiver"
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

func (p ProcessControl) StartProcess(path string, bundleID string, envVars map[string]interface{}, arguments []interface{}, options map[string]interface{}) (int, error) {
	const objcMethodName = "launchSuspendedProcessWithDevicePath:bundleIdentifier:environment:arguments:options:"

	payload, _ := nskeyedarchiver.ArchiveBin(objcMethodName)
	auxiliary := dtx.NewDtxPrimitiveDictionary()

	arch, err := nskeyedarchiver.ArchiveBin(path)
	dieOnErr(err)
	auxiliary.AddBytes(arch)
	arch, err = nskeyedarchiver.ArchiveBin(bundleID)
	dieOnErr(err)
	auxiliary.AddBytes(arch)
	arch, err = nskeyedarchiver.ArchiveBin(envVars)
	dieOnErr(err)
	auxiliary.AddBytes(arch)
	arch, err = nskeyedarchiver.ArchiveBin(arguments)
	dieOnErr(err)
	auxiliary.AddBytes(arch)
	arch, err = nskeyedarchiver.ArchiveBin(options)
	dieOnErr(err)
	auxiliary.AddBytes(arch)

	log.WithFields(log.Fields{"channel_id": channelName, "bundleID": bundleID}).Info("Launching process")

	msg, err := p.processControlChannel.SendAndAwaitReply(true, dtx.MethodinvocationWithoutExpectedReply, payload, auxiliary)
	if err != nil {
		log.WithFields(log.Fields{"channel_id": channelName, "error": err}).Info("failed starting process")
	}
	if msg.HasError() {
		return -1, fmt.Errorf("Failed starting process: %s", msg.Payload[0])
	}
	return 0, nil
}
func dieOnErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
