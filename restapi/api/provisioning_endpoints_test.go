package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/danielpaulus/go-ios/ios/mcinstall"
)

// TestGetPrepareSkipOptions verifies the device-free skip-options endpoint: it
// returns the full mcinstall option list as JSON with a matching count. This
// runs end-to-end because it needs no device.
func TestGetPrepareSkipOptions(t *testing.T) {
	w, c := newHandlerCtx("GET", "/prepare/skip-options", "")
	GetPrepareSkipOptions(c)

	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200 (body=%s)", w.Code, w.Body.String())
	}

	var resp struct {
		Options []string `json:"options"`
		Count   int      `json:"count"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response not valid JSON: %v (body=%s)", err, w.Body.String())
	}

	want := mcinstall.GetAllSetupSkipOptions()
	if resp.Count != len(want) {
		t.Fatalf("count = %d, want %d", resp.Count, len(want))
	}
	if len(resp.Options) != len(want) {
		t.Fatalf("options len = %d, want %d", len(resp.Options), len(want))
	}
	// Spot-check a couple of well-known options survive the round trip.
	found := map[string]bool{}
	for _, o := range resp.Options {
		found[o] = true
	}
	for _, expected := range []string{"WiFi", "Passcode"} {
		if !found[expected] {
			t.Fatalf("expected skip option %q missing from response", expected)
		}
	}
}
