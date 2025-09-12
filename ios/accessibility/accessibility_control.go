package accessibility

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	log "github.com/sirupsen/logrus"
)

// ControlInterface provides a simple interface to controlling the AX service on the device
// It only needs the global dtx channel as all AX methods are invoked on it.
type ControlInterface struct {
	channel                     *dtx.Channel
	currentPlatformElementValue string
	currentFocusedLabel         string
	currentCaptionText          string
	lastFetchAt                 time.Time
	labelTraverseCount          map[string]int
	wdaHost                     string        // WDA host for alert detection and actions
	wdaActionWaitTime           time.Duration // Wait time after WDA action before checking alerts again
	elementChangeTimeout        time.Duration // Timeout for waiting for element changes
}

// Direction represents navigation direction values used by AX service
const (
	DirectionPrevious int32 = 3
	DirectionNext     int32 = 4
	DirectionFirst    int32 = 5
	DirectionLast     int32 = 6
)

// NewControlInterface creates a new ControlInterface with the given channel and optional WDA host
func NewControlInterface(channel *dtx.Channel, wdaHost string) *ControlInterface {
	return &ControlInterface{
		channel:              channel,
		wdaHost:              wdaHost,
		labelTraverseCount:   make(map[string]int),
		wdaActionWaitTime:    100 * time.Millisecond, // Default wait time
		elementChangeTimeout: 5 * time.Second,        // Default timeout for element changes
	}
}

// SetWDAHost sets the WDA host for alert detection and actions
func (a *ControlInterface) SetWDAHost(wdaHost string) {
	log.Infof("Setting WDA host to: %q", wdaHost)
	a.wdaHost = wdaHost
}

// ClearWDAHost clears the WDA host, disabling alert detection
func (a *ControlInterface) ClearWDAHost() {
	log.Infof("Clearing WDA host (was: %q)", a.wdaHost)
	a.wdaHost = ""
}

// SetWDAActionWaitTime sets the wait time after WDA action before checking alerts again
func (a *ControlInterface) SetWDAActionWaitTime(waitTime time.Duration) {
	a.wdaActionWaitTime = waitTime
}

// SetElementChangeTimeout sets the timeout for waiting for element changes
func (a *ControlInterface) SetElementChangeTimeout(timeout time.Duration) {
	a.elementChangeTimeout = timeout
}

