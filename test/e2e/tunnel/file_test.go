//go:build e2e

package tunnel_test

import "testing"

// File listing over RemoteXPC (iOS 17+), which requires the tunnel.

func TestFileLsCrash(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smokeJSON(t, udid, "file", "ls", "--crash") })
}

func TestFileLsTemp(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smokeJSON(t, udid, "file", "ls", "--temp") })
}
