//go:build e2e

package tunnel_test

import (
	"strconv"
	"testing"
)

// TestInstrumentsNetwork streams network-monitoring samples from the instruments
// network service over the iOS 17+ tunnel for a bounded duration and asserts the
// command self-terminates and emits well-formed samples. Each sample is a
// {"type": <number>, "data": {...}} object (see networkSampleOutput in
// cmd_device_debug.go). The service reports interface and per-connection
// counters; go-ios keeps the decoded payload under "data", whose exact shape
// varies by message type, so we assert the envelope (numeric type + data map)
// rather than a fixed inner schema.
func TestInstrumentsNetwork(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		assertNetworkSamples(t, udid)
	})
}

func assertNetworkSamples(t *testing.T, udid string) {
	t.Helper()
	samples := streamNDJSON(t, udid, instrumentsSampleWindow,
		"instruments", "network", "--duration="+itoa(networkDurationSeconds))

	if len(samples) == 0 {
		t.Fatalf("instruments network: no samples in %ds window", networkDurationSeconds)
	}
	for i, s := range samples {
		if _, ok := s["type"]; !ok {
			t.Fatalf("instruments network sample %d missing \"type\": %v", i, s)
		}
		if _, ok := s["type"].(float64); !ok {
			t.Fatalf("instruments network sample %d \"type\" is not numeric: %T", i, s["type"])
		}
		// "data" is present (may be null for some message types), but when it is
		// an object it must decode as a map — the CLI marshals it as such.
		if d, ok := s["data"]; ok && d != nil {
			if _, ok := d.(map[string]any); !ok {
				t.Fatalf("instruments network sample %d \"data\" is not an object: %T", i, d)
			}
		}
	}
	logSamples(t, "network", udid, samples)
}

// itoa is a tiny local alias so the streaming tests read cleanly.
func itoa(n int) string { return strconv.Itoa(n) }

// logSamples t.Logs a handful of the real decoded samples so `go test -v` shows
// actual on-device data (fps values / network envelopes) in the CI logs.
func logSamples(t *testing.T, kind, udid string, samples []map[string]any) {
	t.Helper()
	const max = 5
	shown := samples
	if len(shown) > max {
		shown = shown[:max]
	}
	t.Logf("instruments %s [%s]: %d samples, first %d:", kind, udid, len(samples), len(shown))
	for i, s := range shown {
		t.Logf("  sample %d: %v", i, s)
	}
}
