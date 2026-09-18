package api

import (
	"encoding/json"
	"io"
	"time"

	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/gin-gonic/gin"
)

// sseHeartbeatInterval is how long a stream may sit idle before a heartbeat
// frame is emitted, so clients can tell a live-but-idle connection from a dead
// one and intermediaries keep the connection open.
const sseHeartbeatInterval = 15 * time.Second

// sseHeartbeatIntervalForTest overrides the heartbeat interval when > 0. It
// exists only so unit tests can force an idle heartbeat without waiting the full
// production interval; production code never sets it.
var sseHeartbeatIntervalForTest time.Duration

// sseEvent is a single Server-Sent Event: the event name and the payload that
// will be serialized as compact JSON into the data field.
type sseEvent struct {
	event   string
	payload any
}

// writeSSEFrame writes one SSE frame to w in the exact wire framing the SDK
// contract requires:
//
//	event: <event-name>\n
//	data: <compact-json-of-payload>\n
//	\n
//
// The data payload is always compact (single-line) JSON.
func writeSSEFrame(w io.Writer, event string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		// Fall back to an error envelope so a marshal failure of one frame
		// never tears down the whole stream.
		data, _ = json.Marshal(GenericResponse{Error: "failed to marshal event payload: " + err.Error()})
	}
	if _, err := io.WriteString(w, "event: "+event+"\n"); err != nil {
		return err
	}
	if _, err := w.Write(append([]byte("data: "), data...)); err != nil {
		return err
	}
	_, err = io.WriteString(w, "\n\n")
	return err
}

// streamSSE drives a Server-Sent Events response. It pulls events from next in
// a background goroutine and writes them to the client as SSE frames, emitting a
// `heartbeat` frame whenever the stream is idle for sseHeartbeatInterval. It
// returns when next signals end-of-stream (ok=false) or the client disconnects.
//
// next is called repeatedly; each call should block until the next event is
// available and return (event, true), or (_, false) once the source is
// exhausted or errors out. next runs on its own goroutine so a blocking read
// never starves the heartbeat timer.
//
// udid is attached to log lines for filterability; it may be empty for
// host-scoped streams (e.g. /listen).
func streamSSE(c *gin.Context, udid string, next func() (sseEvent, bool)) {
	events := make(chan sseEvent)
	done := make(chan struct{})
	go func() {
		defer close(events)
		for {
			ev, ok := next()
			if !ok {
				return
			}
			select {
			case events <- ev:
			case <-done:
				return
			}
		}
	}()
	defer close(done)

	interval := sseHeartbeatInterval
	if sseHeartbeatIntervalForTest > 0 {
		interval = sseHeartbeatIntervalForTest
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	w := c.Writer
	for {
		select {
		case ev, ok := <-events:
			if !ok {
				return
			}
			if err := writeSSEFrame(w, ev.event, ev.payload); err != nil {
				golog.Debug("sse client write failed", "module", logModule, "udid", udid, "error", err.Error())
				return
			}
			w.Flush()
			ticker.Reset(interval)
		case <-ticker.C:
			if err := writeSSEFrame(w, "heartbeat", struct{}{}); err != nil {
				golog.Debug("sse heartbeat write failed", "module", logModule, "udid", udid, "error", err.Error())
				return
			}
			w.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}
