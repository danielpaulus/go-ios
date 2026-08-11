//go:build e2e

package tunnel_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/danielpaulus/go-ios/test/e2e/harness"
)

const pasteboardServiceName = "com.apple.coredevice.pasteboardservice"

// skipUnlessPasteboardService skips the test when the device's RSD service
// list does not advertise the pasteboard service. The service ships with the
// Developer Disk Image (first seen in DDI 27A5194q), not with iOS itself, so
// devices with an older DDI mounted simply don't have it.
func skipUnlessPasteboardService(t *testing.T, udid string) {
	t.Helper()
	out := runIOSForDevice(t, udid, "rsd", "ls")
	var services map[string]any
	if err := json.Unmarshal(out, &services); err != nil {
		t.Fatalf("rsd ls: invalid json: %v\noutput: %s", err, out)
	}
	if _, ok := services[pasteboardServiceName]; !ok {
		t.Skipf("%s is not in the RSD service list; mount a newer DDI to enable it", pasteboardServiceName)
	}
}

func TestPasteboardRoundtrip(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		skipUnlessPasteboardService(t, udid)

		const text = "go-ios e2e pasteboard ✓ emöji 🎉"
		runIOSForDevice(t, udid, "pasteboard", "set", text)
		out := runIOSForDevice(t, udid, "pasteboard", "get")
		if got := strings.TrimSuffix(string(out), "\n"); got != text {
			t.Fatalf("pasteboard get: expected %q, got %q", text, got)
		}
	})
}

func TestPasteboardSetFromStdin(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		skipUnlessPasteboardService(t, udid)

		const text = "pasteboard-from-stdin"
		harness.RunForDeviceWithStdin(t, udid, strings.NewReader(text), "pasteboard", "set")
		out := runIOSForDevice(t, udid, "pasteboard", "get")
		if got := strings.TrimSuffix(string(out), "\n"); got != text {
			t.Fatalf("pasteboard get after stdin set: expected %q, got %q", text, got)
		}
	})
}

// TestPasteboardSetExplicitEmpty guards the docopt regression where an
// explicit empty <text> argument fell into the stdin branch: with data waiting
// on stdin and `set ""`, the pasteboard must end up empty — the stdin content
// must not be consumed.
func TestPasteboardSetExplicitEmpty(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		skipUnlessPasteboardService(t, udid)

		harness.RunForDeviceWithStdin(t, udid, strings.NewReader("MUST-NOT-APPEAR"), "pasteboard", "set", "")
		out := runIOSForDevice(t, udid, "pasteboard", "get")
		if got := strings.TrimSpace(string(out)); got != "" {
			t.Fatalf("expected empty pasteboard after `set \"\"`, got %q", got)
		}
	})
}
