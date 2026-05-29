//go:build e2e

package tunnel_test

import (
	"testing"
	"time"

	"github.com/danielpaulus/go-ios/test/e2e/harness"
)

// Streaming commands run until killed, so they are smoke-tested by letting them
// stream for a short window and asserting they produced output.
const streamWindow = 6 * time.Second

func TestSyslog(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { harness.StreamSmoke(t, udid, streamWindow, "syslog") })
}

func TestOstrace(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { harness.StreamSmoke(t, udid, streamWindow, "ostrace") })
}
