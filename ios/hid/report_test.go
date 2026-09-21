package hid

import (
	"encoding/hex"
	"testing"
)

// goldenTimestamp is an arbitrary 48-bit value with distinct bytes, so a
// mis-ordered or mis-sized timestamp field shows up immediately.
const goldenTimestamp uint64 = 0x123456789ABC

func TestBuildTouchscreenReport(t *testing.T) {
	tests := []struct {
		name  string
		state touchState
		x, y  uint16
		want  string
	}{
		{
			name:  "contact",
			state: touchContact,
			x:     500,
			y:     1000,
			want: "090105c2f401e803000000000000000000000000000000000000000000000000" +
				"000000000000000002000000bc9a785634120000000000000000",
		},
		{
			name:  "release",
			state: touchRelease,
			x:     500,
			y:     1000,
			want: "09010502f401e803000000000000000000000000000000000000000000000000" +
				"000000000000000002000000bc9a785634120000000000000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildTouchscreenReport(tt.state, tt.x, tt.y, goldenTimestamp)
			if len(got) != touchscreenReportLen {
				t.Errorf("length = %d, want %d", len(got), touchscreenReportLen)
			}
			if hex.EncodeToString(got) != tt.want {
				t.Errorf("report mismatch\n got %s\nwant %s", hex.EncodeToString(got), tt.want)
			}
		})
	}
}

// The timestamp field is six bytes wide, so a value that does not fit has to be
// truncated to its low 48 bits instead of corrupting the trailing reserved bytes.
func TestTimestampIsTruncatedToFieldWidth(t *testing.T) {
	got := buildTouchscreenReport(touchContact, 0, 0, 0xFFFFFFFFFFFFFFFF)
	if len(got) != touchscreenReportLen {
		t.Fatalf("length = %d, want %d", len(got), touchscreenReportLen)
	}
	if ts := hex.EncodeToString(got[44:50]); ts != "ffffffffffff" {
		t.Errorf("timestamp = %s, want ffffffffffff", ts)
	}
	// The bytes after the field must be untouched by the overflow.
	if tail := hex.EncodeToString(got[50:]); tail != "0000000000000000" {
		t.Errorf("overflow corrupted the trailing bytes: %s", tail)
	}
}
