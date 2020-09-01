package accessibility

import (
	dtx "github.com/danielpaulus/go-ios/usbmux/dtx_codec"
	log "github.com/sirupsen/logrus"
)

type AccessibilityControl struct {
	channel *dtx.DtxChannel
}

func (a AccessibilityControl) Init() error {
	err := a.notifyPublishedCapabilities()
	if err != nil {
		return err
	}
	//a list of methods we are allowed to call on the device
	deviceCapabilities, err := a.deviceCapabilities()
	if err != nil {
		return err
	}
	log.Info("Device Capabilities:", deviceCapabilities)

	a.deviceAllAuditCaseIDs()
	return nil
}

func (a AccessibilityControl) notifyPublishedCapabilities() error {
	capabs := map[string]interface{}{
		"com.apple.private.DTXBlockCompression": uint64(2),
		"com.apple.private.DTXConnection":       uint64(1),
	}
	return a.channel.MethodCallAsync("_notifyOfPublishedCapabilities:", []interface{}{capabs})

}

func (a AccessibilityControl) deviceCapabilities() ([]string, error) {
	response, err := a.channel.MethodCall("deviceCapabilities", []interface{}{})
	if err != nil {
		return nil, err
	}
	return convertToStringList(response.Payload), nil
}

func (a AccessibilityControl) deviceAllAuditCaseIDs() {
	response, err := a.channel.MethodCall("deviceAllAuditCaseIDs", []interface{}{})
	log.Info(err)

	log.Info(response)
}
func (a AccessibilityControl) deviceApiVersion()                           {}
func (a AccessibilityControl) deviceInspectorCanNavWhileMonitoringEvents() {}
func (a AccessibilityControl) deviceInspectorSupportedEventTypes()         {}
