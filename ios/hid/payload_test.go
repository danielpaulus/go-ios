package hid

import (
	"bytes"
	"testing"

	"github.com/danielpaulus/go-ios/ios/xpc"
)

// roundTrip encodes a payload with the real XPC codec and decodes it again.
//
// This is the check that matters for a wire-protocol port: the encoder rejects
// any Go type it cannot represent, and it maps Go's integer types onto distinct
// XPC types. dtuhidd's Swift decoder is strict about those widths, so a uint64
// silently written as int64 would be rejected on the device where no unit test
// on the map alone would notice.
func roundTrip(t *testing.T, payload map[string]interface{}) map[string]interface{} {
	t.Helper()
	buf := bytes.NewBuffer(nil)
	if err := xpc.EncodeMessage(buf, xpc.Message{
		Flags: xpc.AlwaysSetFlag | xpc.DataFlag,
		Body:  payload,
	}); err != nil {
		t.Fatalf("payload is not XPC-encodable: %v", err)
	}
	msg, err := xpc.DecodeMessage(buf)
	if err != nil {
		t.Fatalf("failed to decode the payload we just encoded: %v", err)
	}
	return msg.Body
}

func dict(t *testing.T, m map[string]interface{}, key string) map[string]interface{} {
	t.Helper()
	v, ok := m[key].(map[string]interface{})
	if !ok {
		t.Fatalf("%q is not a dictionary, got %T", key, m[key])
	}
	return v
}

func TestButtonPayload(t *testing.T) {
	body := roundTrip(t, buildButtonPayload(0x0C, 0x40, ButtonDown))

	if got := body["messageType"]; got != "IndigoButtonEvent" {
		t.Errorf("messageType = %v, want IndigoButtonEvent", got)
	}
	if got := body["featureIdentifier"]; got != buttonFeatureIdentifier {
		t.Errorf("featureIdentifier = %v, want %s", got, buttonFeatureIdentifier)
	}

	payload := dict(t, body, "payload")
	// All three must survive as UInt64 - dtuhidd's decoder rejects Int64 here.
	for key, want := range map[string]uint64{
		"state":     uint64(ButtonDown),
		"usagePage": 0x0C,
		"usageCode": 0x40,
	} {
		got, ok := payload[key].(uint64)
		if !ok {
			t.Errorf("%s is %T, want uint64", key, payload[key])
			continue
		}
		if got != want {
			t.Errorf("%s = %d, want %d", key, got, want)
		}
	}
}

func TestListServicesPayload(t *testing.T) {
	body := roundTrip(t, buildListServicesPayload())

	if got := body["messageType"]; got != "Request" {
		t.Errorf("messageType = %v, want Request", got)
	}
	if got := body["featureIdentifier"]; got != universalFeatureIdentifier {
		t.Errorf("featureIdentifier = %v, want %s", got, universalFeatureIdentifier)
	}
	payload := dict(t, body, "payload")
	if _, ok := payload["connectedServices"].(map[string]interface{}); !ok {
		t.Errorf("connectedServices is %T, want an empty dictionary", payload["connectedServices"])
	}
}

func TestSendReportPayload(t *testing.T) {
	report := BuildTouchscreenReport(TouchContact, 42, 43, goldenTimestamp)
	body := roundTrip(t, buildSendReportPayload(SurfaceMainTouchscreen, report))

	send := dict(t, dict(t, body, "payload"), "send")

	// The report must cross the wire as an XPC data object, byte-for-byte.
	got, ok := send["_0"].([]byte)
	if !ok {
		t.Fatalf("_0 is %T, want []byte encoded as XPC data", send["_0"])
	}
	if !bytes.Equal(got, report) {
		t.Errorf("report was altered in transit\n got %x\nwant %x", got, report)
	}

	serviceID, ok := send["_1"].(uint64)
	if !ok {
		t.Fatalf("_1 is %T, want uint64", send["_1"])
	}
	if serviceID != SurfaceMainTouchscreen {
		t.Errorf("_1 = %d, want %d", serviceID, SurfaceMainTouchscreen)
	}
}