// GetWDAHost returns the current WDA host configuration
func (a *ControlInterface) GetWDAHost() string {
	return a.wdaHost
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

func (a ControlInterface) readhostFoundAuditIssue() {
	for {
		msg := a.channel.ReceiveMethodCall("hostFoundAuditIssue:")
		// optionally, parse or log; for now, just drain to avoid blocking
		_ = msg
	}
}

func (a ControlInterface) readhostAppendAuditLog() {
	for {
		msg := a.channel.ReceiveMethodCall("hostAppendAuditLog:")
		_ = msg
	}
}

// init wires up event receivers and gets Info from the device
func (a ControlInterface) init() error {
	a.channel.RegisterMethodForRemote("hostInspectorCurrentElementChanged:")
	a.channel.RegisterMethodForRemote("hostInspectorMonitoredEventTypeChanged:")
	a.channel.RegisterMethodForRemote("hostAppStateChanged:")
	a.channel.RegisterMethodForRemote("hostInspectorNotificationReceived:")
	// audit results callback
	a.channel.RegisterMethodForRemote("hostDeviceDidCompleteAuditCategoriesWithAuditIssues:")
	// some older versions might emit a slightly different selector
	a.channel.RegisterMethodForRemote("hostDeviceDidCompleteAuditCaseIDsWithAuditIssues:")
	// issue-stream callback during audit
	a.channel.RegisterMethodForRemote("hostFoundAuditIssue:")
	// audit log lines (can contain rects as text)
	a.channel.RegisterMethodForRemote("hostAppendAuditLog:")
	go a.readhostAppStateChanged()
	go a.readhostInspectorNotificationReceived()
	// drain audit issue stream and logs so they don't block dispatch
	go a.readhostFoundAuditIssue()
	go a.readhostAppendAuditLog()

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

	if a.labelTraverseCount == nil {
		a.labelTraverseCount = make(map[string]int)
	}
	// auditCaseIds, err := a.deviceAllAuditCaseIDs()
	// if err != nil {
	// 	return err
	// }
	// log.Info("AuditCaseIDs", auditCaseIds)

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

	// for _, v := range auditCaseIds {
	// 	name, err := a.deviceHumanReadableDescriptionForAuditCaseID(v)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	log.Infof("%s -- %s", v, name)
	// }
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
	a.awaitHostInspectorCurrentElementChanged()
	a.deviceInspectorPreviewOnElement()
	a.deviceHighlightIssue()
}

// TurnOff disable AX
func (a ControlInterface) TurnOff() {
	a.deviceInspectorSetMonitoredEventType(0)
	a.awaitHostInspectorMonitoredEventTypeChanged()
	a.deviceInspectorFocusOnElement()
	a.awaitHostInspectorCurrentElementChanged()
	a.deviceInspectorPreviewOnElement()
	a.deviceHighlightIssue()
	a.deviceInspectorShowVisuals(false)
}

// GetElement moves the green selection rectangle one element further
func (a *ControlInterface) GetElement() {
	log.Info("changing")
	a.deviceInspectorMoveWithOptions(DirectionNext)
	// a.deviceInspectorMoveWithOptions()
	log.Info("before changed")

	resp := a.awaitHostInspectorCurrentElementChanged()

	// Assume 'resp' is your top-level map[string]interface{} object

	value, ok := resp["Value"].(map[string]interface{})
	if !ok {
		log.Warn("resp[\"Value\"] is not a map")
		return
	}

	innerValue, ok := value["Value"].(map[string]interface{})
	if !ok {
		log.Warn("Value[\"Value\"] is not a map")
		return
	}

	// Capture caption text if present
	if capRaw, ok := innerValue["CaptionTextValue_v1"]; ok {
		capVal := a.deserializeObject(capRaw)
		if s, ok := capVal.(string); ok && s != "" {
			a.currentCaptionText = s
			log.Infof("caption: %q", s)
		} else if capVal != nil {
			a.currentCaptionText = fmt.Sprintf("%v", capVal)
		}
	}

	elementValue, ok := innerValue["ElementValue_v1"].(map[string]interface{})
	if !ok {
		log.Warn("ElementValue_v1 is not a map")
		return
	}

	axElement, ok := elementValue["Value"].(map[string]interface{})
	if !ok {
		log.Warn("ElementValue_v1[\"Value\"] is not a map")
		return
	}

	platformElement, ok := axElement["Value"].(map[string]interface{})["PlatformElementValue_v1"].(map[string]interface{})
	if !ok {
		log.Warn("PlatformElementValue_v1 is not a map")
		return
	}

	byteArray, ok := platformElement["Value"].([]byte)
	if !ok {
		log.Warn("PlatformElementValue_v1[\"Value\"] is not a []byte")
		return
	}
	encoded := base64.StdEncoding.EncodeToString(byteArray)
	// a.GetRectForPlatformElement(encoded)
	// // Now decode the byte array
	// log.Infof("Attempting to decode PlatformElementValue_v1 raw bytes: %x", byteArray)
	// decoded, err := nskeyedarchiver.Unarchive(byteArray)
	// if err != nil {
	// 	log.Warnf("Failed to decode PlatformElementValue_v1: %v", err)
	// } else {
	// 	log.Infof("Decoded PlatformElementValue_v1: %#v", decoded)
	// }
	// if err != nil {
	// 	log.Warnf("Failed to decode PlatformElementValue_v1: %v", err)
	// 	return
	// }
	a.currentPlatformElementValue = encoded

	if label, err := a.GetElementLabel(); err == nil && label != "" {
		if a.labelTraverseCount == nil {
			a.labelTraverseCount = make(map[string]int)
		}
		a.labelTraverseCount[label] = a.labelTraverseCount[label] + 1
		a.currentFocusedLabel = label
		log.Infof("label '%s' traverse count=%d", label, a.labelTraverseCount[label])
	}

	// _, _ = a.SupportedAuditTypes()

	// // Convert []interface{} -> []int32
	// var ids []int32
	// for _, v := range types {
	// 	switch t := v.(type) {
	// 	case int32:
	// 		ids = append(ids, t)
	// 	case int, int64, uint64, float64:
	// 		ids = append(ids, int32(int64(fmt.Sprintf("%v", t)[0]))) // replace with real numeric conversion in your code
	// 	}
	// }

	// removed automatic RunAudit() to avoid delaying navigation

	// log.Infof("Label: %q (err=%v)", label, err)

	// rect, err := a.GetFocusedElementRectViaAudit()
	// if err != nil {
	// 	log.Error(err)
	// 	return
	// }
	// log.Infof("rect: %#v", rect)

	// xNorm := 195.0 / 390.0
	// yNorm := 245.0 / 844.0

	// rect, _, _ = a.FetchElementAtNormalizedDeviceCoordinate(xNorm, yNorm)
	// if err != nil {
	// 	log.Error(err)
	// 	return
	// }
	// log.Infof("rect: %#v", rect)

	// _, _, _ = a.FetchElementAtNormalizedDeviceCoordinateViaAttribute(22.0, 468.0)

	// pos, err := a.GetElementFrame()
	// log.Infof("PlatformElementValue_v1 raw bytes: %v", encoded)
	// if err != nil {
	// 	log.Error(err)
	// 	return
	// }
	// log.Infof("cursor position: %#v", pos)
}

// checkForAlerts checks if there are any XCUIElementTypeAlert elements present using WDA
func (a *ControlInterface) checkForAlerts() (bool, error) {
	if a.wdaHost == "" {
		return false, fmt.Errorf("WDA host not configured")
	}

	findURL := fmt.Sprintf("%s/wda/elementsWithCoords", a.wdaHost)
	payload := map[string]string{
		"using": "predicate string",
		"value": "type == \"XCUIElementTypeAlert\"",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return false, err
	}

	resp, err := http.Post(findURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("WDA /elementsWithCoords %d: %s", resp.StatusCode, string(b))
	}

	var result struct {
		Value []struct {
			Element     string             `json:"ELEMENT"`
			ElementUUID string             `json:"element-6066-11e4-a52e-4f735466cecf"`
			Rect        map[string]float64 `json:"rect"`
		} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	// Return true if any alert elements are found
	return len(result.Value) > 0, nil
}

func (a *ControlInterface) PerformAction(actionName string) error {
	if a.currentPlatformElementValue == "" {
		return fmt.Errorf("no selected element to perform action on")
	}

	log.Infof("PerformAction called with WDA host: %q", a.wdaHost)

	// Check for alerts first - if alerts are present, use WDA action instead
	if a.wdaHost != "" {
		hasAlerts, err := a.checkForAlerts()
		if err != nil {
			log.Warnf("Failed to check for alerts, falling back to default action: %v", err)
		} else if hasAlerts {
			log.Info("Alert detected, switching to WDA action")
			_, err := a.PerformWDAAction()
			if err != nil {
				return err
			}

			// After WDA action, check again for alerts to determine next action
			// Wait for UI to update after the action
			// time.Sleep(a.wdaActionWaitTime)

			// hasAlertsAfter, err := a.checkForAlerts()
			// if err != nil {
			// 	log.Warnf("Failed to check for alerts after WDA action: %v", err)
			// 	return nil // WDA action succeeded, just log the warning
			// }

			// if hasAlertsAfter {
			// 	log.Info("Alert still present after WDA action")
			// 	return nil // Alert still there, WDA action completed
			// } else {
			// 	log.Info("Alert dismissed, switching back to default action")
			// 	// Alert was dismissed, now perform the original action
			// 	return a.performDefaultAction(actionName)
			// }
		}
	}

	// If no alerts detected or WDA not configured, perform default action
	return a.performDefaultAction(actionName)
}

// performDefaultAction performs the standard accessibility action without alert checking
func (a *ControlInterface) performDefaultAction(actionName string) error {
	platformBytes, err := base64.StdEncoding.DecodeString(a.currentPlatformElementValue)
	if err != nil {
		return fmt.Errorf("invalid PlatformElementValue_v1 base64: %w", err)
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

	attributeArg := nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
		"ObjectType": "AXAuditElementAttribute_v1",
		"Value": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
			"ObjectType": "passthrough",
			"Value": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
				"AttributeNameValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough", "Value": actionName,
				}),
				"HumanReadableNameValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough", "Value": "Activate",
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

func (a ControlInterface) awaitHostInspectorCurrentElementChanged() map[string]interface{} {
	// Use configurable timeout to prevent indefinite blocking
	msg, err := a.channel.ReceiveMethodCallWithTimeout("hostInspectorCurrentElementChanged:", a.elementChangeTimeout)
	if err != nil {
		log.Errorf("Timeout waiting for hostInspectorCurrentElementChanged (timeout: %v): %v", a.elementChangeTimeout, err)
		panic(fmt.Sprintf("Timeout waiting for hostInspectorCurrentElementChanged: %s", err))
	}
	log.Info("received hostInspectorCurrentElementChanged")
	result, err := nskeyedarchiver.Unarchive(msg.Auxiliary.GetArguments()[0].([]byte))
	if err != nil {
		panic(fmt.Sprintf("Failed unarchiving: %s this is a bug and should not happen", err))
	}
	return result[0].(map[string]interface{})
}

func (a ControlInterface) awaitHostInspectorMonitoredEventTypeChanged() {
	msg := a.channel.ReceiveMethodCall("hostInspectorMonitoredEventTypeChanged:")
	n, _ := nskeyedarchiver.Unarchive(msg.Auxiliary.GetArguments()[0].([]byte))
	log.Infof("hostInspectorMonitoredEventTypeChanged: was set to %d by the device", n[0])
}

func (a ControlInterface) GetCurrentCursorPosition() (interface{}, error) {
	resp, err := a.channel.MethodCall("deviceInspectorInformCurrentCursorPosition:", nskeyedarchiver.NewNSNull())
	if err != nil {
		return nil, err
	}
	if len(resp.Payload) > 0 {
		return resp.Payload[0], nil
	}
	// Some replies may return data in Auxiliary as an archived object
	auxArgs := resp.Auxiliary.GetArguments()
	if len(auxArgs) > 0 {
		if b, ok := auxArgs[0].([]byte); ok {
			if decoded, err := nskeyedarchiver.Unarchive(b); err == nil && len(decoded) > 0 {
				return decoded[0], nil
			}
			return b, nil
		}
		return auxArgs[0], nil
	}
	return nil, fmt.Errorf("no cursor position in reply")
}

func (a ControlInterface) deviceInspectorMoveWithOptions(direction int32) {
	method := "deviceInspectorMoveWithOptions:"
	options := nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
		"ObjectType": "passthrough",
		"Value": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
			"allowNonAX":        nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{"ObjectType": "passthrough", "Value": false}),
			"direction":         nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{"ObjectType": "passthrough", "Value": direction}),
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

func (a ControlInterface) deviceAllAuditCaseIDs() ([]string, error) {
	response, err := a.channel.MethodCall("deviceAllAuditCaseIDs")
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
	return response.Payload[0].(string), nil
}

func (a ControlInterface) deviceInspectorShowIgnoredElements(val bool) error {
	return a.channel.MethodCallAsync("deviceInspectorShowIgnoredElements:", val)
}

func (a ControlInterface) deviceSetAuditTargetPid(pid uint64) error {
	return a.channel.MethodCallAsync("deviceSetAuditTargetPid:", pid)
}

func (a ControlInterface) deviceInspectorFocusOnElement() error {
	return a.channel.MethodCallAsync("deviceInspectorFocusOnElement:", map[string]interface{}{})
}

func (a ControlInterface) deviceInspectorPreviewOnElement() error {
	resp, err := a.channel.MethodCall("deviceInspectorPreviewOnElement:", nskeyedarchiver.NewNSNull())
	if err != nil {
		return err
	}
	log.Infof("deviceInspectorPreviewOnElement reply: header=%s payload=%#v", resp.PayloadHeader.String(), resp.Payload)
	return nil
}

func (a ControlInterface) deviceHighlightIssue() error {
	return a.channel.MethodCallAsync("deviceHighlightIssue:", map[string]interface{}{})
}

func (a ControlInterface) deviceInspectorSetMonitoredEventType(eventtype uint64) error {
	return a.channel.MethodCallAsync("deviceInspectorSetMonitoredEventType:", eventtype)
}

func (a *ControlInterface) GetElementAttribute(attributeName string) (interface{}, error) {
	if a.currentPlatformElementValue == "" {
		return nil, fmt.Errorf("no selected element to query attribute for")
	}

	platformBytes, err := base64.StdEncoding.DecodeString(a.currentPlatformElementValue)
	if err != nil {
		return nil, fmt.Errorf("invalid PlatformElementValue_v1 base64: %w", err)
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

	attributeArg := nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
		"ObjectType": "AXAuditElementAttribute_v1",
		"Value": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
			"ObjectType": "passthrough",
			"Value": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
				"AttributeNameValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough",
					"Value":      attributeName,
				}),
				"DisplayAsTree_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough", "Value": false,
				}),
				"DisplayInlineValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough", "Value": false,
				}),
				"HumanReadableNameValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough", "Value": attributeName,
				}),
				"IsInternal_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough", "Value": false,
				}),
				"PerformsActionValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough", "Value": false,
				}),
				"SettableValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough", "Value": false,
				}),
				"ValueTypeValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough", "Value": int32(2),
				}),
			}),
		}),
	})

	resp, err := a.channel.MethodCall("deviceElement:valueForAttribute:", elementArg, attributeArg)
	if err != nil {
		return nil, err
	}
	if len(resp.Payload) > 0 {
		return resp.Payload[0], nil
	}
	auxArgs := resp.Auxiliary.GetArguments()
	if len(auxArgs) > 0 {
		if b, ok := auxArgs[0].([]byte); ok {
			if decoded, err := nskeyedarchiver.Unarchive(b); err == nil && len(decoded) > 0 {
				return decoded[0], nil
			}
			return b, nil
		}
		return auxArgs[0], nil
	}
	return nil, fmt.Errorf("attribute %s not found in reply", attributeName)
}

