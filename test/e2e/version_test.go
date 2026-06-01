//go:build e2e

package e2e_test

import (
	"encoding/json"
	"testing"
)

// TestVersion does not target a device: it verifies the built binary runs and
// prints a JSON object with a non-empty version string.
func TestVersion(t *testing.T) {
	var m map[string]any
	if err := json.Unmarshal(runIOS(t, "version"), &m); err != nil {
		t.Fatalf("version: not JSON: %v", err)
	}
	if v, _ := m["version"].(string); v == "" {
		t.Fatalf("version: missing/empty version field in %v", m)
	}
}
