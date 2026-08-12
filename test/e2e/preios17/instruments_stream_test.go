//go:build e2e

// Streaming instruments/DTX commands on pre-iOS17 devices: these reach the
// graphics (FPS) and network-monitoring DTX services over usbmuxd with only a
// mounted Developer Disk Image (done in TestMain) and NO tunnel. The tunnel
// suite proves the same commands over the iOS 17+ tunnel; this proves them on
// the classic transport.
package preios17_test

import (
	"bufio"
	"encoding/json"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// instrumentsSampleWindow bounds a --duration=<n> streaming run: the command
// must self-terminate within it (slack over the sampling duration for DTX
// connection setup and teardown).
const (
	instrSampleDurationSeconds = 5
	instrumentsSampleWindow    = 30 * time.Second
)

// TestInstrumentsFPS streams frames-per-second samples from the graphics OpenGL
// service and asserts the command self-terminates and emits well-formed
// {"fps": <number>} samples (an idle screen still emits CoreAnimation frames).
func TestInstrumentsFPS(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		samples := streamNDJSON(t, udid, instrumentsSampleWindow,
			"instruments", "fps", "--duration="+strconv.Itoa(instrSampleDurationSeconds))
		if len(samples) == 0 {
			t.Fatalf("instruments fps: no samples in %ds window", instrSampleDurationSeconds)
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

// networkStreamSeconds bounds the network stream. The instruments
// network-monitoring service only emits samples when there is actual network
// activity (an idle device produces none), so we stream for a longer window
// while generating traffic (launch/kill churn) and assert well-formed
// {"type": <number>, "data": ...} envelopes arrived before the command
// self-terminated via --duration.
const networkStreamSeconds = 20

// TestInstrumentsNetwork streams network-monitoring samples while generating
// traffic on the device and asserts the command self-terminates and emits
// well-formed {"type": <number>, "data": ...} envelopes.
func TestInstrumentsNetwork(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		output, stop := startBackground(t, udid, syscall.SIGTERM,
			"instruments", "network", "--duration="+strconv.Itoa(networkStreamSeconds))
		defer stop()

		// Generate network traffic for most of the streaming window by churning
		// app launches (Settings phones home on launch on these OS versions).
		deadline := time.Now().Add(time.Duration(networkStreamSeconds-3) * time.Second)
		for time.Now().Before(deadline) {
			runIOSForDevice(t, udid, "launch", "com.apple.Preferences")
			runIOSForDevice(t, udid, "launch", "com.apple.mobilesafari")
			time.Sleep(2 * time.Second)
		}

		// Let the --duration-bounded command reach its deadline and flush.
		time.Sleep(6 * time.Second)

		samples := parseNetworkSamples(t, output())
		if len(samples) == 0 {
			t.Fatalf("instruments network: no samples in %ds window even with generated traffic", networkStreamSeconds)
		}
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

// parseNetworkSamples decodes the JSON-object lines the streaming command wrote,
// ignoring any interleaved non-JSON log noise from the background runner.
func parseNetworkSamples(t *testing.T, out string) []map[string]any {
	t.Helper()
	var samples []map[string]any
	sc := bufio.NewScanner(strings.NewReader(out))
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if !strings.HasPrefix(line, "{") {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			continue
		}
		if _, ok := m["type"]; ok {
			samples = append(samples, m)
		}
	}
	return samples
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
