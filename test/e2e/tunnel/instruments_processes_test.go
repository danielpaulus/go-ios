//go:build e2e

package tunnel_test

import "testing"

const processesStreamSeconds = 5

// TestInstrumentsProcesses streams per-process sysmontap samples for
// SpringBoard (always running) over the iOS 17+ tunnel and asserts the command
// self-terminates and emits well-formed samples for that process with a
// numeric cpuUsage (the CLI skips the first snapshot, in which the device
// reports cpuUsage as null).
func TestInstrumentsProcesses(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		samples := streamNDJSON(t, udid, instrumentsSampleWindow,
			"instruments", "processes", "--process=SpringBoard", "--duration="+itoa(processesStreamSeconds))
		if len(samples) == 0 {
			t.Fatalf("instruments processes: no SpringBoard samples in %ds window", processesStreamSeconds)
		}
		for i, s := range samples {
			if s["name"] != "SpringBoard" {
				t.Fatalf("instruments processes sample %d is not SpringBoard: %v", i, s["name"])
			}
			if _, ok := s["pid"].(float64); !ok {
				t.Fatalf("instruments processes sample %d \"pid\" missing or not numeric: %v", i, s["pid"])
			}
			if cpu, ok := s["cpuUsage"].(float64); !ok || cpu < 0 {
				t.Fatalf("instruments processes sample %d \"cpuUsage\" missing, null or negative: %v", i, s["cpuUsage"])
			}
		}
		logSamples(t, "processes", udid, samples)
	})
}
