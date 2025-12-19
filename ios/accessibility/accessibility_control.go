package accessibility

import (
	"context"
	"encoding/base64"
	"fmt"
	"slices"
	"sync"

	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	log "github.com/sirupsen/logrus"
)

type Notification struct {
	Value interface{}
	Err   error
}

// ControlInterface provides a simple interface to controlling the AX service on the device
// It only needs the global dtx channel as all AX methods are invoked on it.
type ControlInterface struct {
	cm          *dtx.Connection
	channel     *dtx.Channel
	subscribers []chan Notification
	mu          sync.RWMutex
}

// broadcast sends a notification to all active subscribers safely.
func (a *ControlInterface) broadcast(n Notification) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	for _, ch := range a.subscribers {
		ch <- n
	}
}

// Close shuts down the connection and closes all subscriber channels.
func (a *ControlInterface) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cm != nil {
		a.cm.Close()
	}
	for _, ch := range a.subscribers {
		close(ch)
	}
	a.subscribers = nil
	return nil
}

type Action int

const (
	ActionTap Action = iota
)

type actionMeta struct {
	AttributeName string
	HumanReadable string
}

func getActionMeta(action Action) actionMeta {
	switch action {
	case ActionTap:
		return actionMeta{AttributeName: "AXAction-2010", HumanReadable: "Activate"}
	default:
		return actionMeta{}
	}
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
	SpokenDescription    string `json:"spokenDescription"`    // Spoken description of the element
}

func (a *ControlInterface) readhostAppStateChanged() {
	for {
		msg := a.channel.ReceiveMethodCall("hostAppStateChanged:")
		stateChange, err := nskeyedarchiver.Unarchive(msg.Auxiliary.GetArguments()[0].([]byte))
		if err != nil {
			log.Errorf("Error unarchiving app state change: %v", err)
			continue
		}
		value := stateChange[0]
		log.Infof("hostAppStateChanged:%s", value)
	}
}

// Subscribe returns a read-only channel for the consumer and a "cancel" function to unsubscribe
func (a *ControlInterface) Subscribe() (<-chan Notification, func()) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// create a new buffered channel for this specific consumer
	ch := make(chan Notification, 100)
	a.subscribers = append(a.subscribers, ch)

	// create a clean-up function for the caller
	unsubscribe := func() {
		a.mu.Lock()
		defer a.mu.Unlock()

		idx := slices.Index(a.subscribers, ch)
		if idx >= 0 {
			close(a.subscribers[idx])
			a.subscribers = slices.Delete(a.subscribers, idx, idx+1)
		}
	}

	return ch, unsubscribe
}

func (a *ControlInterface) readhostInspectorNotificationReceived() {
	for {
		msg := a.channel.ReceiveMethodCall("hostInspectorNotificationReceived:")
		rawBytes := msg.Auxiliary.GetArguments()[0].([]byte)

		var notification Notification
		decoded, err := nskeyedarchiver.Unarchive(rawBytes)

		if err != nil {
			notification = Notification{Err: err}
		} else {
			val := decoded[0].(map[string]interface{})["Value"]
			log.Infof("hostInspectorNotificationReceived:%s", val)
			notification = Notification{Value: val}
		}

		a.broadcast(notification)
	}
}

