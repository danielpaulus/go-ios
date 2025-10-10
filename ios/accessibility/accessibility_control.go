package accessibility

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	log "github.com/sirupsen/logrus"
)

// ControlInterface provides a simple interface to controlling the AX service on the device
// It only needs the global dtx channel as all AX methods are invoked on it.
type ControlInterface struct {
	channel *dtx.Channel
}

// Direction represents navigation direction values used by AX service
type MoveDirection int32

const (
	DirectionPrevious MoveDirection = 3
	DirectionNext     MoveDirection = 4
	DirectionFirst    MoveDirection = 5
	DirectionLast     MoveDirection = 6
)

// AXElementData represents the data returned from Move operations
type AXElementData struct {
	PlatformElementValue string `json:"platformElementValue"` // Base64-encoded platform element data
}

func (a ControlInterface) readhostAppStateChanged() {
	for {
		msg := a.channel.ReceiveMethodCall("hostAppStateChanged:")
		stateChange, err := nskeyedarchiver.Unarchive(msg.Auxiliary.GetArguments()[0].([]byte))
		if err != nil {
			panic(err)
		}
		value := stateChange[0]
		log.Infof("hostAppStateChanged:%s", value)
	}
}

func (a ControlInterface) readhostInspectorNotificationReceived() {
	for {
		msg := a.channel.ReceiveMethodCall("hostInspectorNotificationReceived:")
		notification, err := nskeyedarchiver.Unarchive(msg.Auxiliary.GetArguments()[0].([]byte))
		if err != nil {
			panic(err)
		}
		value := notification[0].(map[string]interface{})["Value"]
		log.Infof("hostInspectorNotificationReceived:%s", value)
	}
}

// init wires up event receivers and gets Info from the device
func (a ControlInterface) init() error {
	a.channel.RegisterMethodForRemote("hostInspectorCurrentElementChanged:")
	a.channel.RegisterMethodForRemote("hostInspectorMonitoredEventTypeChanged:")
	a.channel.RegisterMethodForRemote("hostAppStateChanged:")
	a.channel.RegisterMethodForRemote("hostInspectorNotificationReceived:")
	go a.readhostAppStateChanged()
	go a.readhostInspectorNotificationReceived()

	err := a.notifyPublishedCapabilities()
	if err != nil {
		return err
	}
	// a list of methods we are allowed to call on the device
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

	auditCaseIds, err := a.deviceAllAuditCaseIDs(apiVersion)
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
			log.Warnf("Failed to get human readable description for audit case ID %s: %v", v, err)
			continue
		}
		log.Infof("%s -- %s", v, name)
	}
	return nil
}

// EnableSelectionMode enables the UI element selection mode on the device,
// it is the same as clicking the little crosshair in AX Inspector
func (a ControlInterface) EnableSelectionMode() {
	a.deviceInspectorSetMonitoredEventType(2)
	a.deviceInspectorShowVisuals(true)
	a.awaitHostInspectorMonitoredEventTypeChanged()
}

// SwitchToDevice is the same as switching to the Device in AX inspector.
// After running this, notifications and events should be received.
func (a ControlInterface) SwitchToDevice() {
	a.TurnOff()
	resp, _ := a.deviceAccessibilitySettings()
	log.Info("AX Settings received:", resp)
	a.deviceInspectorShowIgnoredElements(false)
	a.deviceSetAuditTargetPid(0)
	a.deviceInspectorFocusOnElement()
	_, err := a.awaitHostInspectorCurrentElementChanged(context.Background())
	if err != nil {
		log.Warnf("await element change failed during SwitchToDevice: %v", err)
	}
	a.deviceInspectorPreviewOnElement()
	a.deviceHighlightIssue()
}

// TurnOff disable AX
func (a ControlInterface) TurnOff() {
	a.deviceInspectorSetMonitoredEventType(0)
	a.awaitHostInspectorMonitoredEventTypeChanged()
	a.deviceInspectorFocusOnElement()
	_, err := a.awaitHostInspectorCurrentElementChanged(context.Background())
	if err != nil {
		log.Warnf("await element change failed during TurnOff: %v", err)
	}
	a.deviceInspectorPreviewOnElement()
	a.deviceHighlightIssue()
	a.deviceInspectorShowVisuals(false)
}

