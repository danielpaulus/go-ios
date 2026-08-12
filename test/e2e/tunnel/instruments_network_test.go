//go:build e2e

package tunnel_test

import (
	"bufio"
	"encoding/json"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// The instruments network-monitoring service only emits samples when there is
// actual network activity — an idle device produces none (go-ios's own README
// notes "open apple maps or similar to force network traffic"). So we stream for
// a longer, --duration-bounded window while generating traffic on the device (a
// networked app plus repeated launch/kill churn), then assert well-formed
// {"type": <number>, "data": ...} envelopes arrived and the command self-stopped.
const (
	networkStreamSeconds = 20
	// networkApp is a system app that phones home on launch, generating traffic.
	networkApp = "com.apple.AppStore"
)

func TestInstrumentsNetwork(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		assertNetworkSamples(t, udid)
	})
}

func assertNetworkSamples(t *testing.T, udid string) {
	t.Helper()

	// Start the bounded network stream in the background so we can drive traffic
	// on the device while it collects samples. --duration makes it self-terminate.
	output, stop := startBackground(t, udid, syscall.SIGTERM,
		"instruments", "network", "--duration="+strconv.Itoa(networkStreamSeconds))
	defer stop()

	// Generate network traffic for most of the streaming window: launch a
	// networked app (phones home) and churn launch/kill so the network service
	// has activity to report.
	deadline := time.Now().Add(time.Duration(networkStreamSeconds-3) * time.Second)
	runIOSForDevice(t, udid, "launch", networkApp)
	for time.Now().Before(deadline) {
		runIOSForDevice(t, udid, "launch", "com.apple.Preferences")
		runIOSForDevice(t, udid, "launch", networkApp)
		time.Sleep(2 * time.Second)
	}
	runIOSForDevice(t, udid, "kill", networkApp)

	// Give the --duration-bounded command time to reach its deadline and flush
	// its final samples, then read what it collected. The deferred stop() cleans
	// up the (by now self-terminated) process group.
	time.Sleep(6 * time.Second)

	samples := parseNetworkSamples(t, output())
	if len(samples) == 0 {
		t.Fatalf("instruments network: no samples in %ds window even with generated traffic (real service/CLI issue?)", networkStreamSeconds)
	}
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
}

// parseNetworkSamples decodes the JSON-object lines the streaming command wrote
// to stdout, ignoring any non-JSON log noise the background runner may have
// interleaved on the same stream.
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
