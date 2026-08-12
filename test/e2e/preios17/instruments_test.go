//go:build e2e

// These commands are served by the instruments/DTX services. On iOS 17+ they
// require the tunnel (and live in test/e2e/tunnel); on pre-iOS17 devices they
// work directly over usbmuxd with only the Developer Disk Image mounted (done
// in TestMain), which is exactly what this suite proves.
package preios17_test

import (
	"os"
	"path/filepath"
	"testing"
)

// TestPs lists running processes via the instruments device-info service.
func TestPs(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		smokeArrResilient(t, udid, []string{"Name", "Pid"}, "ps")
	})
}

// TestScreenshot captures a screenshot to a file and asserts it is a real,
// non-empty PNG.
func TestScreenshot(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		out := filepath.Join(t.TempDir(), "shot.png")
		runIOSForDeviceResilient(t, udid, "screenshot", "--output="+out)

		b, err := os.ReadFile(out)
		if err != nil {
			t.Fatalf("screenshot: reading %s: %v", out, err)
		}
		if len(b) < 1024 {
			t.Fatalf("screenshot: %s is only %d bytes, expected a real image", out, len(b))
		}
		if string(b[:8]) != "\x89PNG\r\n\x1a\n" {
			t.Fatalf("screenshot: %s is not a PNG (magic %x)", out, b[:8])
		}
	})
}

// TestLaunchKill launches a known system app via the instruments process-control
// service and then kills it. Both report success (non-zero exit fails the test).
func TestLaunchKill(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		runIOSForDeviceResilient(t, udid, "launch", knownSystemApp)
		runIOSForDeviceResilient(t, udid, "kill", knownSystemApp)
	})
}
