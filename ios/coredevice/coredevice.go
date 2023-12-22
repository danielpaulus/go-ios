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
