package accessibility

import (
	dtx "github.com/danielpaulus/go-ios/usbmux/dtx_codec"
	"github.com/danielpaulus/go-ios/usbmux/nskeyedarchiver"
	log "github.com/sirupsen/logrus"
)

type AccessibilityControl struct {
	channel *dtx.DtxChannel
}

func (a AccessibilityControl) readhostAppStateChanged() {
	for {
		msg := a.channel.ReceiveMethodCall("hostAppStateChanged:")
		stateChange, err := nskeyedarchiver.Unarchive(msg.Auxiliary.GetArguments()[0].([]byte))
		if err != nil {
			log.Fatal(err)
		}
		value := stateChange[0]
		log.Infof("hostAppStateChanged:%s", value)
	}
}

func (a AccessibilityControl) readhostInspectorNotificationReceived() {
	for {
		msg := a.channel.ReceiveMethodCall("hostInspectorNotificationReceived:")
		notification, err := nskeyedarchiver.Unarchive(msg.Auxiliary.GetArguments()[0].([]byte))
		if err != nil {
			log.Fatal(err)
		}
		value := notification[0].(map[string]interface{})["Value"]
		log.Infof("hostInspectorNotificationReceived:%s", value)
	}
}

func (a AccessibilityControl) Init() error {
	a.channel.RegisterMethodForRemote("hostInspectorCurrentElementChanged:")
	a.channel.RegisterMethodForRemote("hostInspectorMonitoredEventTypeChanged:")
	a.channel.RegisterMethodForRemote("hostAppStateChanged:")
	a.channel.RegisterMethodForRemote("hostInspectorNotificationReceived:")
	go a.readhostAppStateChanged()
	go a.readhostInspectorNotificationReceived()

	/*err := a.notifyPublishedCapabilities()
	if err != nil {
		return err
	}*/
	//a list of methods we are allowed to call on the device
	deviceCapabilities, err := a.deviceCapabilities()
	if err != nil {
		return err
	}

	log.Info("Device Capabilities:", deviceCapabilities)
	apiVersion, err := a.deviceAPIVersion()
	if err != nil {
		return err
	}
	log.Info("Api version:", apiVersion)

	auditCaseIds, err := a.deviceAllAuditCaseIDs()
	if err != nil {
		return err
	}
	log.Info("AuditCaseIDs", auditCaseIds)

	deviceInspectorSupportedEventTypes, err := a.deviceInspectorSupportedEventTypes()
	if err != nil {
		return err
	}
	log.Info("deviceInspectorSupportedEventTypes:", deviceInspectorSupportedEventTypes)

	canNav, err := a.deviceInspectorCanNavWhileMonitoringEvents()
	if err != nil {
		return err
	}
	log.Info("deviceInspectorCanNavWhileMonitoringEvents:", canNav)

	err = a.deviceSetAppMonitoringEnabled(true)
	if err != nil {
		return err
	}

	for _, v := range auditCaseIds {
		name, err := a.deviceHumanReadableDescriptionForAuditCaseID(v)
		if err != nil {
			return err
		}
		log.Infof("%s -- %s", v, name)
	}
	return nil
}

func (a AccessibilityControl) EnableSelectionMode() {
	a.deviceInspectorSetMonitoredEventType(2)
	a.deviceInspectorShowVisuals(true)
	a.awaitHostInspectorMonitoredEventTypeChanged()
}

//should receive AX notifications now
func (a AccessibilityControl) SwitchToDevice() {
	a.TurnOff()
	resp, _ := a.deviceAccessibilitySettings()
	log.Info("AX Settings received:", resp)
	a.deviceInspectorShowIgnoredElements(false)
	a.deviceSetAuditTargetPid(0)
	a.deviceInspectorFocusOnElement()
	a.awaitHostInspectorCurrentElementChanged()
	a.deviceInspectorPreviewOnElement()
	a.deviceHighlightIssue()

}

func (a AccessibilityControl) TurnOff() {
	a.deviceInspectorSetMonitoredEventType(0)
	a.awaitHostInspectorMonitoredEventTypeChanged()
	a.deviceInspectorFocusOnElement()
	a.awaitHostInspectorCurrentElementChanged()
	a.deviceInspectorPreviewOnElement()
	a.deviceHighlightIssue()
	a.deviceInspectorShowVisuals(false)
}

func (a AccessibilityControl) GetElement() {
	log.Info("changing")
	a.deviceInspectorMoveWithOptions()
	//a.deviceInspectorMoveWithOptions()

	resp := a.awaitHostInspectorCurrentElementChanged()
	log.Info("item changed", resp)
}

