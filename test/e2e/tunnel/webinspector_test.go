//go:build e2e

package tunnel_test

import (
	"encoding/json"
	"testing"

	"github.com/danielpaulus/go-ios/test/e2e/harness"
)

func TestWebInspectorBrowserControl(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		harness.RunWebInspectorBrowserControl(t, udid)

		// CLI-level guard for `ios webinspector list`: it shares the "list" arg
		// with the global `ios list` (see isDeviceListCommand in cmd_global.go),
		// so a dispatch regression would silently return the device list instead
		// of inspectable pages. Run it here, after the browser-control session
		// above has closed, so there is only ever one webinspector connection at
		// a time. Assert it reaches the webinspector handler — a JSON array of
		// pages, which may legitimately be empty — not the device-list object.
		out := runIOSForDevice(t, udid, "webinspector", "list", "--timeout=20")
		var pages []any
		if err := json.Unmarshal(out, &pages); err != nil {
			t.Fatalf("webinspector list: output is not a JSON array of pages: %v\n%s", err, out)
		}
	})
}
