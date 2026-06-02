//go:build e2e

package tunnel_test

import "testing"

// File listing over RemoteXPC (iOS 17+), which requires the tunnel. Which files
// exist is volatile, but assert count is consistent with the files array and a
// path is reported.

func fileLs(t *testing.T, udid, mode string) {
	t.Helper()
	m := smokeObj(t, udid, []string{"count", "files", "path"}, "file", "ls", mode)
	files, _ := m["files"].([]any)
	count, _ := m["count"].(float64)
	if int(count) != len(files) {
		t.Fatalf("file ls %s: count=%d but files has %d entries", mode, int(count), len(files))
	}
	if p, _ := m["path"].(string); p == "" {
		t.Fatalf("file ls %s: empty path", mode)
	}
}

func TestFileLsCrash(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { fileLs(t, udid, "--crash") })
}

func TestFileLsTemp(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { fileLs(t, udid, "--temp") })
}
