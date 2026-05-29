//go:build e2e

package tunnel_test

import (
	"path/filepath"
	"testing"
)

// These exercise the developer disk image flow over the iOS 17+ tunnel. They
// need the tunnel running on the host but not Developer Mode (mounting works
// without it; only the developer services do not).

// TestImageAuto downloads the DDI from deviceboxhq into a fresh dir and mounts
// it, covering the full image-auto path end to end. These commands log to
// stderr and print nothing to stdout, so success is asserted via a clean exit
// (runIOSForDevice fails the test on a non-zero exit).
func TestImageAuto(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		runIOSForDevice(t, udid, "image", "auto", "--basedir="+t.TempDir())
	})
}

// TestImageMount mounts the DDI from a local path: it first uses image auto to
// fetch the image into a temp dir, then unmounts and mounts it explicitly via
// --path.
func TestImageMount(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		dir := t.TempDir()
		runIOSForDevice(t, udid, "image", "auto", "--basedir="+dir)

		restores, err := filepath.Glob(filepath.Join(dir, "*", "Restore"))
		if err != nil || len(restores) == 0 {
			t.Fatalf("no downloaded Restore dir under %s: %v", dir, err)
		}

		runIOSForDevice(t, udid, "image", "unmount")
		runIOSForDevice(t, udid, "image", "mount", "--path="+restores[0])
	})
}
