//go:build e2e

package e2e_test

import (
	"encoding/json"
	"testing"
)

// toggleRoundTrip reads the boolean state of an accessibility setting via
// "<cmd> get", flips it with enable/disable, asserts it changed, and always
// restores the original value.
func toggleRoundTrip(t *testing.T, udid, cmd, field string) {
	t.Helper()
	get := func() bool {
		var m map[string]any
		if err := json.Unmarshal(runIOSForDevice(t, udid, cmd, "get"), &m); err != nil {
			t.Fatalf("%s get: parse: %v", cmd, err)
		}
		b, ok := m[field].(bool)
		if !ok {
			t.Fatalf("%s get: missing bool field %q", cmd, field)
		}
		return b
	}
	set := func(on bool) {
		op := "disable"
		if on {
			op = "enable"
		}
		runIOSForDevice(t, udid, cmd, op)
	}

	orig := get()
	defer set(orig) // always restore the original state
	set(!orig)
	if get() == orig {
		t.Fatalf("%s: state did not change after setting %v", cmd, !orig)
	}
}

func TestAssistivetouchToggle(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		toggleRoundTrip(t, udid, "assistivetouch", "AssistiveTouchEnabled")
	})
}

func TestVoiceoverToggle(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		toggleRoundTrip(t, udid, "voiceover", "VoiceOverTouchEnabled")
	})
}

func TestZoomToggle(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		toggleRoundTrip(t, udid, "zoom", "ZoomTouchEnabled")
	})
}

// TestTimeformatRoundTrip sets the time format to the opposite of the current
// value, asserts the change, and restores the original.
func TestTimeformatRoundTrip(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		get := func() string {
			var m map[string]string
			if err := json.Unmarshal(runIOSForDevice(t, udid, "timeformat", "get"), &m); err != nil {
				t.Fatalf("timeformat get: parse: %v", err)
			}
			return m["TimeFormat"]
		}

		orig := get()
		other := "24h"
		if orig == "24h" {
			other = "12h"
		}
		defer runIOSForDevice(t, udid, "timeformat", orig) // restore
		runIOSForDevice(t, udid, "timeformat", other)
		if got := get(); got != other {
			t.Fatalf("timeformat: set %s but got %s", other, got)
		}
	})
}