// init wires up event receivers and gets Info from the device
func (a *ControlInterface) init() error {
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
func (a *ControlInterface) EnableSelectionMode() {
	a.deviceInspectorSetMonitoredEventType(2)
	a.deviceInspectorShowVisuals(true)
	a.awaitHostInspectorMonitoredEventTypeChanged()
}

// SwitchToDevice is the same as switching to the Device in AX inspector.
// After running this, notifications and events should be received.
func (a *ControlInterface) SwitchToDevice() {
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
func (a *ControlInterface) TurnOff() {
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
func (a *ControlInterface) Move(ctx context.Context, direction MoveDirection) (AXElementData, error) {
	log.Info("changing")
	a.deviceInspectorMoveWithOptions(direction)
	log.Info("before changed")

	resp, err := a.awaitHostInspectorCurrentElementChanged(ctx)
	if err != nil {
		return AXElementData{}, err
	}

	innerValue, err := getInnerValue(resp)
	if err != nil {
		return AXElementData{}, err
	}

	spokenDescription := a.extractSpokenDescription(innerValue)
	platformElementBytes, err := a.extractPlatformElementBytes(innerValue)
	if err != nil {
		return AXElementData{}, err
	}

	return AXElementData{
		PlatformElementValue: base64.StdEncoding.EncodeToString(platformElementBytes),
		SpokenDescription:    spokenDescription,
	}, nil
}

func (a *ControlInterface) extractSpokenDescription(innerValue map[string]interface{}) string {
	// Try SpokenDescriptionValue_v1 first
	if desc := a.extractStringFromField(innerValue, "SpokenDescriptionValue_v1"); desc != "" {
		return desc
	}

	// Fallback to CaptionTextValue_v1
	if desc := a.extractStringFromField(innerValue, "CaptionTextValue_v1"); desc != "" {
		return desc
	}

	return ""
}

/*
PlatformElementValue_v1: A base64-encoded string that uniquely identifies an accessibility element.
It is required to perform actions on the element.

Extraction path:
    ElementValue_v1
      └── Value
          └── Value
              └── PlatformElementValue_v1
                  └── Value ([]byte)

Binary ([]byte):
    ┌──────────────────────────────┐
    │ [0x50, 0x67, 0x41, ...]      │   // Raw bytes from dtx message payload
    └──────────────────────────────┘

Base64 (string):
    ┌──────────────────────────────┐
    │ "PgAAAACikAEBAAAACg..."      │   // base64 encoded unique ID of AX element
    └──────────────────────────────┘
*/

func (a *ControlInterface) extractPlatformElementBytes(innerValue map[string]interface{}) ([]byte, error) {
	elementValue, err := getNestedMap(innerValue, "ElementValue_v1")
	if err != nil {
		return nil, fmt.Errorf("failed to get ElementValue_v1: %w", err)
	}

	axElement, err := getNestedMap(elementValue, "Value")
	if err != nil {
		return nil, fmt.Errorf("failed to get ElementValue_v1.Value: %w", err)
	}

	valMap, err := getNestedMap(axElement, "Value")
	if err != nil {
		return nil, fmt.Errorf("failed to get AX element inner Value: %w", err)
	}

	platformElement, err := getNestedMap(valMap, "PlatformElementValue_v1")
	if err != nil {
		return nil, fmt.Errorf("failed to get PlatformElementValue_v1: %w", err)
	}

	byteArray, ok := platformElement["Value"].([]byte)
	if !ok {
		return nil, fmt.Errorf("PlatformElementValue_v1.Value is not []byte, got %T", platformElement["Value"])
	}

	return byteArray, nil
}

// performAction performs the standard accessibility action without alert checking
func (a *ControlInterface) PerformAction(actionName Action, currentPlatformElementValue string) error {
	platformBytes, err := base64.StdEncoding.DecodeString(currentPlatformElementValue)
	if err != nil {
		return fmt.Errorf("invalid currentPlatformElementValue base64: %w", err)
	}

	elementArg := nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
		"ObjectType": "AXAuditElement_v1",
		"Value": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
			"ObjectType": "passthrough",
			"Value": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
				"PlatformElementValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough",
					"Value":      platformBytes,
				}),
			}),
		}),
	})

	meta := getActionMeta(actionName)

	attributeArg := nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
		"ObjectType": "AXAuditElementAttribute_v1",
		"Value": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
			"ObjectType": "passthrough",
			"Value": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
				"AttributeNameValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough", "Value": meta.AttributeName,
				}),
				"HumanReadableNameValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough", "Value": meta.HumanReadable,
				}),
				"PerformsActionValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough", "Value": true,
				}),
				"SettableValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough", "Value": false,
				}),
				"DisplayAsTree_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough", "Value": false,
				}),
				"DisplayInlineValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough", "Value": false,
				}),
				"IsInternal_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough", "Value": false,
				}),
				"ValueTypeValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough", "Value": int32(1),
				}),
			}),
		}),
	})

	valueArg := nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{})

	if _, err := a.channel.MethodCall("deviceElement:performAction:withValue:", elementArg, attributeArg, valueArg); err != nil {
		return fmt.Errorf("failed to send performAction DTX message: %w", err)
	}
	return nil
}

// GetElement moves the green selection rectangle one element further
func (a *ControlInterface) GetElement(ctx context.Context) (AXElementData, error) {
	return a.Move(ctx, DirectionNext)
}

func (a *ControlInterface) UpdateAccessibilitySetting(name string, val interface{}) {
	log.Info("Updating Accessibility Setting")

	resp, err := a.updateAccessibilitySetting(name, val)
	if err != nil {
		panic(fmt.Sprintf("Failed setting: %s", err))
	}
	log.Info("Setting Updated", resp)
}

func (a *ControlInterface) ResetToDefaultAccessibilitySettings() error {
	err := a.channel.MethodCallAsync("deviceResetToDefaultAccessibilitySettings")
	if err != nil {
		return err
	}
	return nil
}

