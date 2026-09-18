package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/ostrace"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWriteSSEFrame asserts the exact wire framing: `event: <name>\n`,
// `data: <compact-json>\n`, terminated by a blank line.
func TestWriteSSEFrame(t *testing.T) {
	var buf bytes.Buffer
	err := writeSSEFrame(&buf, "syslog", SyslogMessage{Message: "hello", Timestamp: 42})
	require.NoError(t, err)

	got := buf.String()
	assert.Equal(t, "event: syslog\ndata: {\"message\":\"hello\",\"timestamp\":42}\n\n", got)

	// data is compact single-line JSON (no embedded newlines before the terminator).
	dataLine := strings.SplitN(got, "\n", 3)[1]
	assert.True(t, strings.HasPrefix(dataLine, "data: "))
	assert.NotContains(t, strings.TrimPrefix(dataLine, "data: "), "\n")
}

// TestWriteSSEFrameHeartbeat asserts the heartbeat frame is an empty JSON object.
func TestWriteSSEFrameHeartbeat(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, writeSSEFrame(&buf, "heartbeat", struct{}{}))
	assert.Equal(t, "event: heartbeat\ndata: {}\n\n", buf.String())
}

// sseFrame is a parsed SSE frame.
type sseFrame struct {
	event string
	data  string
}

// parseSSE splits an SSE response body into its frames.
func parseSSE(t *testing.T, body string) []sseFrame {
	t.Helper()
	var frames []sseFrame
	for _, block := range strings.Split(strings.TrimRight(body, "\n"), "\n\n") {
		if strings.TrimSpace(block) == "" {
			continue
		}
		var f sseFrame
		for _, line := range strings.Split(block, "\n") {
			switch {
			case strings.HasPrefix(line, "event: "):
				f.event = strings.TrimPrefix(line, "event: ")
			case strings.HasPrefix(line, "data: "):
				f.data = strings.TrimPrefix(line, "data: ")
			}
		}
		frames = append(frames, f)
	}
	return frames
}

// runStreamSSE drives streamSSE with the given next until it ends, returning the
// parsed frames. The request context is cancelled after the deadline so a stuck
// stream does not hang the test.
func runStreamSSE(t *testing.T, next func() (sseEvent, bool)) []sseFrame {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/stream", nil)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	c.Request = req.WithContext(ctx)

	done := make(chan struct{})
	go func() {
		defer close(done)
		streamSSE(c, "udid-1", next)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("streamSSE did not terminate")
	}
	return parseSSE(t, w.Body.String())
}

// TestStreamSSEEmitsTypedFrames asserts streamSSE writes proper event/data frames
// with the event name the caller chose, then stops when next signals end.
func TestStreamSSEEmitsTypedFrames(t *testing.T) {
	msgs := []string{"a", "b", "c"}
	i := 0
	frames := runStreamSSE(t, func() (sseEvent, bool) {
		if i >= len(msgs) {
			return sseEvent{}, false
		}
		m := msgs[i]
		i++
		return sseEvent{event: "syslog", payload: SyslogMessage{Message: m}}, true
	})

	require.Len(t, frames, 3)
	for idx, f := range frames {
		assert.Equal(t, "syslog", f.event)
		var payload SyslogMessage
		require.NoError(t, json.Unmarshal([]byte(f.data), &payload))
		assert.Equal(t, msgs[idx], payload.Message)
	}
}

// TestStreamSSEHeartbeatOnIdle asserts a heartbeat frame is emitted when the
// stream is idle longer than the heartbeat interval, then real events resume.
func TestStreamSSEHeartbeatOnIdle(t *testing.T) {
	// Shrink the interval for the test via the package-level knob.
	orig := sseHeartbeatIntervalForTest
	sseHeartbeatIntervalForTest = 20 * time.Millisecond
	defer func() { sseHeartbeatIntervalForTest = orig }()

	step := 0
	frames := runStreamSSE(t, func() (sseEvent, bool) {
		step++
		switch step {
		case 1:
			// Idle long enough to force at least one heartbeat.
			time.Sleep(60 * time.Millisecond)
			return sseEvent{event: "syslog", payload: SyslogMessage{Message: "after-idle"}}, true
		default:
			return sseEvent{}, false
		}
	})

	var sawHeartbeat, sawSyslog bool
	for _, f := range frames {
		switch f.event {
		case "heartbeat":
			sawHeartbeat = true
			assert.Equal(t, "{}", f.data)
		case "syslog":
			sawSyslog = true
		}
	}
	assert.True(t, sawHeartbeat, "expected at least one heartbeat frame on idle; got %+v", frames)
	assert.True(t, sawSyslog, "expected the post-idle syslog frame")
}

// --- payload mapper tests ---

func TestToAppStateNotification(t *testing.T) {
	n := toAppStateNotification(map[string]interface{}{"bundleId": "com.apple.Preferences", "state": "foreground"})
	assert.Equal(t, "com.apple.Preferences", n.BundleID)
	assert.Equal(t, "foreground", n.State)
	assert.NotZero(t, n.Timestamp)
}

func TestToOsTraceEntry(t *testing.T) {
	ts := time.UnixMilli(1723200000000)
	e := ostrace.LogEntry{
		PID:       123,
		Timestamp: ts,
		LevelName: "info",
		ImageName: "SpringBoard",
		Message:   "hi",
		Label:     &ostrace.LogLabel{Subsystem: "com.apple.network", Category: "boringssl"},
	}
	out := toOsTraceEntry(e)
	assert.Equal(t, uint32(123), out.PID)
	assert.Equal(t, "SpringBoard", out.ProcessName)
	assert.Equal(t, "info", out.Level)
	assert.Equal(t, "com.apple.network", out.Subsystem)
	assert.Equal(t, "boringssl", out.Category)
	assert.Equal(t, "hi", out.Message)
	assert.Equal(t, int64(1723200000000), out.Timestamp)

	// camelCase JSON per the spec.
	b, err := json.Marshal(out)
	require.NoError(t, err)
	assert.Contains(t, string(b), "\"processName\"")
}

func TestToAttachDetachEvent(t *testing.T) {
	attached := toAttachDetachEvent(ios.AttachedMessage{
		MessageType: "Attached",
		DeviceID:    5,
		Properties:  ios.DeviceProperties{SerialNumber: "00008110-x", ConnectionType: "USB"},
	})
	assert.Equal(t, "attached", attached.Event)
	assert.Equal(t, 5, attached.DeviceID)
	assert.Equal(t, "00008110-x", attached.UDID)
	require.NotNil(t, attached.Properties)
	assert.Equal(t, "00008110-x", attached.Properties.SerialNumber)
	assert.Equal(t, "USB", attached.Properties.ConnectionType)

	// properties are camelCase per the spec.
	b, err := json.Marshal(attached)
	require.NoError(t, err)
	assert.Contains(t, string(b), "\"serialNumber\"")
	assert.Contains(t, string(b), "\"connectionType\"")

	detached := toAttachDetachEvent(ios.AttachedMessage{MessageType: "Detached", DeviceID: 5})
	assert.Equal(t, "detached", detached.Event)
	assert.Nil(t, detached.Properties)
}
