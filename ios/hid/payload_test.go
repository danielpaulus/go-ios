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

func TestSendReportPayload(t *testing.T) {
	report := buildTouchscreenReport(touchContact, 42, 43, goldenTimestamp)
	body := roundTrip(t, buildSendReportPayload(surfaceMainTouchscreen, report))

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
	if serviceID != surfaceMainTouchscreen {
		t.Errorf("_1 = %d, want %d", serviceID, surfaceMainTouchscreen)
	}
}