// getLabelForPlatformElement queries the Label attribute for a raw PlatformElement bytes (base64-encoded)
func (a *ControlInterface) getLabelForPlatformElement(platformBase64 string) (string, error) {
	platformBytes, err := base64.StdEncoding.DecodeString(platformBase64)
	if err != nil {
		return "", fmt.Errorf("invalid platform element base64: %w", err)
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

	attributeArg := nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
		"ObjectType": "AXAuditElementAttribute_v1",
		"Value": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
			"ObjectType": "passthrough",
			"Value": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
				"AttributeNameValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough",
					"Value":      "Label",
				}),
				"DisplayAsTree_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough", "Value": false,
				}),
				"DisplayInlineValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough", "Value": false,
				}),
				"HumanReadableNameValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough", "Value": "Label",
				}),
				"IsInternal_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough", "Value": false,
				}),
				"PerformsActionValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough", "Value": false,
				}),
				"SettableValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough", "Value": false,
				}),
				"ValueTypeValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough", "Value": int32(2),
				}),
			}),
		}),
	})

	resp, err := a.channel.MethodCall("deviceElement:valueForAttribute:", elementArg, attributeArg)
	if err != nil {
		return "", err
	}
	if len(resp.Payload) > 0 {
		if s, ok := resp.Payload[0].(string); ok {
			return s, nil
		}
		return fmt.Sprintf("%v", resp.Payload[0]), nil
	}
	auxArgs := resp.Auxiliary.GetArguments()
	if len(auxArgs) > 0 {
		if b, ok := auxArgs[0].([]byte); ok {
			if decoded, err := nskeyedarchiver.Unarchive(b); err == nil && len(decoded) > 0 {
				if s, ok := decoded[0].(string); ok {
					return s, nil
				}
				return fmt.Sprintf("%v", decoded[0]), nil
			}
		}
		return fmt.Sprintf("%v", auxArgs[0]), nil
	}
	return "", fmt.Errorf("label not found in reply")
}

func (a *ControlInterface) GetElementLabel() (string, error) {

	val, err := a.GetElementAttribute("Label")
	if err != nil {
		return "", err
	}
	s, ok := val.(string)
	if ok {
		return s, nil
	}
	return fmt.Sprintf("%v", val), nil
}

// GetRectForPlatformElement returns a rect for the provided PlatformElementValue_v1 (base64).
// Tries AXFrame, then fallbacks used by GetElementFrame.
func (a *ControlInterface) GetRectForPlatformElement(platformBase64 string) (map[string]float64, error) {
	platformBytes, err := base64.StdEncoding.DecodeString(platformBase64)
	if err != nil {
		return nil, fmt.Errorf("invalid platform element base64: %w", err)
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

	attr := func(name string) (interface{}, error) {
		attributeArg := nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
			"ObjectType": "AXAuditElementAttribute_v1",
			"Value": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
				"ObjectType": "passthrough",
				"Value": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"AttributeNameValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
						"ObjectType": "passthrough",
						"Value":      name,
					}),
					"DisplayAsTree_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
						"ObjectType": "passthrough", "Value": false,
					}),
					"DisplayInlineValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
						"ObjectType": "passthrough", "Value": false,
					}),
					"HumanReadableNameValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
						"ObjectType": "passthrough", "Value": name,
					}),
					"IsInternal_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
						"ObjectType": "passthrough", "Value": false,
					}),
					"PerformsActionValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
						"ObjectType": "passthrough", "Value": false,
					}),
					"SettableValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
						"ObjectType": "passthrough", "Value": false,
					}),
					"ValueTypeValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
						"ObjectType": "passthrough", "Value": int32(256),
					}),
				}),
			}),
		})
		resp, err := a.channel.MethodCall("deviceElement:valueForAttribute:", elementArg, attributeArg)
		if err != nil {
			return nil, err
		}
		if len(resp.Payload) > 0 {
			return resp.Payload[0], nil
		}
		aux := resp.Auxiliary.GetArguments()
		if len(aux) > 0 {
			if b, ok := aux[0].([]byte); ok {
				if decoded, err := nskeyedarchiver.Unarchive(b); err == nil && len(decoded) > 0 {
					return decoded[0], nil
				}
				return b, nil
			}
			return aux[0], nil
		}
		return nil, fmt.Errorf("attribute %s not found", name)
	}

	// Primary: AXFrame
	if v, err := attr("AXFrame"); err == nil {
		if rect, ok := tryExtractRect(v); ok {
			return rect, nil
		}
	}
	// Fallbacks
	for _, name := range []string{"AXBounds", "AXFrameInContainerSpace"} {
		if v, err := attr(name); err == nil {
			if rect, ok := tryExtractRect(v); ok {
				return rect, nil
			}
		}
	}
	// Combine pos+size
	vpos, _ := attr("AXPosition")
	vsz, _ := attr("AXSize")
	if pos, ok := tryExtractPoint(vpos); ok {
		if size, ok := tryExtractSize(vsz); ok {
			return map[string]float64{"x": pos[0], "y": pos[1], "width": size[0], "height": size[1]}, nil
		}
	}
	return nil, fmt.Errorf("rect not available for element")
}

