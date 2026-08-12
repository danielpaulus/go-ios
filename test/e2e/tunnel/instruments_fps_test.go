//go:build e2e

package tunnel_test

import (
	"testing"
	"time"
)

// instrumentsSampleWindow bounds a --duration=<fpsDurationSeconds> streaming
// run: the command must terminate on its own within this window (a few seconds
// of slack over the sampling duration for connection setup and teardown).
const (
	fpsDurationSeconds      = 5
	networkDurationSeconds  = 5
	instrumentsSampleWindow = 30 * time.Second
)

// TestInstrumentsFPS streams frames-per-second samples from the instruments
// graphics (OpenGL) service over the iOS 17+ tunnel for a bounded duration and
// asserts the command self-terminates and emits well-formed {"fps": <number>}
// samples. Even an idle screen still emits CoreAnimation frame samples, so at
// least one sample must arrive within the window.
func TestInstrumentsFPS(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		assertFPSSamples(t, udid)
	})
}

func assertFPSSamples(t *testing.T, udid string) {
	t.Helper()
	samples := streamNDJSON(t, udid, instrumentsSampleWindow,
		"instruments", "fps", "--duration="+itoa(fpsDurationSeconds))

	if len(samples) == 0 {
		t.Fatalf("instruments fps: no samples in %ds window (idle screen still emits frames)", fpsDurationSeconds)
	}
	for i, s := range samples {
		v, ok := s["fps"]
		if !ok {
			t.Fatalf("instruments fps sample %d missing \"fps\" field: %v", i, s)
		}
		f, ok := v.(float64)
		if !ok {
			t.Fatalf("instruments fps sample %d \"fps\" is not numeric: %T %v", i, v, v)
		}
		if f < 0 {
			t.Fatalf("instruments fps sample %d has negative fps: %v", i, f)
		}
	}
	logSamples(t, "fps", udid, samples)
}
