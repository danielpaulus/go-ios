package coredevice

import "github.com/google/uuid"

func BuildRequest(deviceId, feature string, input map[string]interface{}) map[string]interface{} {
	u := uuid.New()
	return map[string]interface{}{
		"CoreDevice.CoreDeviceDDIProtocolVersion": int64(0),
		"CoreDevice.action":                       map[string]interface{}{},
		"CoreDevice.coreDeviceVersion": map[string]interface{}{
			"components":              []interface{}{uint64(0x15c), uint64(0x1), uint64(0x0), uint64(0x0), uint64(0x0)},
			"originalComponentsCount": int64(2),
			"stringValue":             "348.1",
		},
		"CoreDevice.deviceIdentifier":     deviceId,
		"CoreDevice.featureIdentifier":    feature,
		"CoreDevice.input":                input,
		"CoreDevice.invocationIdentifier": u.String(),
	}
}

// BuildRequestWithAction builds a CoreDevice request that names the action to
// invoke, which some services require in addition to the feature identifier.
//
// It also advertises a newer DDI protocol version and coreDeviceVersion than
// BuildRequest: those are the values devicectl sends, and services like
// com.apple.coredevice.displayservice reject the older envelope. Existing
// callers of BuildRequest are left on the envelope they were verified against.
func BuildRequestWithAction(deviceId, feature, action string, input map[string]interface{}) map[string]interface{} {
	u := uuid.New()
	return map[string]interface{}{
		"CoreDevice.CoreDeviceDDIProtocolVersion": int64(2),
		"CoreDevice.action":                       map[string]interface{}{},
		"CoreDevice.actionIdentifier":             action,
		"CoreDevice.coreDeviceVersion": map[string]interface{}{
			"components":              []interface{}{uint64(629), uint64(3)},
			"originalComponentsCount": int64(2),
			"stringValue":             "629.3",
		},
		"CoreDevice.deviceIdentifier":     deviceId,
		"CoreDevice.featureIdentifier":    feature,
		"CoreDevice.input":                input,
		"CoreDevice.invocationIdentifier": u.String(),
	}
}