func (a *ControlInterface) GetElementFrame() (map[string]float64, error) {
	if a.currentPlatformElementValue == "" {
		return nil, fmt.Errorf("no selected element to query frame for")
	}

	platformBytes, err := base64.StdEncoding.DecodeString(a.currentPlatformElementValue)
	if err != nil {
		return nil, fmt.Errorf("invalid PlatformElementValue_v1 base64: %w", err)
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

	// Try simplest form first: attribute as plain string (NSString)
	attributeArg := nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
		"ObjectType": "AXAuditElementAttribute_v1",
		"Value": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
			"ObjectType": "passthrough",
			"Value": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
				"AttributeNameValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough",
					"Value":      "AXFrame",
				}),
				"DisplayAsTree_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough",
					"Value":      false,
				}),
				"DisplayInlineValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough",
					"Value":      false,
				}),
				"HumanReadableNameValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough",
					"Value":      "Frame",
				}),
				"IsInternal_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough",
					"Value":      false,
				}),
				"PerformsActionValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough",
					"Value":      false,
				}),
				"SettableValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough",
					"Value":      false,
				}),
				"ValueTypeValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough",
					"Value":      int32(256),
				}),
			}),
		}),
	})

	resp, err := a.channel.MethodCall("deviceElement:valueForAttribute:", elementArg, attributeArg)
	if err != nil {
		return nil, err
	}
	if rect, ok := a.getPreviewRect(); ok {
		return rect, nil
	}

	// Attempt to extract rect from payload
	if len(resp.Payload) > 0 {
		if rect, ok := tryExtractRect(resp.Payload[0]); ok {
			return rect, nil
		}
	}
	// Try auxiliary (archived value)
	auxArgs := resp.Auxiliary.GetArguments()
	if len(auxArgs) > 0 {
		if b, ok := auxArgs[0].([]byte); ok {
			if decoded, err := nskeyedarchiver.Unarchive(b); err == nil && len(decoded) > 0 {
				if rect, ok := tryExtractRect(decoded[0]); ok {
					return rect, nil
				}
			}
		}
		if rect, ok := tryExtractRect(auxArgs[0]); ok {
			return rect, nil
		}
	}

	// Fallback 1: try additional common attribute names
	candidates := []string{"AXBounds", "AXFrameInContainerSpace"}
	for _, name := range candidates {
		val, err := a.GetElementAttribute(name)
		if err == nil {
			if rect, ok := tryExtractRect(val); ok {
				return rect, nil
			}
		}
	}

	// Fallback 2: combine position + size
	if pos, ok := tryExtractPoint(mustGetAttr(a, "AXPosition")); ok {
		if size, ok := tryExtractSize(mustGetAttr(a, "AXSize")); ok {
			return map[string]float64{"x": pos[0], "y": pos[1], "width": size[0], "height": size[1]}, nil
		}
	}

	// Fallback 3: use preview reply (often includes a rect-like structure)
	if rect, ok := a.getPreviewRect(); ok {
		return rect, nil
	}

	// Fallback 4: use current cursor position + size
	if cur, err := a.GetCurrentCursorPosition(); err == nil {
		if pos, ok := tryExtractPoint(cur); ok {
			if size, ok := tryExtractSize(mustGetAttr(a, "AXSize")); ok {
				return map[string]float64{"x": pos[0], "y": pos[1], "width": size[0], "height": size[1]}, nil
			}
		}
	}

	return nil, fmt.Errorf("frame not found in reply")
}

func tryExtractRect(v interface{}) (map[string]float64, bool) {
	// handle NSValue decoded rects
	if nv, ok := v.(nskeyedarchiver.NSValue); ok {
		if rect, ok := parseNSRectString(nv.NSRectval); ok {
			return rect, true
		}
	}
	if s, ok := v.(string); ok {
		if rect, ok := parseNSRectString(s); ok {
			return rect, true
		}
	}

	m, ok := v.(map[string]interface{})
	if !ok {
		return nil, false
	}
	// Common key variants
	keys := [][]string{
		{"X", "Y", "Width", "Height"},
		{"x", "y", "width", "height"},
		{"XValue", "YValue", "WidthValue", "HeightValue"},
	}
	for _, ks := range keys {
		x, xok := toFloat(m[ks[0]])
		y, yok := toFloat(m[ks[1]])
		w, wok := toFloat(m[ks[2]])
		h, hok := toFloat(m[ks[3]])
		if xok && yok && wok && hok {
			return map[string]float64{"x": x, "y": y, "width": w, "height": h}, true
		}
	}
	return nil, false
}

var rectNumRe = regexp.MustCompile(`-?\d+(?:\.\d+)?`)

func parseNSRectString(s string) (map[string]float64, bool) {
	nums := rectNumRe.FindAllString(s, 4)
	if len(nums) < 4 {
		return nil, false
	}
	parse := func(ss string) (float64, bool) {
		f, err := strconv.ParseFloat(ss, 64)
		return f, err == nil
	}
	x, ok1 := parse(nums[0])
	y, ok2 := parse(nums[1])
	w, ok3 := parse(nums[2])
	h, ok4 := parse(nums[3])
	if ok1 && ok2 && ok3 && ok4 {
		return map[string]float64{"x": x, "y": y, "width": w, "height": h}, true
	}
	return nil, false
}

func toFloat(v interface{}) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int32:
		return float64(t), true
	case int64:
		return float64(t), true
	case uint:
		return float64(t), true
	case uint32:
		return float64(t), true
	case uint64:
		return float64(t), true
	default:
		return 0, false
	}
}

// tryExtractPoint attempts to parse a point-like value and returns [x,y]
func tryExtractPoint(v interface{}) ([2]float64, bool) {
	// NSValue string, like "{x, y}" or "x= y="
	if s, ok := v.(string); ok {
		nums := rectNumRe.FindAllString(s, 2)
		if len(nums) >= 2 {
			xOk, yOk := false, false
			var x, y float64
			if xv, ok := strconv.ParseFloat(nums[0], 64); ok == nil {
				x = xv
				xOk = true
			}
			if yv, ok := strconv.ParseFloat(nums[1], 64); ok == nil {
				y = yv
				yOk = true
			}
			if xOk && yOk {
				return [2]float64{x, y}, true
			}
		}
	}
	// map-like
	if m, ok := v.(map[string]interface{}); ok {
		if x, xok := toFloat(m["x"]); xok {
			if y, yok := toFloat(m["y"]); yok {
				return [2]float64{x, y}, true
			}
		}
		if x, xok := toFloat(m["X"]); xok {
			if y, yok := toFloat(m["Y"]); yok {
				return [2]float64{x, y}, true
			}
		}
	}
	return [2]float64{}, false
}

// tryExtractSize attempts to parse a size-like value and returns [w,h]
func tryExtractSize(v interface{}) ([2]float64, bool) {
	if s, ok := v.(string); ok {
		nums := rectNumRe.FindAllString(s, 2)
		if len(nums) >= 2 {
			if wv, err := strconv.ParseFloat(nums[0], 64); err == nil {
				if hv, err := strconv.ParseFloat(nums[1], 64); err == nil {
					return [2]float64{wv, hv}, true
				}
			}
		}
	}
	if m, ok := v.(map[string]interface{}); ok {
		if w, wok := toFloat(m["width"]); wok {
			if h, hok := toFloat(m["height"]); hok {
				return [2]float64{w, h}, true
			}
		}
		if w, wok := toFloat(m["Width"]); wok {
			if h, hok := toFloat(m["Height"]); hok {
				return [2]float64{w, h}, true
			}
		}
	}
	return [2]float64{}, false
}