// Move navigates focus using the given direction and returns selected element data.
func (a ControlInterface) Move(direction MoveDirection) (AXElementData, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	log.Info("changing")
	a.deviceInspectorMoveWithOptions(direction)
	log.Info("before changed")

	resp, err := a.awaitHostInspectorCurrentElementChanged(ctx)
	if err != nil {
		return AXElementData{}, err
	}

	// Extraction path for platform element bytes:
	// Value -> Value -> ElementValue_v1 -> Value -> Value -> PlatformElementValue_v1 -> Value ([]byte)
	value, ok := resp["Value"].(map[string]interface{})
	if !ok {
		log.Warn("resp[\"Value\"] is not a map")
		return AXElementData{}, nil
	}

	innerValue, ok := value["Value"].(map[string]interface{})
	if !ok {
		log.Warn("Value[\"Value\"] is not a map")
		return AXElementData{}, nil
	}

	elementValue, ok := innerValue["ElementValue_v1"].(map[string]interface{})
	if !ok {
		log.Warn("ElementValue_v1 is not a map")
		return AXElementData{}, nil
	}

	axElement, ok := elementValue["Value"].(map[string]interface{})
	if !ok {
		log.Warn("ElementValue_v1[\"Value\"] is not a map")
		return AXElementData{}, nil
	}

	// Split assertions for safety/readability
	valMap, ok := axElement["Value"].(map[string]interface{})
	if !ok {
		log.Warn("AX element inner \"Value\" is not a map")
		return AXElementData{}, nil
	}
	platformElement, ok := valMap["PlatformElementValue_v1"].(map[string]interface{})
	if !ok {
		log.Warn("PlatformElementValue_v1 is not a map")
		return AXElementData{}, nil
	}

	byteArray, ok := platformElement["Value"].([]byte)
	if !ok {
		log.Warn("PlatformElementValue_v1[\"Value\"] is not a []byte")
		return AXElementData{}, nil
	}
	encoded := base64.StdEncoding.EncodeToString(byteArray)

	return AXElementData{
		PlatformElementValue: encoded,
	}, nil
}

// GetElement moves the green selection rectangle one element further
func (a ControlInterface) GetElement() (AXElementData, error) {
	return a.Move(DirectionNext)
}

func (a ControlInterface) UpdateAccessibilitySetting(name string, val interface{}) {
	log.Info("Updating Accessibility Setting")

	resp, err := a.updateAccessibilitySetting(name, val)
	if err != nil {
		panic(fmt.Sprintf("Failed setting: %s", err))
	}
	log.Info("Setting Updated", resp)
}

func (a ControlInterface) ResetToDefaultAccessibilitySettings() error {
	err := a.channel.MethodCallAsync("deviceResetToDefaultAccessibilitySettings")
	if err != nil {
		return err
	}
	return nil
}

func (a ControlInterface) awaitHostInspectorCurrentElementChanged(ctx context.Context) (map[string]interface{}, error) {
	msg, err := a.channel.ReceiveMethodCallWithTimeout("hostInspectorCurrentElementChanged:", ctx)
	if err != nil {
		log.Errorf("Failed to receive hostInspectorCurrentElementChanged: %v", err)
		return nil, fmt.Errorf("failed to receive hostInspectorCurrentElementChanged: %w", err)
	}
	log.Info("received hostInspectorCurrentElementChanged")
	result, err := nskeyedarchiver.Unarchive(msg.Auxiliary.GetArguments()[0].([]byte))
	if err != nil {
		panic(fmt.Sprintf("Failed unarchiving: %s this is a bug and should not happen", err))
	}
	return result[0].(map[string]interface{}), nil
}

func (a ControlInterface) awaitHostInspectorMonitoredEventTypeChanged() {
	msg := a.channel.ReceiveMethodCall("hostInspectorMonitoredEventTypeChanged:")
	n, _ := nskeyedarchiver.Unarchive(msg.Auxiliary.GetArguments()[0].([]byte))
	log.Infof("hostInspectorMonitoredEventTypeChanged: was set to %d by the device", n[0])
}

func (a ControlInterface) deviceInspectorMoveWithOptions(direction MoveDirection) {
	method := "deviceInspectorMoveWithOptions:"
	options := nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
		"ObjectType": "passthrough",
		"Value": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
			"allowNonAX":        nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{"ObjectType": "passthrough", "Value": false}),
			"direction":         nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{"ObjectType": "passthrough", "Value": int32(direction)}),
			"includeContainers": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{"ObjectType": "passthrough", "Value": true}),
		}),
	})
	// str, _ := nskeyedarchiver.ArchiveXML(options)
	// println(str)
	a.channel.MethodCallAsync(method, options)
}

func (a ControlInterface) notifyPublishedCapabilities() error {
	capabs := map[string]interface{}{
		"com.apple.private.DTXBlockCompression": uint64(2),
		"com.apple.private.DTXConnection":       uint64(1),
	}
	return a.channel.MethodCallAsync("_notifyOfPublishedCapabilities:", capabs)
}

