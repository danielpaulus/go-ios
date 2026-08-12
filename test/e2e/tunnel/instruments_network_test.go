//go:build e2e

package tunnel_test

import (
	"strconv"
	"testing"
)

// The instruments network-monitoring service only emits samples when the device
// has actual network activity; a quiescent CI device can legitimately produce
// zero samples in a short window (go-ios's own README notes "open apple maps to
// force network traffic"). Driving traffic by hammering app launches over the
// same instruments/DTX transport that carries the network stream destabilises
// that shared channel (launch/kill time out), so we do NOT do that. Instead this
// test pins the reliable contract: the command connects to the network service,
// streams for its bounded --duration, self-terminates cleanly, and every sample
// it DOES emit is a well-formed {"type": <number>, "data": ...} envelope. The
// sample count is logged (real on-device data when traffic happens to occur) but
// not required, mirroring how the suite treats other volatile streamed output.

const networkStreamSeconds = 8

func TestInstrumentsNetwork(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		samples := streamNDJSON(t, udid, instrumentsSampleWindow,
			"instruments", "network", "--duration="+itoa(networkStreamSeconds))

		for i, s := range samples {
			if _, ok := s["type"].(float64); !ok {
				t.Fatalf("instruments network sample %d \"type\" missing or not numeric: %v", i, s)
			}
			if d, ok := s["data"]; ok && d != nil {
				if _, ok := d.(map[string]any); !ok {
					t.Fatalf("instruments network sample %d \"data\" is not an object: %T", i, d)
				}
			}
		}
		logSamples(t, "network", udid, samples)
	})
}

// itoa is a tiny local alias so the streaming tests read cleanly.
func itoa(n int) string { return strconv.Itoa(n) }

// logSamples t.Logs a handful of the real decoded samples so `go test -v` shows
// actual on-device data in the CI logs.
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