// mustGetAttr fetches an attribute; returns nil if missing
func mustGetAttr(a *ControlInterface, name string) interface{} {
	v, err := a.GetElementAttribute(name)
	if err != nil {
		return nil
	}
	return v
}

// getPreviewRect calls deviceInspectorPreviewOnElement and tries to extract a rect from reply
func (a ControlInterface) getPreviewRect() (map[string]float64, bool) {
	resp, err := a.channel.MethodCall("deviceInspectorPreviewOnElement:", nskeyedarchiver.NewNSNull())
	if err == nil {
		// payload first
		for _, p := range resp.Payload {
			if rect, ok := tryExtractRect(p); ok {
				return rect, true
			}
		}
		// aux archived
		for _, arg := range resp.Auxiliary.GetArguments() {
			if b, ok := arg.([]byte); ok {
				if decoded, err := nskeyedarchiver.Unarchive(b); err == nil {
					for _, dv := range decoded {
						if rect, ok := tryExtractRect(dv); ok {
							return rect, true
						}
					}
				}
			} else if rect, ok := tryExtractRect(arg); ok {
				return rect, true
			}
		}
	}
	return nil, false
}

func (a *ControlInterface) ProbeAttributes(attributeNames []string) {
	for _, name := range attributeNames {
		val, err := a.GetElementAttribute(name)
		if err != nil {
			log.Infof("attr %s: err=%v", name, err)
			continue
		}
		if rect, ok := tryExtractRect(val); ok {
			log.Infof("attr %s: RECT=%#v", name, rect)
			continue
		}
		log.Infof("attr %s: val=%#v", name, val)
	}
}

func (a *ControlInterface) ProbeDefaultGeometryAttributes() {
	defaultAttrs := []string{
		// "AXFrame",
		// "AXActivationPoint",
		// "AXPosition",
		// "AXSize",
		// "AXBounds",
		// "AXFrameInContainerSpace",
		// sanity checks
		"axElement",
		"elementRef",
		"Label",
		"Value",
		"Title",
		"Name",
		"Header",
		"Identifier",
	}
	a.ProbeAttributes(defaultAttrs)
}

func (a *ControlInterface) ProbeRectLikeAttributes() {
	candidates := []string{
		"Header",
		"ElementMemoryAddress",
		"Label",
		// "Label",
		// "_elementRect",
		// "elementRect",
		// "V_elementRect",
		// "ElementRect",
		// "elementRect",
		// "ElementRectValue_v1",
		// "elementRectValue_v1",
		// "ElementRectValue_v1",

		// "Value",
		// "Type",
		// "Identifier",

		// "AXFrame",
		// "AXBounds",
		// "axposition",
		// "AXSize",
		// "AXActivationPoint",
		// "AXFrameInContainerSpace",
		// "AXAuditRect_v1",
		// "RectValue_v1",
		// "AXFrameValue_v1",
		// "AXBoundsValue_v1",
		// "AXPositionValue_v1",
		// // mac-style variants that may be bridged
		// "accessibilityFrame",
		// "accessibilityActivationPoint",
		// // generic fallbacks observed in symbol dumps
		// "Frame",
		// "frame",
		// "bounds",
		// "position",
		// "size",
		// "rect",
		// "rect",
		// "ElementRect",
		// "elementRect",
		// "displayBounds",
		// "borderFrame",
		// "imageFrame",
	}
	a.ProbeAttributes(candidates)
}

// ===== Accessibility Audit Support =====

// AuditType mirrors known audit categories observed in third-party clients
// and symbol dumps. Values are best-effort and may vary by iOS version.
const (
	auditTypeDynamicText           = 3001
	auditTypeDynamicTextAlt        = 3002
	auditTypeTextClipped           = 3003
	auditTypeElementDetection      = 1000
	auditTypeSufficientDescription = 5000
	auditTypeHitRegion             = 100
	auditTypeContrast              = 12
	auditTypeContrastAlt           = 13
)

var auditTypeDescriptions = map[int]string{
	auditTypeDynamicText:           "testTypeDynamicText",
	auditTypeDynamicTextAlt:        "testTypeDynamicText",
	auditTypeTextClipped:           "testTypeTextClipped",
	auditTypeElementDetection:      "testTypeElementDetection",
	auditTypeSufficientDescription: "testTypeSufficientElementDescription",
	auditTypeHitRegion:             "testTypeHitRegion",
	auditTypeContrast:              "testTypeContrast",
	auditTypeContrastAlt:           "testTypeContrast",
}

// FetchElementAtNormalizedDeviceCoordinate queries the AX service for the element
// located at the normalized device coordinate (x,y) where both are in [0.0, 1.0].
// It returns the PlatformElementValue_v1 as base64 (if found) and the decoded object tree.
func (a *ControlInterface) FetchElementAtNormalizedDeviceCoordinate(x, y float64) (string, interface{}, error) {
	// Build CGPoint-like structure via passthrough dictionary
	// coord := nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
	// 	"ObjectType": "passthrough",
	// 	"Value": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
	// 		"x": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
	// 			"ObjectType": "passthrough", "Value": x,
	// 		}),
	// 		"y": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
	// 			"ObjectType": "passthrough", "Value": y,
	// 		}),
	// 	}),
	// })
	// if needed: yNorm = 1.0 - (yPx / h)

	coord := nskeyedarchiver.NewNSValuePoint(x, y)
	// if xml, err := nskeyedarchiver.ArchiveXML(coord); err == nil {
	// 	log.Infof("AX arg (coord) plist:\n%s", xml)
	// }
	resp, err := a.channel.MethodCall("deviceFetchElementAtNormalizedDeviceCoordinate:", coord)
	if err != nil {
		return "", nil, err
	}

	// Try payload first
	if len(resp.Payload) > 0 {
		if b, ok := resp.Payload[0].([]byte); ok {
			if decoded, err := nskeyedarchiver.Unarchive(b); err == nil && len(decoded) > 0 {
				obj := a.deserializeObject(decoded)
				if be, ok := findPlatformElementBase64(obj); ok {
					// Remember as current for follow-up calls
					a.currentPlatformElementValue = be
					return be, obj, nil
				}
				return "", obj, nil
			}
		} else {
			obj := a.deserializeObject(resp.Payload[0])
			if be, ok := findPlatformElementBase64(obj); ok {
				a.currentPlatformElementValue = be
				return be, obj, nil
			}
			return "", obj, nil
		}
	}

	// Try auxiliary archived arguments
	for _, arg := range resp.Auxiliary.GetArguments() {
		if b, ok := arg.([]byte); ok {
			if decoded, err := nskeyedarchiver.Unarchive(b); err == nil && len(decoded) > 0 {
				obj := a.deserializeObject(decoded)
				if be, ok := findPlatformElementBase64(obj); ok {
					a.currentPlatformElementValue = be
					return be, obj, nil
				}
				return "", obj, nil
			}
		} else {
			obj := a.deserializeObject(arg)
			if be, ok := findPlatformElementBase64(obj); ok {
				a.currentPlatformElementValue = be
				return be, obj, nil
			}
			return "", obj, nil
		}
	}
	return "", nil, fmt.Errorf("no element returned for normalized coordinate")
}