func (a *ControlInterface) awaitHostInspectorCurrentElementChanged(ctx context.Context) (map[string]interface{}, error) {
	msg, err := a.channel.ReceiveMethodCallWithTimeout(ctx, "hostInspectorCurrentElementChanged:")
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

func (a *ControlInterface) awaitHostInspectorMonitoredEventTypeChanged() {
	msg := a.channel.ReceiveMethodCall("hostInspectorMonitoredEventTypeChanged:")
	n, _ := nskeyedarchiver.Unarchive(msg.Auxiliary.GetArguments()[0].([]byte))
	log.Infof("hostInspectorMonitoredEventTypeChanged: was set to %d by the device", n[0])
}

func (a *ControlInterface) deviceInspectorMoveWithOptions(direction MoveDirection) {
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

func (a *ControlInterface) notifyPublishedCapabilities() error {
	capabs := map[string]interface{}{
		"com.apple.private.DTXBlockCompression": uint64(2),
		"com.apple.private.DTXConnection":       uint64(1),
	}
	return a.channel.MethodCallAsync("_notifyOfPublishedCapabilities:", capabs)
}

func (a *ControlInterface) deviceCapabilities() ([]string, error) {
	response, err := a.channel.MethodCall("deviceCapabilities")
	if err != nil {
		return nil, err
	}
	return convertToStringList(response.Payload)
}

func (a *ControlInterface) deviceAllAuditCaseIDs(api uint64) ([]string, error) {
	var response dtx.Message
	var err error
	// api version 21 corresponds to iOS 15.
	if api >= 21 {
		response, err = a.channel.MethodCall("deviceAllSupportedAuditTypes")
	} else {
		response, err = a.channel.MethodCall("deviceAllAuditCaseIDs")
	}
	if err != nil {
		return nil, err
	}
	return convertToStringList(response.Payload)
}

func (a *ControlInterface) deviceAccessibilitySettings() (map[string]interface{}, error) {
	response, err := a.channel.MethodCall("deviceAccessibilitySettings")
	if err != nil {
		return nil, err
	}
	return response.Payload[0].(map[string]interface{}), nil
}

func (a *ControlInterface) deviceInspectorSupportedEventTypes() (uint64, error) {
	response, err := a.channel.MethodCall("deviceInspectorSupportedEventTypes")
	if err != nil {
		return 0, err
	}
	return response.Payload[0].(uint64), nil
}

func (a *ControlInterface) deviceAPIVersion() (uint64, error) {
	response, err := a.channel.MethodCall("deviceApiVersion")
	if err != nil {
		return 0, err
	}
	return response.Payload[0].(uint64), nil
}

func (a *ControlInterface) deviceInspectorCanNavWhileMonitoringEvents() (bool, error) {
	response, err := a.channel.MethodCall("deviceInspectorCanNavWhileMonitoringEvents")
	if err != nil {
		return false, err
	}
	return response.Payload[0].(bool), nil
}

func (a *ControlInterface) deviceSetAppMonitoringEnabled(val bool) error {
	err := a.channel.MethodCallAsync("deviceSetAppMonitoringEnabled:", val)
	if err != nil {
		return err
	}
	return nil
}

func (a *ControlInterface) updateAccessibilitySetting(settingName string, val interface{}) (string, error) {
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

func (a *ControlInterface) deviceHumanReadableDescriptionForAuditCaseID(auditCaseID string) (string, error) {
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

func (a *ControlInterface) deviceInspectorShowIgnoredElements(val bool) error {
	return a.channel.MethodCallAsync("deviceInspectorShowIgnoredElements:", val)
}

func (a *ControlInterface) deviceSetAuditTargetPid(pid uint64) error {
	return a.channel.MethodCallAsync("deviceSetAuditTargetPid:", pid)
}

func (a *ControlInterface) deviceInspectorFocusOnElement() error {
	return a.channel.MethodCallAsync("deviceInspectorFocusOnElement:", nskeyedarchiver.NewNSNull())
}

func (a *ControlInterface) deviceInspectorPreviewOnElement() error {
	return a.channel.MethodCallAsync("deviceInspectorPreviewOnElement:", nskeyedarchiver.NewNSNull())
}

func (a *ControlInterface) deviceHighlightIssue() error {
	return a.channel.MethodCallAsync("deviceHighlightIssue:", map[string]interface{}{})
}

func (a *ControlInterface) deviceInspectorSetMonitoredEventType(eventtype uint64) error {
	return a.channel.MethodCallAsync("deviceInspectorSetMonitoredEventType:", eventtype)
}

func (a *ControlInterface) deviceInspectorShowVisuals(val bool) error {
	return a.channel.MethodCallAsync("deviceInspectorShowVisuals:", val)
}
