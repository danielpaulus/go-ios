//go:build e2e

package tunnel_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/danielpaulus/go-ios/test/e2e/harness"
)

// Streaming commands run until killed, so they are tested by letting them stream
// for a short window and asserting the output is well-formed.
const streamWindow = 6 * time.Second

// assertJSONLogLine scans streamed output for at least one line that parses as a
// JSON object containing every key in requiredKeys.
func assertJSONLogLine(t *testing.T, out []byte, requiredKeys ...string) {
	t.Helper()
	sc := bufio.NewScanner(bytes.NewReader(out))
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		var m map[string]any
		if json.Unmarshal(sc.Bytes(), &m) != nil {
			continue
		}
		ok := true
		for _, k := range requiredKeys {
			if _, present := m[k]; !present {
				ok = false
				break
			}
		}
		if ok {
			return
		}
	}
	t.Fatalf("no JSON log line with keys %v in streamed output", requiredKeys)
}

// TestSyslog asserts syslog streams JSON lines of the form {"msg": "..."}.
func TestSyslog(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		out := harness.StreamSmoke(t, udid, streamWindow, "syslog")
		assertJSONLogLine(t, out, "msg")
	})
}

// TestOstrace asserts ostrace streams structured log entries with a timestamp
// and message.
func TestOstrace(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		out := harness.StreamSmoke(t, udid, streamWindow, "ostrace")
		assertJSONLogLine(t, out, "timestamp", "message")
	})
}