// FetchElementAtNormalizedDeviceCoordinateViaAttribute uses the system-wide element
// and attribute 91701 with a NSValue(CGPoint) parameter to fetch an element.
// Coordinates must be normalized [0..1]. This respects a 0.1s throttle like the
// ObjC reference implementation.
func (a *ControlInterface) FetchElementAtNormalizedDeviceCoordinateViaAttribute(x, y float64) (string, interface{}, error) {
	if dt := time.Since(a.lastFetchAt); dt < 100*time.Millisecond {
		// throttle: return last known if available
		if a.currentPlatformElementValue != "" {
			return a.currentPlatformElementValue, nil, nil
		}
	}
	coord := nskeyedarchiver.NewNSValuePoint(x, y)

	// system wide element as an AXAuditElement_v1 wrapper
	sysWide := nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
		"ObjectType": "AXAuditElement_v1",
		"Value": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
			"ObjectType": "passthrough",
			"Value":      nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
				// Convention: system-wide element may be encoded with a sentinel value or empty dict
			}),
		}),
	})

	// Attribute 91701 with parameter = NSValue(CGPoint)
	attr := nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
		"ObjectType": "AXAuditElementAttribute_v1",
		"Value": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
			"ObjectType": "passthrough",
			"Value": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
				"AttributeNameValue_v1": nskeyedarchiver.NewNSMutableDictionary(map[string]interface{}{
					"ObjectType": "passthrough",
					"Value":      int32(91701),
				}),
			}),
		}),
	})

	resp, err := a.channel.MethodCall("deviceElement:valueForAttribute:", sysWide, attr, coord)
	if err != nil {
		return "", nil, err
	}
	// parse like other helpers
	// payload first
	if len(resp.Payload) > 0 {
		if b, ok := resp.Payload[0].([]byte); ok {
			if decoded, err := nskeyedarchiver.Unarchive(b); err == nil && len(decoded) > 0 {
				obj := a.deserializeObject(decoded)
				if be, ok := findPlatformElementBase64(obj); ok {
					a.currentPlatformElementValue = be
					a.lastFetchAt = time.Now()
					return be, obj, nil
				}
				return "", obj, nil
			}
		} else {
			obj := a.deserializeObject(resp.Payload[0])
			if be, ok := findPlatformElementBase64(obj); ok {
				a.currentPlatformElementValue = be
				a.lastFetchAt = time.Now()
				return be, obj, nil
			}
			return "", obj, nil
		}
	}
	for _, arg := range resp.Auxiliary.GetArguments() {
		if b, ok := arg.([]byte); ok {
			if decoded, err := nskeyedarchiver.Unarchive(b); err == nil && len(decoded) > 0 {
				obj := a.deserializeObject(decoded)
				if be, ok := findPlatformElementBase64(obj); ok {
					a.currentPlatformElementValue = be
					a.lastFetchAt = time.Now()
					return be, obj, nil
				}
				return "", obj, nil
			}
		} else {
			obj := a.deserializeObject(arg)
			if be, ok := findPlatformElementBase64(obj); ok {
				a.currentPlatformElementValue = be
				a.lastFetchAt = time.Now()
				return be, obj, nil
			}
			return "", obj, nil
		}
	}
	return "", nil, fmt.Errorf("no element for normalized coordinate via attribute")
}

// findPlatformElementBase64 walks an arbitrary decoded structure and returns the
// PlatformElementValue_v1 bytes as base64, if present.
func findPlatformElementBase64(v interface{}) (string, bool) {
	switch t := v.(type) {
	case []interface{}:
		for _, it := range t {
			if be, ok := findPlatformElementBase64(it); ok {
				return be, true
			}
		}
	case map[string]interface{}:
		if pe, ok := t["PlatformElementValue_v1"]; ok {
			if pm, ok := pe.(map[string]interface{}); ok {
				if b, ok := pm["Value"].([]byte); ok {
					return base64.StdEncoding.EncodeToString(b), true
				}
			}
		}
		// Dive deeper
		for _, vv := range t {
			if be, ok := findPlatformElementBase64(vv); ok {
				return be, true
			}
		}
	}
	return "", false
}

// AXAuditIssueV1 is a simplified Go representation of AXAuditIssue_v1
// Only fields useful for consumers are exposed.
type AXAuditIssueV1 struct {
	ElementRect              map[string]float64 `json:"element_rect_value,omitempty"`
	IssueClassificationRaw   interface{}        `json:"-"`
	IssueClassificationLabel string             `json:"issue_classification"`
	FontSize                 interface{}        `json:"font_size,omitempty"`
	MLGeneratedDescription   interface{}        `json:"ml_generated_description,omitempty"`
	LongDescriptionExtraInfo interface{}        `json:"long_description_extra_info,omitempty"`
	ForegroundColor          interface{}        `json:"foreground_color,omitempty"`
	BackgroundColor          interface{}        `json:"background_color,omitempty"`
	PlatformElementBase64    string             `json:"platform_element_value,omitempty"`
	Label                    string             `json:"label,omitempty"`
}

