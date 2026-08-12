//go:build e2e

package tunnel_test

import "testing"

// File listing over RemoteXPC (iOS 17+), which requires the tunnel. Which files
// exist is volatile, but assert count is consistent with the files array and a
// path is reported.

// fileLs runs `ios file ls <mode>`, asserts the JSON envelope (path/files/count
// are self-consistent) and that every listed entry is a non-empty filename
// string (fileservice.ListDirectory returns a []string of names). It returns the
// decoded file names so callers can make stronger per-domain assertions.
func fileLs(t *testing.T, udid, mode string) []string {
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
	names := make([]string, 0, len(files))
	for i, f := range files {
		name, ok := f.(string)
		if !ok {
			t.Fatalf("file ls %s: entry %d is not a filename string: %T %v", mode, i, f, f)
		}
		if name == "" {
			t.Fatalf("file ls %s: entry %d is an empty filename", mode, i)
		}
		names = append(names, name)
	}
	t.Logf("file ls %s [%s]: %d entries: %v", mode, udid, len(names), names)
	return names
}

// TestFileLsCrash lists the system crash-logs domain root. This domain reliably
// contains entries on a real device (crash .ips files and standard subdirectories
// such as Retired / DiagnosticLogs), so require a non-empty listing — an empty
// crash domain root would be surprising and worth investigating rather than
// silently passing.
func TestFileLsCrash(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		names := fileLs(t, udid, "--crash")
		if len(names) == 0 {
			t.Fatalf("file ls --crash: empty listing for the system crash-logs domain root (unexpected on a real device)")
		}
	})
}

// TestFileLsTemp lists the app temporary domain root. It may legitimately be
// empty, so only the envelope + entry shape (asserted in fileLs) are checked.
func TestFileLsTemp(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { fileLs(t, udid, "--temp") })
}
