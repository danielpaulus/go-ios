package hid

// Declares the surface as a keyboard, which is what hides the on-screen one. Our
// reports do not match this layout, and the device tolerates that.
var keyboardReportDescriptor = []byte{
	0x05, 0x01, // Usage Page (Generic Desktop)
	0x09, 0x06, // Usage (Keyboard)
	0xA1, 0x01, // Collection (Application)
	0x05, 0x07, // Usage Page (Keyboard)
	0x19, 0xE0, // Usage Minimum (0xE0)
	0x29, 0xE7, // Usage Maximum (0xE7) - the 8 modifier keys
	0x15, 0x00, // Logical Minimum (0)
	0x25, 0x01, // Logical Maximum (1)
	0x95, 0x08, // Report Count (8)
	0x75, 0x01, // Report Size (1)
	0x81, 0x02, // Input (Data, Variable, Absolute) - modifier byte
	0x95, 0x01, // Report Count (1)
	0x75, 0x08, // Report Size (8)
	0x81, 0x01, // Input (Constant) - reserved byte
	0x05, 0x07, // Usage Page (Keyboard)
	0x19, 0x00, // Usage Minimum (0)
	0x29, 0xFF, // Usage Maximum (0xFF)
	0x15, 0x00, // Logical Minimum (0)
	0x26, 0xFF, 0x00, // Logical Maximum (255)
	0x95, 0x06, // Report Count (6)
	0x75, 0x08, // Report Size (8)
	0x81, 0x00, // Input (Data, Array) - 6-key array
	0x05, 0x08, // Usage Page (LED)
	0x19, 0x01, // Usage Minimum (1)
	0x29, 0x05, // Usage Maximum (5)
	0x15, 0x00, // Logical Minimum (0)
	0x25, 0x01, // Logical Maximum (1)
	0x95, 0x05, // Report Count (5)
	0x75, 0x01, // Report Size (1)
	0x91, 0x02, // Output (Data, Variable, Absolute) - LEDs
	0x95, 0x01, // Report Count (1)
	0x75, 0x03, // Report Size (3)
	0x91, 0x01, // Output (Constant) - padding
	0xC0, // End Collection
}

func buildButtonPayload(usagePage, usageCode uint64, state ButtonState) map[string]interface{} {
	return map[string]interface{}{
		"messageType": "IndigoButtonEvent",
		"payload": map[string]interface{}{
			"state":     uint64(state),
			"usagePage": usagePage,
			"usageCode": usageCode,
		},
		"featureIdentifier": buttonFeatureIdentifier,
	}
}

// buildListServicesPayload builds the request enumerating registered HID surfaces.
func buildListServicesPayload() map[string]interface{} {
	return map[string]interface{}{
		"featureIdentifier": universalFeatureIdentifier,
		"messageType":       "Request",
		"payload": map[string]interface{}{
			"connectedServices": map[string]interface{}{},
		},
	}
}

// buildSendReportPayload builds the request posting a raw HID report to a surface.
// The report goes over the wire as an XPC data object and the surface ID as UInt64.
func buildSendReportPayload(serviceID uint64, report []byte) map[string]interface{} {
	return map[string]interface{}{
		"featureIdentifier": universalFeatureIdentifier,
		"messageType":       "Request",
		"payload": map[string]interface{}{
			"send": map[string]interface{}{
				"_0": report,
				"_1": serviceID,
			},
		},
	}
}

// Registers a virtual keyboard. Integer widths differ by position, and every leaf
// inside the storage block must carry its type, or the request is rejected.
func buildCreateKeyboardPayload(serviceID uint64, product, manufacturer string, vendorID, productID int64) map[string]interface{} {
	const (
		usagePageGenericDesktop int64 = 1
		usageKeyboard           int64 = 6
	)

	storage := map[string]interface{}{
		"Manufacturer":     map[string]interface{}{"string": manufacturer},
		"Product":          map[string]interface{}{"string": product},
		"ProductID":        map[string]interface{}{"int": productID},
		"VendorID":         map[string]interface{}{"int": vendorID},
		"PrimaryUsage":     map[string]interface{}{"int": usageKeyboard},
		"PrimaryUsagePage": map[string]interface{}{"int": usagePageGenericDesktop},
		"DeviceUsagePairs": map[string]interface{}{
			"array": []interface{}{
				map[string]interface{}{
					"dictionary": map[string]interface{}{
						"DeviceUsage":     map[string]interface{}{"int": usageKeyboard},
						"DeviceUsagePage": map[string]interface{}{"int": usagePageGenericDesktop},
					},
				},
			},
		},
		"Transport":                      map[string]interface{}{"string": "USB"},
		"ReportDescriptor":               map[string]interface{}{"data": keyboardReportDescriptor},
		"UniversalControlVirtualService": map[string]interface{}{"bool": true},
		"_ServiceID":                     map[string]interface{}{"uint": serviceID},
	}

	return map[string]interface{}{
		"featureIdentifier": universalFeatureIdentifier,
		"messageType":       "Request",
		"payload": map[string]interface{}{
			"createService": map[string]interface{}{
				"_0": map[string]interface{}{
					"DeviceUsagePairs": []interface{}{
						map[string]interface{}{
							"DeviceUsage":     usageKeyboard,
							"DeviceUsagePage": usagePageGenericDesktop,
						},
					},
					"PrimaryUsage":                       uint64(usageKeyboard),
					"PrimaryUsagePage":                   uint64(usagePageGenericDesktop),
					"Product":                            product,
					"ProductID":                          productID,
					"VendorID":                           vendorID,
					"_CoreDevice_codablePropertyStorage": storage,
					"_ServiceID":                         serviceID,
				},
			},
		},
	}
}
