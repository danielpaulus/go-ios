//go:build e2e

package tunnel_test

import (
	"testing"

	"github.com/danielpaulus/go-ios/test/e2e/harness"
)

func TestWebInspectorBrowserControl(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		harness.RunWebInspectorBrowserControl(t, udid)
	})
}
