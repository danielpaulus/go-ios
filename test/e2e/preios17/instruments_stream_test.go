//go:build e2e

// Streaming instruments/DTX commands on pre-iOS17 devices: these reach the
// graphics (FPS) and network-monitoring DTX services over usbmuxd with only a
// mounted Developer Disk Image (done in TestMain) and NO tunnel. The tunnel
// suite proves the same commands over the iOS 17+ tunnel; this proves them on
// the classic transport.
package preios17_test

import (
	"strconv"
	"testing"
	"time"
)

// instrumentsSampleWindow bounds a --duration=<n> streaming run: the command
// must self-terminate within it (slack over the sampling duration for DTX
// connection setup and teardown).
const (
	fpsDurationSeconds      = 5
	networkDurationSeconds  = 8
	instrumentsSampleWindow = 30 * time.Second
)

// TestInstrumentsFPS streams frames-per-second samples from the graphics OpenGL
// service and asserts the command self-terminates and emits well-formed
// {"fps": <number>} samples (an idle screen still emits CoreAnimation frames).
func TestInstrumentsFPS(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		samples := streamNDJSON(t, udid, instrumentsSampleWindow,
			"instruments", "fps", "--duration="+strconv.Itoa(fpsDurationSeconds))
		if len(samples) == 0 {
			t.Fatalf("instruments fps: no samples in %ds window", fpsDurationSeconds)
		}
		for i, s := range samples {
			v, ok := s["fps"]
			if !ok {
				t.Fatalf("instruments fps sample %d missing \"fps\": %v", i, s)
			}
			f, ok := v.(float64)
			if !ok {
				t.Fatalf("instruments fps sample %d \"fps\" not numeric: %T %v", i, v, v)
			}
			if f < 0 {
				t.Fatalf("instruments fps sample %d negative fps: %v", i, f)
			}
		}
		logInstrumentsSamples(t, "fps", udid, samples)
	})
}

// TestInstrumentsNetwork streams network-monitoring samples and asserts the
// command self-terminates and that every sample it emits is a well-formed
// {"type": <number>, "data": ...} envelope. The network service only reports
// samples when the device has real traffic, so a quiescent CI device can emit
// zero; the count is logged, not required (see the tunnel suite's network test
// for why traffic is not force-generated here).
func TestInstrumentsNetwork(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		samples := streamNDJSON(t, udid, instrumentsSampleWindow,
			"instruments", "network", "--duration="+strconv.Itoa(networkDurationSeconds))
		for i, s := range samples {
			if _, ok := s["type"].(float64); !ok {
				t.Fatalf("instruments network sample %d \"type\" missing or not numeric: %v", i, s)
			}
			if d, ok := s["data"]; ok && d != nil {
				if _, ok := d.(map[string]any); !ok {
					t.Fatalf("instruments network sample %d \"data\" not an object: %T", i, d)
				}
			}
		}
		logInstrumentsSamples(t, "network", udid, samples)
	})
}

func logInstrumentsSamples(t *testing.T, kind, udid string, samples []map[string]any) {
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