func (a AccessibilityControl) awaitHostInspectorCurrentElementChanged() map[string]interface{} {
	msg := a.channel.ReceiveMethodCall("hostInspectorCurrentElementChanged:")
	log.Info("received hostInspectorCurrentElementChanged")
	result, err := nskeyedarchiver.Unarchive(msg.Auxiliary.GetArguments()[0].([]byte))
	if err != nil {
		log.Fatalf("Failed unarchiving: %s this is a bug and should not happen", err)
	}
	return result[0].(map[string]interface{})
}

func (a AccessibilityControl) awaitHostInspectorMonitoredEventTypeChanged() {
	msg := a.channel.ReceiveMethodCall("hostInspectorMonitoredEventTypeChanged:")
	n, _ := nskeyedarchiver.Unarchive(msg.Auxiliary.GetArguments()[0].([]byte))
	log.Infof("hostInspectorMonitoredEventTypeChanged: was set to %d by the device", n[0])
}

func (a AccessibilityControl) deviceInspectorMoveWithOptions() {
	method := "deviceInspectorMoveWithOptions:"
	options := nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
		"ObjectType": "passthrough",
		"Value": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
			"allowNonAX":        nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{"ObjectType": "passthrough", "Value": "false"}),
			"direction":         nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{"ObjectType": "passthrough", "Value": "4"}),
			"includeContainers": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{"ObjectType": "passthrough", "Value": "true"}),
		}),
	})
	a.channel.MethodCallAsync(method, []interface{}{options})

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

func (a AccessibilityControl) deviceAllAuditCaseIDs() ([]string, error) {
	response, err := a.channel.MethodCall("deviceAllAuditCaseIDs", []interface{}{})
	if err != nil {
		return nil, err
	}
	return convertToStringList(response.Payload), nil
}

func (a AccessibilityControl) deviceAccessibilitySettings() (map[string]interface{}, error) {
	response, err := a.channel.MethodCall("deviceAccessibilitySettings", []interface{}{})
	if err != nil {
		return nil, err
	}
	return response.Payload[0].(map[string]interface{}), nil
}

func (a AccessibilityControl) deviceInspectorSupportedEventTypes() (uint64, error) {
	response, err := a.channel.MethodCall("deviceInspectorSupportedEventTypes", []interface{}{})
	if err != nil {
		return 0, err
	}
	return response.Payload[0].(uint64), nil
}

func (a AccessibilityControl) deviceAPIVersion() (uint64, error) {
	response, err := a.channel.MethodCall("deviceApiVersion", []interface{}{})
	if err != nil {
		return 0, err
	}
	return response.Payload[0].(uint64), nil
}
func (a AccessibilityControl) deviceInspectorCanNavWhileMonitoringEvents() (bool, error) {
	response, err := a.channel.MethodCall("deviceInspectorCanNavWhileMonitoringEvents", []interface{}{})
	if err != nil {
		return false, err
	}
	return response.Payload[0].(bool), nil
}

func (a AccessibilityControl) deviceSetAppMonitoringEnabled(val bool) error {
	err := a.channel.MethodCallAsync("deviceSetAppMonitoringEnabled:", []interface{}{val})
	if err != nil {
		return err
	}
	return nil
}

func (a AccessibilityControl) deviceHumanReadableDescriptionForAuditCaseID(auditCaseID string) (string, error) {
	response, err := a.channel.MethodCall("deviceHumanReadableDescriptionForAuditCaseID:", []interface{}{auditCaseID})
	if err != nil {
		return "", err
	}
	return response.Payload[0].(string), nil
}

func (a AccessibilityControl) deviceInspectorShowIgnoredElements(val bool) error {
	return a.channel.MethodCallAsync("deviceInspectorShowIgnoredElements:", []interface{}{val})
}
func (a AccessibilityControl) deviceSetAuditTargetPid(pid uint64) error {
	return a.channel.MethodCallAsync("deviceSetAuditTargetPid:", []interface{}{pid})
}
func (a AccessibilityControl) deviceInspectorFocusOnElement() error {
	return a.channel.MethodCallAsync("deviceInspectorFocusOnElement:", []interface{}{nskeyedarchiver.NewNSNull()})
}
func (a AccessibilityControl) deviceInspectorPreviewOnElement() error {
	return a.channel.MethodCallAsync("deviceInspectorPreviewOnElement:", []interface{}{nskeyedarchiver.NewNSNull()})
}
func (a AccessibilityControl) deviceHighlightIssue() error {
	return a.channel.MethodCallAsync("deviceHighlightIssue:", []interface{}{map[string]interface{}{}})
}
func (a AccessibilityControl) deviceInspectorSetMonitoredEventType(eventtype uint64) error {
	return a.channel.MethodCallAsync("deviceInspectorSetMonitoredEventType:", []interface{}{eventtype})
}
func (a AccessibilityControl) deviceInspectorShowVisuals(val bool) error {
	return a.channel.MethodCallAsync("deviceInspectorShowVisuals:", []interface{}{val})
}
