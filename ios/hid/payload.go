package hid

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