// The create-keyboard payload is the one place where the same logical field takes
// a different integer width depending on where it appears. Getting it wrong fails
// only on the device, so pin the widths here.
func TestCreateKeyboardPayloadIntegerWidths(t *testing.T) {
	body := roundTrip(t, buildCreateKeyboardPayload(SurfaceKeyboardDefault, "test kbd", "go-ios", 0x05AC, 0x0250))

	svc := dict(t, dict(t, dict(t, body, "payload"), "createService"), "_0")

	// Top level: usages are UInt64, IDs are Int64, _ServiceID is UInt64.
	if _, ok := svc["PrimaryUsage"].(uint64); !ok {
		t.Errorf("top-level PrimaryUsage is %T, want uint64", svc["PrimaryUsage"])
	}
	if _, ok := svc["PrimaryUsagePage"].(uint64); !ok {
		t.Errorf("top-level PrimaryUsagePage is %T, want uint64", svc["PrimaryUsagePage"])
	}
	if _, ok := svc["ProductID"].(int64); !ok {
		t.Errorf("top-level ProductID is %T, want int64", svc["ProductID"])
	}
	if _, ok := svc["VendorID"].(int64); !ok {
		t.Errorf("top-level VendorID is %T, want int64", svc["VendorID"])
	}
	if got, ok := svc["_ServiceID"].(uint64); !ok {
		t.Errorf("top-level _ServiceID is %T, want uint64", svc["_ServiceID"])
	} else if got != SurfaceKeyboardDefault {
		t.Errorf("_ServiceID = %#x, want %#x", got, SurfaceKeyboardDefault)
	}

	// Storage block: the same usages are Int64 here, and every leaf is wrapped
	// in a Swift-Codable type envelope.
	storage := dict(t, svc, "_CoreDevice_codablePropertyStorage")

	if _, ok := dict(t, storage, "PrimaryUsage")["int"].(int64); !ok {
		t.Errorf("storage PrimaryUsage is not wrapped as {int: int64}")
	}
	if _, ok := dict(t, storage, "PrimaryUsagePage")["int"].(int64); !ok {
		t.Errorf("storage PrimaryUsagePage is not wrapped as {int: int64}")
	}
	if _, ok := dict(t, storage, "_ServiceID")["uint"].(uint64); !ok {
		t.Errorf("storage _ServiceID is not wrapped as {uint: uint64}")
	}
	if got, ok := dict(t, storage, "UniversalControlVirtualService")["bool"].(bool); !ok || !got {
		t.Errorf("storage UniversalControlVirtualService is not wrapped as {bool: true}")
	}
	if got, ok := dict(t, storage, "Manufacturer")["string"].(string); !ok || got != "go-ios" {
		t.Errorf("storage Manufacturer = %v, want {string: go-ios}", storage["Manufacturer"])
	}
	if got, ok := dict(t, storage, "Transport")["string"].(string); !ok || got != "USB" {
		t.Errorf("storage Transport = %v, want {string: USB}", storage["Transport"])
	}

	// The report descriptor has to arrive as data, unmodified.
	descriptor, ok := dict(t, storage, "ReportDescriptor")["data"].([]byte)
	if !ok {
		t.Fatalf("storage ReportDescriptor is not wrapped as {data: []byte}")
	}
	if !bytes.Equal(descriptor, keyboardReportDescriptor) {
		t.Errorf("report descriptor altered in transit")
	}

	// Fields that belong only in the storage block must not leak to the top level.
	for _, key := range []string{"Manufacturer", "Transport", "ReportDescriptor", "UniversalControlVirtualService"} {
		if _, present := svc[key]; present {
			t.Errorf("%s must live only inside the storage block", key)
		}
	}
}

func TestCreateKeyboardPayloadDeviceUsagePairs(t *testing.T) {
	body := roundTrip(t, buildCreateKeyboardPayload(SurfaceKeyboardDefault, "p", "m", 1, 2))
	svc := dict(t, dict(t, dict(t, body, "payload"), "createService"), "_0")

	// Top-level pairs are a plain array of plain dictionaries.
	pairs, ok := svc["DeviceUsagePairs"].([]interface{})
	if !ok {
		t.Fatalf("top-level DeviceUsagePairs is %T, want []interface{}", svc["DeviceUsagePairs"])
	}
	if len(pairs) != 1 {
		t.Fatalf("top-level DeviceUsagePairs has %d entries, want 1", len(pairs))
	}
	pair, ok := pairs[0].(map[string]interface{})
	if !ok {
		t.Fatalf("pair is %T, want a dictionary", pairs[0])
	}
	if _, ok := pair["DeviceUsage"].(int64); !ok {
		t.Errorf("top-level DeviceUsage is %T, want int64", pair["DeviceUsage"])
	}

	// Storage pairs are the same data wrapped twice: {array: [{dictionary: {...}}]}.
	storage := dict(t, svc, "_CoreDevice_codablePropertyStorage")
	wrapped, ok := dict(t, storage, "DeviceUsagePairs")["array"].([]interface{})
	if !ok {
		t.Fatalf("storage DeviceUsagePairs is not wrapped as {array: [...]}")
	}
	if len(wrapped) != 1 {
		t.Fatalf("storage DeviceUsagePairs has %d entries, want 1", len(wrapped))
	}
	entry, ok := wrapped[0].(map[string]interface{})
	if !ok {
		t.Fatalf("storage pair is %T, want a dictionary", wrapped[0])
	}
	inner, ok := entry["dictionary"].(map[string]interface{})
	if !ok {
		t.Fatalf("storage pair is not wrapped as {dictionary: {...}}")
	}
	if _, ok := dict(t, inner, "DeviceUsage")["int"].(int64); !ok {
		t.Errorf("storage DeviceUsage is not wrapped as {int: int64}")
	}
}