// SupportedAuditTypes returns the list of supported audit category identifiers.
// On iOS 15+ this maps to deviceAllSupportedAuditTypes, otherwise deviceAllAuditCaseIDs.
func (a *ControlInterface) SupportedAuditTypes() ([]interface{}, error) {
	api, err := a.deviceAPIVersion()
	if err != nil {
		return nil, err
	}
	var resp dtx.Message
	if api >= 15 {
		resp, err = a.channel.MethodCall("deviceAllSupportedAuditTypes")
	} else {
		resp, err = a.channel.MethodCall("deviceAllAuditCaseIDs")
	}
	if err != nil {
		return nil, err
	}
	// Flatten common return shapes
	if len(resp.Payload) > 0 {
		if list, ok := resp.Payload[0].([]interface{}); ok {
			return list, nil
		}
	}
	aux := resp.Auxiliary.GetArguments()
	if len(aux) > 0 {
		if b, ok := aux[0].([]byte); ok {
			if decoded, err := nskeyedarchiver.Unarchive(b); err == nil && len(decoded) > 0 {
				if list, ok := decoded[0].([]interface{}); ok {
					return list, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("unsupported supported-audits reply format")
}

// RunAudit triggers the accessibility audit for all supported audit types and waits
// for the completion callback with the resulting issues.
// For iOS >= 15, auditTypes are category integers. For older iOS, case IDs are used.
func (a *ControlInterface) RunAudit() ([]AXAuditIssueV1, error) {
	// Fetch the supported types from the device first
	supported, err := a.SupportedAuditTypes()
	if err != nil {
		return nil, err
	}
	// Convert the []interface{} returned by device to []int32
	auditTypes := make([]int32, 0, len(supported))
	for _, v := range supported {
		switch t := v.(type) {
		case int32:
			auditTypes = append(auditTypes, t)
		case int64:
			auditTypes = append(auditTypes, int32(t))
		case int:
			auditTypes = append(auditTypes, int32(t))
		case float64:
			auditTypes = append(auditTypes, int32(t))
		case uint64:
			auditTypes = append(auditTypes, int32(t))
		}
	}
	api, err := a.deviceAPIVersion()
	if err != nil {
		return nil, err
	}

	// NSKeyedArchiver requires []interface{} (NSArray), convert the ints
	values := make([]interface{}, len(auditTypes))
	for i, v := range auditTypes {
		values[i] = int32(v)
	}

	// Start audit. The result is delivered asynchronously as a host* callback.
	if api >= 15 {
		if err := a.channel.MethodCallAsync("deviceBeginAuditTypes:", values); err != nil {
			return nil, err
		}
	} else {
		if err := a.channel.MethodCallAsync("deviceBeginAuditCaseIDs:", values); err != nil {
			return nil, err
		}
	}

	// Wait for completion event and parse issues; support selector variants
	msg := a.channel.ReceiveMethodCall("hostDeviceDidCompleteAuditCategoriesWithAuditIssues:")
	if len(msg.Auxiliary.GetArguments()) == 0 && len(msg.Payload) == 0 {
		msg = a.channel.ReceiveMethodCall("hostDeviceDidCompleteAuditCaseIDsWithAuditIssues:")
	}

	var allIssues []AXAuditIssueV1

	// Look into auxiliary arguments first
	args := msg.Auxiliary.GetArguments()
	for _, aarg := range args {
		if b, ok := aarg.([]byte); ok {
			if decoded, err := nskeyedarchiver.Unarchive(b); err == nil {
				issues := a.extractIssuesFromInterface(a.deserializeObject(decoded))
				allIssues = append(allIssues, issues...)
			}
			continue
		}
		issues := a.extractIssuesFromInterface(aarg)
		allIssues = append(allIssues, issues...)
	}

	// Also scan payload in case values are embedded there
	for _, p := range msg.Payload {
		issues := a.extractIssuesFromInterface(p)
		allIssues = append(allIssues, issues...)
	}

	if len(allIssues) == 0 {
		return nil, fmt.Errorf("no audit issues found in reply")
	}
	// Enrich with labels when element bytes are available
	for i := range allIssues {
		if allIssues[i].PlatformElementBase64 == "" {
			continue
		}
		label, err := a.getLabelForPlatformElement(allIssues[i].PlatformElementBase64)
		if err == nil && label != "" {
			allIssues[i].Label = label
		}
	}
	return allIssues, nil
}

// extractIssuesFromInterface walks an arbitrary decoded structure to find a list
// of AXAuditIssue_v1-like maps and converts them to AXAuditIssueV1.
func (a ControlInterface) extractIssuesFromInterface(root interface{}) []AXAuditIssueV1 {
	// Attempt to locate a list of maps that contain known AXAuditIssue_v1 keys
	var out []AXAuditIssueV1

	var walk func(v interface{}) (found []map[string]interface{})
	walk = func(v interface{}) (found []map[string]interface{}) {
		switch t := v.(type) {
		case []interface{}:
			// If this slice looks like issues, return it directly
			if len(t) > 0 {
				if _, ok := t[0].(map[string]interface{}); ok {
					// Check sentinel keys on first element
					m := t[0].(map[string]interface{})
					if hasAnyKey(m, "AXAuditIssue_v1", "IssueClassificationValue_v1", "ElementRectValue_v1") {
						for _, it := range t {
							if mm, ok := it.(map[string]interface{}); ok {
								found = append(found, mm)
							}
						}
						return
					}
				}
			}
			// Otherwise recurse
			for _, it := range t {
				found = append(found, walk(it)...)
			}
		case map[string]interface{}:
			// Direct value under common keys
			if val, ok := t["value"]; ok {
				found = append(found, walk(val)...)
			}
			if val, ok := t["Value"]; ok {
				found = append(found, walk(val)...)
			}
			for _, v2 := range t {
				found = append(found, walk(v2)...)
			}
		}
		return
	}

	candidates := walk(root)
	for _, m := range candidates {
		// Some decoders keep the typed object name; accept both flat and nested forms
		if val, ok := m["Value"]; ok {
			if mv, ok := val.(map[string]interface{}); ok {
				m = mv
			}
		}
		issue := AXAuditIssueV1{}

		// IssueClassificationValue_v1
		if ic, ok := m["IssueClassificationValue_v1"]; ok {
			issue.IssueClassificationRaw = ic
			if iv, ok := toFloat(ic); ok {
				// map to human label when possible
				lbl, ok := auditTypeDescriptions[int(iv)]
				if ok {
					issue.IssueClassificationLabel = lbl
				} else {
					issue.IssueClassificationLabel = fmt.Sprintf("%v", ic)
				}
			} else {
				issue.IssueClassificationLabel = fmt.Sprintf("%v", ic)
			}
		}

		// Rect extraction
		if r, ok := m["ElementRectValue_v1"]; ok {
			if rect, ok := tryExtractRect(a.deserializeObject(r)); ok {
				issue.ElementRect = rect
			}
		}

		// Optional: extract element bytes for later attribute lookups
		if ev, ok := m["ElementValue_v1"]; ok {
			if mv, ok := a.deserializeObject(ev).(map[string]interface{}); ok {
				if pe, ok := mv["PlatformElementValue_v1"]; ok {
					if pm, ok := a.deserializeObject(pe).(map[string]interface{}); ok {
						if b, ok := pm["Value"].([]byte); ok {
							issue.PlatformElementBase64 = base64.StdEncoding.EncodeToString(b)
						}
					}
				}
			}
		}

		// Optional fields
		if v, ok := m["FontSizeValue_v1"]; ok {
			issue.FontSize = a.deserializeObject(v)
		}
		if v, ok := m["MLGeneratedDescriptionValue_v1"]; ok {
			issue.MLGeneratedDescription = a.deserializeObject(v)
		}
		if v, ok := m["ElementLongDescExtraInfo_v1"]; ok {
			issue.LongDescriptionExtraInfo = a.deserializeObject(v)
		}
		if v, ok := m["ForegroundColorValue_v1"]; ok {
			issue.ForegroundColor = a.deserializeObject(v)
		}
		if v, ok := m["BackgroundColorValue_v1"]; ok {
			issue.BackgroundColor = a.deserializeObject(v)
		}

		// Only append if we recognized it as an issue
		if issue.IssueClassificationLabel != "" || issue.ElementRect != nil {
			out = append(out, issue)
		}
	}
	return out
}

// deserializeObject mirrors the Python helper: unwraps {ObjectType: 'passthrough', Value: ...}
// and recursively processes containers. For other typed objects, returns their Value recursively.
func (a ControlInterface) deserializeObject(d interface{}) interface{} {
	switch t := d.(type) {
	case []interface{}:
		out := make([]interface{}, 0, len(t))
		for _, v := range t {
			out = append(out, a.deserializeObject(v))
		}
		return out
	case map[string]interface{}:
		if ot, ok := t["ObjectType"]; ok {
			if ot == "passthrough" {
				return a.deserializeObject(t["Value"])
			}
			// For other typed objects, we generally care about their 'Value'
			if v, ok := t["Value"]; ok {
				return a.deserializeObject(v)
			}
			return t
		}
		// Plain dictionary: recursively process values
		out := make(map[string]interface{}, len(t))
		for k, v := range t {
			out[k] = a.deserializeObject(v)
		}
		return out
	default:
		return d
	}
}

func hasAnyKey(m map[string]interface{}, keys ...string) bool {
	for _, k := range keys {
		if _, ok := m[k]; ok {
			return true
		}
	}
	return false
}

// Removed WDAElementInfo usage; PerformWDAAction returns uuid after tapping the element
func (a *ControlInterface) PerformWDAAction() (string, error) {
	// Use the currently focused label captured during GetElement()
	lastLabel := a.currentFocusedLabel
	if lastLabel == "" {
		return "", fmt.Errorf("no currently focused label to act on")
	}
	// Fetch the traverse count for this focused label; default to 1 when missing
	lastCount := 1
	if a.labelTraverseCount != nil {
		if c, ok := a.labelTraverseCount[lastLabel]; ok && c > 0 {
			lastCount = c
		}
	}
	// normalize cases where label was stored like: map[ObjectType:passthrough Value:Photos]
	if strings.HasPrefix(lastLabel, "map[") {
		re := regexp.MustCompile(`Value:([^\]]+)`)
		m := re.FindStringSubmatch(lastLabel)
		if len(m) > 1 {
			lastLabel = strings.TrimSpace(m[1])
		}
	}

	// Check if we're dealing with a switch based on currentCaptionText
	isSwitch := false
	if a.currentCaptionText != "" {
		// Look for patterns like "Airplane Mode, 0, Button, Toggle" or "Airplane Mode, 0, Button,"
		caption := strings.ToLower(a.currentCaptionText)
		if strings.Contains(caption, "toggle") ||
			(strings.Contains(caption, "button") && strings.Contains(caption, ",")) {
			isSwitch = true
			log.Infof("Detected switch element based on caption: %q", a.currentCaptionText)
		}
	}

	// 1) find elements by label predicate via WDA
	log.Infof("PerformWDAAction: WDA host is: %q", a.wdaHost)
	if a.wdaHost == "" {
		return "", fmt.Errorf("WDA host is not set - please enable accessibility service with wda_host parameter")
	}
	findURL := fmt.Sprintf("%s/wda/elementsWithCoords", a.wdaHost)
	payload := map[string]string{"using": "predicate string", "value": fmt.Sprintf("label == \"%s\"", lastLabel)}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(findURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("WDA /elements %d: %s", resp.StatusCode, string(b))
	}
	var result struct {
		Value []struct {
			Element     string             `json:"ELEMENT"`
			ElementUUID string             `json:"element-6066-11e4-a52e-4f735466cecf"`
			Rect        map[string]float64 `json:"rect"`
			Type        string             `json:"type"`
		} `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result.Value) == 0 {
		return "", fmt.Errorf("no elements found for label %q", lastLabel)
	}

	// 2) choose element by index = count-1 (0-based). clamp to range
	idx := lastCount - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(result.Value) {
		idx = len(result.Value) - 1
	}

	// 3) If we detected a switch, look for XCUIElementTypeSwitch in the results
	if isSwitch {
		// First try to find a switch element at the same index
		if idx < len(result.Value) && result.Value[idx].Type == "XCUIElementTypeSwitch" {
			log.Infof("Found switch element at index %d", idx)
		} else {
			// Look for any switch element in the results
			switchFound := false
			for i, elem := range result.Value {
				if elem.Type == "XCUIElementTypeSwitch" {
					idx = i
					switchFound = true
					log.Infof("Found switch element at index %d (was %d)", i, lastCount-1)
					break
				}
			}
			if !switchFound {
				log.Warnf("Switch detected but no XCUIElementTypeSwitch found in %d elements", len(result.Value))
			}
		}
	}

	uuid := result.Value[idx].ElementUUID
	if uuid == "" {
		uuid = result.Value[idx].Element
	}
	if uuid == "" {
		return "", fmt.Errorf("element uuid not present in WDA reply")
	}

	// 4) If it's a switch, use WDA click method; otherwise use testobject tap
	if isSwitch {
		log.Infof("Clicking switch element with UUID: %s", uuid)
		err := a.clickElementByUUID(uuid)
		if err != nil {
			return "", fmt.Errorf("failed to click switch element: %w", err)
		}
	} else {
		// use rect from elementsWithCoords to get x,y and call testobject tap with arguments {x,y,duration}
		rect := result.Value[idx].Rect
		x := rect["x"]
		y := rect["y"]
		if w, ok := rect["width"]; ok {
			x += w / 2.0
		}
		if h, ok := rect["height"]; ok {
			y += h / 2.0
		}
		testObjectURL := fmt.Sprintf("%s/testobject/tap", a.wdaHost)
		tapArgs := map[string]float64{
			"x":        x,
			"y":        y,
			"duration": 0,
		}
		tapBody, _ := json.Marshal(tapArgs)
		tapResp, err := http.Post(testObjectURL, "application/json", bytes.NewReader(tapBody))
		if err != nil {
			return "", err
		}
		defer tapResp.Body.Close()
		if tapResp.StatusCode < 200 || tapResp.StatusCode >= 300 {
			b, _ := io.ReadAll(tapResp.Body)
			return "", fmt.Errorf("testobject/tap %d: %s", tapResp.StatusCode, string(b))
		}
	}
	return uuid, nil
}

// clickElementByUUID clicks on an element using WDA's element click method
func (a *ControlInterface) clickElementByUUID(uuid string) error {
	if a.wdaHost == "" {
		return fmt.Errorf("WDA host is not set")
	}

	// Use WDA's element click method
	clickURL := fmt.Sprintf("%s/element/%s/click", a.wdaHost, uuid)
	log.Infof("Clicking element via WDA: %s", clickURL)

	resp, err := http.Post(clickURL, "application/json", nil)
	if err != nil {
		return fmt.Errorf("failed to send click request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("WDA element click failed %d: %s", resp.StatusCode, string(b))
	}

	log.Infof("Element %s clicked successfully", uuid)
	return nil
}

func (a ControlInterface) deviceInspectorShowVisuals(val bool) error {
	return a.channel.MethodCallAsync("deviceInspectorShowVisuals:", val)
}

// Move navigates focus using the given direction and updates current element and label state
func (a *ControlInterface) Move(direction int32) {
	log.Info("changing")
	a.deviceInspectorMoveWithOptions(direction)
	// a.deviceInspectorMoveWithOptions()
	log.Info("before changed")

	resp := a.awaitHostInspectorCurrentElementChanged()

	// Assume 'resp' is your top-level map[string]interface{} object

	value, ok := resp["Value"].(map[string]interface{})
	if !ok {
		log.Warn("resp[\"Value\"] is not a map")
		return
	}

	innerValue, ok := value["Value"].(map[string]interface{})
	if !ok {
		log.Warn("Value[\"Value\"] is not a map")
		return
	}

	// Capture caption text if present
	if capRaw, ok := innerValue["CaptionTextValue_v1"]; ok {
		capVal := a.deserializeObject(capRaw)
		if s, ok := capVal.(string); ok && s != "" {
			a.currentCaptionText = s
			log.Infof("caption: %q", s)
		} else if capVal != nil {
			a.currentCaptionText = fmt.Sprintf("%v", capVal)
		}
	}

	elementValue, ok := innerValue["ElementValue_v1"].(map[string]interface{})
	if !ok {
		log.Warn("ElementValue_v1 is not a map")
		return
	}

	axElement, ok := elementValue["Value"].(map[string]interface{})
	if !ok {
		log.Warn("ElementValue_v1[\"Value\"] is not a map")
		return
	}

	platformElement, ok := axElement["Value"].(map[string]interface{})["PlatformElementValue_v1"].(map[string]interface{})
	if !ok {
		log.Warn("PlatformElementValue_v1 is not a map")
		return
	}

	byteArray, ok := platformElement["Value"].([]byte)
	if !ok {
		log.Warn("PlatformElementValue_v1[\"Value\"] is not a []byte")
		return
	}
	encoded := base64.StdEncoding.EncodeToString(byteArray)
	a.currentPlatformElementValue = encoded

	if label, err := a.GetElementLabel(); err == nil && label != "" {
		if a.labelTraverseCount == nil {
			a.labelTraverseCount = make(map[string]int)
		}
		current := a.labelTraverseCount[label]
		if direction == DirectionPrevious {
			if current > 1 {
				current = current - 1
			} else {
				current = 1
			}
		} else {
			current = current + 1
		}
		a.labelTraverseCount[label] = current
		a.currentFocusedLabel = label
		log.Infof("label '%s' traverse count=%d", label, current)
	}

	a.ProbeDefaultGeometryAttributes()

	// issues, err := a.RunAudit()
	// if err != nil {
	// 	panic(err)
	// }
	// for _, it := range issues {
	// 	fmt.Printf("type=%s rect=%v font=%v ml=%v fg=%v bg=%v extra=%v\n",
	// 		it.IssueClassificationLabel, it.ElementRect, it.FontSize, it.MLGeneratedDescription,
	// 		it.ForegroundColor, it.BackgroundColor, it.LongDescriptionExtraInfo)
	// }
}

func (a *ControlInterface) Navigate(direction int32) {
	a.Move(direction)
}
