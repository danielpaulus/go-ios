//go:build e2e

package e2e_test

import (
	"strings"
	"testing"
)

// TestCrashLs lists crash/diagnostic reports. Which reports exist is volatile,
// but iOS devices always accumulate them: assert the list is non-empty, count
// is consistent with the files array, and every entry is a real .ips report.
func TestCrashLs(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		m := smokeObj(t, udid, []string{"files", "length"}, "crash", "ls")
		files, _ := m["files"].([]any)
		length, _ := m["length"].(float64)
		if int(length) != len(files) {
			t.Fatalf("crash ls: length=%d but files has %d entries", int(length), len(files))
		}
		if len(files) == 0 {
			t.Fatalf("crash ls: no crash/diagnostic reports found")
		}
		for _, f := range files {
			if name, _ := f.(string); !strings.HasSuffix(name, ".ips") {
				t.Fatalf("crash ls: unexpected non-.ips entry %q", name)
			}
		}
	})
}