func (a ControlInterface) deviceCapabilities() ([]string, error) {
	response, err := a.channel.MethodCall("deviceCapabilities")
	if err != nil {
		return nil, err
	}
	return convertToStringList(response.Payload)
}

func (a ControlInterface) deviceAllAuditCaseIDs(api uint64) ([]string, error) {
	var response dtx.Message
	var err error
	if api >= 15 {
		response, err = a.channel.MethodCall("deviceAllSupportedAuditTypes")
	} else {
		response, err = a.channel.MethodCall("deviceAllAuditCaseIDs")
	}
	if err != nil {
		return nil, err
	}
	return convertToStringList(response.Payload)
}

func (a ControlInterface) deviceAccessibilitySettings() (map[string]interface{}, error) {
	response, err := a.channel.MethodCall("deviceAccessibilitySettings")
	if err != nil {
		return nil, err
	}
	return response.Payload[0].(map[string]interface{}), nil
}

func (a ControlInterface) deviceInspectorSupportedEventTypes() (uint64, error) {
	response, err := a.channel.MethodCall("deviceInspectorSupportedEventTypes")
	if err != nil {
		return 0, err
	}
	return response.Payload[0].(uint64), nil
}

func (a ControlInterface) deviceAPIVersion() (uint64, error) {
	response, err := a.channel.MethodCall("deviceApiVersion")
	if err != nil {
		return 0, err
	}
	return response.Payload[0].(uint64), nil
}

func (a ControlInterface) deviceInspectorCanNavWhileMonitoringEvents() (bool, error) {
	response, err := a.channel.MethodCall("deviceInspectorCanNavWhileMonitoringEvents")
	if err != nil {
		return false, err
	}
	return response.Payload[0].(bool), nil
}

func (a ControlInterface) deviceSetAppMonitoringEnabled(val bool) error {
	err := a.channel.MethodCallAsync("deviceSetAppMonitoringEnabled:", val)
	if err != nil {
		return err
	}
	return nil
}

func (a ControlInterface) updateAccessibilitySetting(settingName string, val interface{}) (string, error) {
	setting := nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
		"ObjectType": "AXAuditDeviceSetting_v1",
		"Value": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
			"ObjectType": "passthrough",
			"Value": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
				"CurrentValueNumber_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough",
					"Value":      true}),
				"EnabledValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough",
					"Value":      true,
				}),
				"IdentiifierValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough",
					"Value":      settingName,
				}),
				"SettingTypeValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough",
					"Value":      0,
				}),
				"SliderTickMarksValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough",
					"Value":      0,
				}),
			}),
		}),
	})

	value := nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
		"ObjectType": "passthrough",
		"Value":      val,
	})

	response, err := a.channel.MethodCall("deviceUpdateAccessibilitySetting:withValue:", setting, value)
	if err != nil {
		return "", err
	}
	return response.PayloadHeader.String(), nil
}

func (a ControlInterface) deviceHumanReadableDescriptionForAuditCaseID(auditCaseID string) (string, error) {
	response, err := a.channel.MethodCall("deviceHumanReadableDescriptionForAuditCaseID:", auditCaseID)
	if err != nil {
		return "", err
	}
	if len(response.Payload) == 0 {
		return "", fmt.Errorf("no payload in response")
	}
	str, ok := response.Payload[0].(string)
	if !ok {
		return "", fmt.Errorf("unexpected payload type: %T", response.Payload[0])
	}
	return str, nil
}

func (a ControlInterface) deviceInspectorShowIgnoredElements(val bool) error {
	return a.channel.MethodCallAsync("deviceInspectorShowIgnoredElements:", val)
}

func (a ControlInterface) deviceSetAuditTargetPid(pid uint64) error {
	return a.channel.MethodCallAsync("deviceSetAuditTargetPid:", pid)
}

func (a ControlInterface) deviceInspectorFocusOnElement() error {
	return a.channel.MethodCallAsync("deviceInspectorFocusOnElement:", nskeyedarchiver.NewNSNull())
}

func (a ControlInterface) deviceInspectorPreviewOnElement() error {
	return a.channel.MethodCallAsync("deviceInspectorPreviewOnElement:", nskeyedarchiver.NewNSNull())
}

func (a ControlInterface) deviceHighlightIssue() error {
	return a.channel.MethodCallAsync("deviceHighlightIssue:", map[string]interface{}{})
}

func (a ControlInterface) deviceInspectorSetMonitoredEventType(eventtype uint64) error {
	return a.channel.MethodCallAsync("deviceInspectorSetMonitoredEventType:", eventtype)
}

func (a ControlInterface) deviceInspectorShowVisuals(val bool) error {
	return a.channel.MethodCallAsync("deviceInspectorShowVisuals:", val)
}
