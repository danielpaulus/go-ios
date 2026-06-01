//go:build e2e

package tunnel_test

import "testing"

// File listing over RemoteXPC (iOS 17+), which requires the tunnel. Which files
// exist is volatile, so assert the response shape.

func TestFileLsCrash(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		smokeObj(t, udid, []string{"count", "files", "path"}, "file", "ls", "--crash")
	})
}

func TestFileLsTemp(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		smokeObj(t, udid, []string{"count", "files", "path"}, "file", "ls", "--temp")
	})
}
