//go:build e2e

package e2e_test

import "testing"

// boolField asserts the get-command output contains field as a boolean.
func boolField(t *testing.T, udid, field, cmd, sub string) {
	t.Helper()
	m := smokeObj(t, udid, []string{field}, cmd, sub)
	if _, ok := m[field].(bool); !ok {
		t.Fatalf("%s %s: %s is not a bool: %v", cmd, sub, field, m[field])
	}
}

func TestAssistivetouchGet(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { boolField(t, udid, "AssistiveTouchEnabled", "assistivetouch", "get") })
}

func TestVoiceoverGet(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { boolField(t, udid, "VoiceOverTouchEnabled", "voiceover", "get") })
}

func TestZoomGet(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { boolField(t, udid, "ZoomTouchEnabled", "zoom", "get") })
}

func TestTimeformatGet(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		m := smokeObj(t, udid, []string{"TimeFormat"}, "timeformat", "get")
		switch m["TimeFormat"] {
		case "12h", "24h":
		default:
			t.Fatalf("timeformat get: TimeFormat = %v, want 12h or 24h", m["TimeFormat"])
		}
	})
}
