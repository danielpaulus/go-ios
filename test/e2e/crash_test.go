//go:build e2e

package e2e_test

import (
	"strings"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
)

// TestCrashLs lists crash/diagnostic reports. Which reports exist is volatile,
// but iOS devices always accumulate them: assert the list is non-empty and the
// count is consistent with the files array. Before iOS 26 every entry is a real
// .ips report, so assert that exactly. iOS 26 began surfacing non-.ips artifacts
// in the same directory — subdirectories like BioLog/ and retired reports such
// as Retired/<name>.ips.ca.synced — so there require only that at least one real
// .ips report is present rather than that every entry is one.
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

		if deviceVersion(t, udid).LessThan(ios.IOS26()) {
			for _, f := range files {
				if name, _ := f.(string); !strings.HasSuffix(name, ".ips") {
					t.Fatalf("crash ls: unexpected non-.ips entry %q", name)
				}
			}
			return
		}

		hasIPS := false
		for _, f := range files {
			if name, _ := f.(string); strings.HasSuffix(name, ".ips") {
				hasIPS = true
				break
			}
		}
		if !hasIPS {
			t.Fatalf("crash ls: no .ips report among %d entries: %v", len(files), files)
		}
	})
}
