//go:build e2e

package tunnel_test

import (
	"os"
	"path/filepath"
	"testing"
)

// TestScreenshot captures the device screen to a file via the screenshot
// service (a developer service, so it needs Developer Mode + the tunnel/DDI)
// and asserts a non-trivial image was written. screenshot logs to stderr and
// prints nothing to stdout, so success is checked via the output file.
func TestScreenshot(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		out := filepath.Join(t.TempDir(), "screen.png")
		runIOSForDevice(t, udid, "screenshot", "--output="+out)

		fi, err := os.Stat(out)
		if err != nil {
			t.Fatalf("screenshot output %s not created: %v", out, err)
		}
		if fi.Size() < 1024 {
			t.Fatalf("screenshot output %s too small: %d bytes", out, fi.Size())
		}
	})
}
