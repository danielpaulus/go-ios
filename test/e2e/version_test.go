//go:build e2e

package e2e_test

import (
	"bytes"
	"testing"
)

// TestVersion does not target a device: it just verifies the built binary
// runs and prints its version.
func TestVersion(t *testing.T) {
	if out := bytes.TrimSpace(runIOS(t, "version")); len(out) == 0 {
		t.Fatal("ios version: empty output")
	}
}
