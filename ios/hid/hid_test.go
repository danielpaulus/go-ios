package hid

import (
	"encoding/binary"
	"testing"
)

func TestBuildTouchscreenReportExactBytes(t *testing.T) {
	// Fixed inputs so every byte is asserted deterministically.
	const (
		x  = uint16(0x1234)
		y  = uint16(0xABCD)
		ts = uint64(0x0000_1122_3344_5566) // fits in 48 bits
	)
	report := buildTouchscreenReport(touchscreenStateContact, x, y, ts)

	if len(report) != 58 {
		t.Fatalf("report length = %d, want 58", len(report))
	}

	want := make([]byte, 58)
	want[0] = 0x09
	want[1] = 0x01
	want[2] = 0x05
	want[3] = 0xC2
	binary.LittleEndian.PutUint16(want[4:6], x)
	binary.LittleEndian.PutUint16(want[6:8], y)
	// want[8:40] zero
	want[40] = 0x02
	// want[41:44] zero
	// 48-bit LE timestamp 0x112233445566 -> 66 55 44 33 22 11
	copy(want[44:50], []byte{0x66, 0x55, 0x44, 0x33, 0x22, 0x11})
	// want[50:58] zero

	for i := range want {
		if report[i] != want[i] {
			t.Errorf("byte %d = 0x%02x, want 0x%02x", i, report[i], want[i])
		}
	}
}

func TestBuildTouchscreenReportReleaseState(t *testing.T) {
	report := buildTouchscreenReport(touchscreenStateRelease, 0, 0, 0)
	if report[3] != 0x02 {
		t.Errorf("release state byte = 0x%02x, want 0x02", report[3])
	}
}

func TestBuildTouchscreenReportTimestampTruncatedTo48Bits(t *testing.T) {
	// High 16 bits must be dropped; low 48 bits preserved.
	report := buildTouchscreenReport(touchscreenStateContact, 0, 0, 0xFFFF_1122_3344_5566)
	got := uint64(report[44]) | uint64(report[45])<<8 | uint64(report[46])<<16 |
		uint64(report[47])<<24 | uint64(report[48])<<32 | uint64(report[49])<<40
	want := uint64(0x1122_3344_5566)
	if got != want {
		t.Errorf("timestamp = 0x%012x, want 0x%012x", got, want)
	}
}

func TestBuildTouchscreenReportCoordinatesLittleEndian(t *testing.T) {
	// Center coordinates: 32768 == 0x8000 -> LE bytes 0x00 0x80.
	report := buildTouchscreenReport(touchscreenStateContact, 32768, 32768, 0)
	if report[4] != 0x00 || report[5] != 0x80 {
		t.Errorf("x bytes = 0x%02x 0x%02x, want 0x00 0x80", report[4], report[5])
	}
	if report[6] != 0x00 || report[7] != 0x80 {
		t.Errorf("y bytes = 0x%02x 0x%02x, want 0x00 0x80", report[6], report[7])
	}
}

func TestFractionToNormalized(t *testing.T) {
	cases := []struct {
		f    float64
		want uint16
	}{
		{-0.5, 0},
		{0, 0},
		{0.5, 32768}, // round(0.5*65535) = round(32767.5) = 32768
		{1, 65535},
		{1.5, 65535},
		{0.25, 16384}, // round(16383.75) = 16384
	}
	for _, c := range cases {
		if got := FractionToNormalized(c.f); got != c.want {
			t.Errorf("FractionToNormalized(%v) = %d, want %d", c.f, got, c.want)
		}
	}
}

// fakeConn records the requests sent so the tap request shape can be asserted
// without a device.
type fakeConn struct {
	sent []map[string]interface{}
}

func (f *fakeConn) Send(data map[string]interface{}, flags ...uint32) error {
	f.sent = append(f.sent, data)
	return nil
}

func (f *fakeConn) ReceiveOnServerClientStream() (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (f *fakeConn) Close() error {
	return nil
}

func TestTapSendsContactThenRelease(t *testing.T) {
	fake := &fakeConn{}
	c := &Connection{conn: fake}
	if err := c.Tap(32768, 32768); err != nil {
		t.Fatalf("Tap returned error: %v", err)
	}
	if len(fake.sent) != 2 {
		t.Fatalf("Tap sent %d reports, want 2 (contact + release)", len(fake.sent))
	}

	for i, req := range fake.sent {
		if req["featureIdentifier"] != featureIdentifier {
			t.Errorf("report %d featureIdentifier = %v, want %s", i, req["featureIdentifier"], featureIdentifier)
		}
		if req["messageType"] != "Request" {
			t.Errorf("report %d messageType = %v, want Request", i, req["messageType"])
		}
		payload, ok := req["payload"].(map[string]interface{})
		if !ok {
			t.Fatalf("report %d payload has wrong type %T", i, req["payload"])
		}
		send, ok := payload["send"].(map[string]interface{})
		if !ok {
			t.Fatalf("report %d payload.send has wrong type %T", i, payload["send"])
		}
		if id, ok := send["_1"].(uint64); !ok || id != MainTouchscreenServiceID {
			t.Errorf("report %d serviceID (_1) = %v, want %d", i, send["_1"], MainTouchscreenServiceID)
		}
		reportBytes, ok := send["_0"].([]byte)
		if !ok {
			t.Fatalf("report %d payload.send._0 has wrong type %T", i, send["_0"])
		}
		if len(reportBytes) != 58 {
			t.Errorf("report %d bytes length = %d, want 58", i, len(reportBytes))
		}
	}

	// First report must be CONTACT (0xC2), second RELEASE (0x02).
	first := fake.sent[0]["payload"].(map[string]interface{})["send"].(map[string]interface{})["_0"].([]byte)
	second := fake.sent[1]["payload"].(map[string]interface{})["send"].(map[string]interface{})["_0"].([]byte)
	if first[3] != 0xC2 {
		t.Errorf("first report state = 0x%02x, want 0xC2 (contact)", first[3])
	}
	if second[3] != 0x02 {
		t.Errorf("second report state = 0x%02x, want 0x02 (release)", second[3])
	}
}
