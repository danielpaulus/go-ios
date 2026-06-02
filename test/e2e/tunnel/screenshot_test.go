//go:build e2e

package tunnel_test

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// TestScreenshot captures the device screen to a file via the screenshot
// service (a developer service, so it needs Developer Mode + the tunnel/DDI)
// and decodes it with image/png to assert it is a real PNG with sane
// dimensions, not just a non-empty file.
func TestScreenshot(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		out := filepath.Join(t.TempDir(), "screen.png")
		runIOSForDevice(t, udid, "screenshot", "--output="+out)

		f, err := os.Open(out)
		if err != nil {
			t.Fatalf("screenshot: output %s not created: %v", out, err)
		}
		defer f.Close()

		// png.Decode validates the PNG signature and decodes the image data;
		// it fails if the bytes are not a valid PNG.
		img, err := png.Decode(f)
		if err != nil {
			t.Fatalf("screenshot: %s is not a valid PNG: %v", out, err)
		}
		if b := img.Bounds(); b.Dx() <= 0 || b.Dy() <= 0 {
			t.Fatalf("screenshot: PNG has invalid dimensions %dx%d", b.Dx(), b.Dy())
		}
	})
}
