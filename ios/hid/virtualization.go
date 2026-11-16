package hid

import (
	"encoding/binary"
	"fmt"
	"math"
)

// A foldable's hinge is a virtualization control, set with a vendor defined
// HID event whose data names the control and its new value.
const (
	vendorDefinedFeatureIdentifier = "com.apple.coredevice.feature.remote.hid.vendordefined"
	vendorDefinedMessageType       = "IndigoVendorDefinedEvent"

	virtualizationUsagePage uint64 = 0xff61
	virtualizationUsage     uint64 = 0x5b

	virtualizationProvider = "com.apple.Virtualization.VirtualMachines"

	hingeControl = "hinge-slider-control"
)

// Hinge angles in degrees, as the device hub's fold buttons set them.
const (
	HingeClosed   = 0.0
	HingeHalfOpen = 130.0
	HingeOpenFlat = 180.0
)

// SetHingeAngle folds the device to degrees, from 0 (closed) to 180 (open
// flat). The display in use follows about half a second later: on an iPhone
// Duo the cover display at 0 and 45 degrees, the inner one at 130 and 180.
func (c *Controls) SetHingeAngle(degrees float64) error {
	if !c.foldable {
		return fmt.Errorf("SetHingeAngle: %w", ErrNotFoldable)
	}
	if math.IsNaN(degrees) || degrees < HingeClosed || degrees > HingeOpenFlat {
		return fmt.Errorf("SetHingeAngle: %v degrees is outside 0 to 180", degrees)
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if err := c.send(buildVirtualizationPayload(hingeControl, "range", degrees)); err != nil {
		return fmt.Errorf("SetHingeAngle: %w", err)
	}
	return nil
}

// buildVirtualizationPayload sets one virtualization control to a value, for a
// range control such as the hinge.
func buildVirtualizationPayload(control, kind string, value float64) map[string]interface{} {
	return map[string]interface{}{
		"featureIdentifier": vendorDefinedFeatureIdentifier,
		"messageType":       vendorDefinedMessageType,
		"payload": map[string]interface{}{
			"usagePage": virtualizationUsagePage,
			"usage":     virtualizationUsage,
			"version":   uint64(0),
			"data":      encodeVirtualizationControl(control, kind, value),
		},
	}
}

// The data is a small serialized dictionary: a header, then a 4 byte header
// for every key and value with the length in its low 3 bytes and the type in
// the top one, each padded to 4 bytes. The last value's type has 0x80 set.
const (
	virtualizationKey    byte = 0x08 // a string with its NUL
	virtualizationString byte = 0x09 // a string without one
	virtualizationDouble byte = 0x04 // 8 bytes, little endian
	virtualizationLast   byte = 0x80

	// A double's header holds no length.
	virtualizationDoubleHeader = 0x3f
)

var virtualizationHeader = []byte{0xd3, 0, 0, 0, 4, 0, 0, 0x81}

func encodeVirtualizationControl(control, kind string, value float64) []byte {
	b := append([]byte{}, virtualizationHeader...)
	b = appendVirtualizationString(b, virtualizationKey, "provider")
	b = appendVirtualizationString(b, virtualizationString, virtualizationProvider)
	b = appendVirtualizationString(b, virtualizationKey, "source")
	b = appendVirtualizationString(b, virtualizationString, control)
	b = appendVirtualizationString(b, virtualizationKey, "type")
	b = appendVirtualizationString(b, virtualizationString, kind)
	b = appendVirtualizationString(b, virtualizationKey, "value")
	b = append(b, virtualizationDoubleHeader, 0, 0, virtualizationDouble|virtualizationLast)
	return binary.LittleEndian.AppendUint64(b, math.Float64bits(value))
}

func appendVirtualizationString(b []byte, typ byte, s string) []byte {
	body := []byte(s)
	if typ == virtualizationKey {
		body = append(body, 0)
	}
	n := len(body)
	b = append(b, byte(n), byte(n>>8), byte(n>>16), typ)
	b = append(b, body...)
	for len(b)%4 != 0 {
		b = append(b, 0)
	}
	return b
}
