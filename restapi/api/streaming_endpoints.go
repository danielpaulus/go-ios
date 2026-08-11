package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/danielpaulus/go-ios/ios/instruments"
	"github.com/danielpaulus/go-ios/ios/ostrace"
	"github.com/danielpaulus/go-ios/ios/syslog"
	"github.com/gin-gonic/gin"
)

// SSE payload models. These mirror the field names in the SDK contract
// (spec/openapi/openapi.yaml component schemas) so generated clients decode the
// data frames directly.

// AppStateNotification is the payload of an `appstate` event (SSE /notifications).
type AppStateNotification struct {
	BundleID  string `json:"bundleId"`
	State     string `json:"state"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

// SyslogMessage is the payload of a `syslog` event (SSE /syslog).
type SyslogMessage struct {
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

// OsTraceEntry is the payload of an `ostrace` event (SSE /ostrace).
type OsTraceEntry struct {
	PID         uint32 `json:"pid,omitempty"`
	ProcessName string `json:"processName,omitempty"`
	Level       string `json:"level,omitempty"`
	Subsystem   string `json:"subsystem,omitempty"`
	Category    string `json:"category,omitempty"`
	Message     string `json:"message"`
	Timestamp   int64  `json:"timestamp,omitempty"`
}

// DeviceProperties mirrors the spec's DeviceProperties schema (camelCase JSON),
// since ios.DeviceProperties has no JSON tags and would marshal PascalCase.
type DeviceProperties struct {
	ConnectionSpeed int    `json:"connectionSpeed,omitempty"`
	ConnectionType  string `json:"connectionType,omitempty"`
	DeviceID        int    `json:"deviceID,omitempty"`
	LocationID      int    `json:"locationID,omitempty"`
	ProductID       int    `json:"productID,omitempty"`
	SerialNumber    string `json:"serialNumber"`
}

// AttachDetachEvent is the payload of an `attachdetach` event (SSE /listen).
type AttachDetachEvent struct {
	Event      string            `json:"event"`
	DeviceID   int               `json:"deviceID,omitempty"`
	UDID       string            `json:"udid,omitempty"`
	Properties *DeviceProperties `json:"properties,omitempty"`
}

// Notifications streams application-state change events as Server-Sent Events.
// Each `appstate` event carries an AppStateNotification; a `heartbeat` event is
// emitted on idle.
// @Summary      Stream app state notifications (SSE)
// @Description  Streams application foreground/background/lifecycle state changes as text/event-stream. Events: `appstate` (AppStateNotification), `heartbeat`.
// @Tags         general
// @Produce      text/event-stream
// @Success      200  {object}  AppStateNotification
// @Router       /notifications [get]
func Notifications(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	udid := device.Properties.SerialNumber
	listenerFunc, closeFunc, err := instruments.ListenAppStateNotifications(device)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer closeFunc()
	golog.Info("notifications stream started", "module", logModule, "udid", udid)
	streamSSE(c, udid, func() (sseEvent, bool) {
		notification, err := listenerFunc()
		if err != nil {
			return sseEvent{}, false
		}
		return sseEvent{event: "appstate", payload: toAppStateNotification(notification)}, true
	})
}

// toAppStateNotification maps the untyped instruments notification map to the
// spec's AppStateNotification shape. The device reports bundle id and state
// under a few known keys; unknown-shaped maps fall through with best effort.
func toAppStateNotification(m map[string]interface{}) AppStateNotification {
	n := AppStateNotification{Timestamp: time.Now().UnixMilli()}
	for _, k := range []string{"bundleId", "bundleID", "appBundleId"} {
		if v, ok := m[k].(string); ok && v != "" {
			n.BundleID = v
			break
		}
	}
	for _, k := range []string{"state", "appState", "runningState"} {
		if v, ok := m[k].(string); ok && v != "" {
			n.State = v
			break
		}
	}
	return n
}

// Syslog streams device syslog lines as Server-Sent Events. Each `syslog` event
// carries a SyslogMessage; a `heartbeat` event is emitted on idle.
// @Summary      Stream device syslog (SSE)
// @Description  Streams raw syslog lines as text/event-stream. Events: `syslog` (SyslogMessage), `heartbeat`.
// @Tags         general
// @Produce      text/event-stream
// @Success      200  {object}  SyslogMessage
// @Router       /syslog [get]
func Syslog(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	udid := device.Properties.SerialNumber
	syslogConnection, err := syslog.New(device)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer syslogConnection.Close()
	golog.Info("syslog stream started", "module", logModule, "udid", udid)
	streamSSE(c, udid, func() (sseEvent, bool) {
		m, err := syslogConnection.ReadLogMessage()
		if err != nil {
			return sseEvent{}, false
		}
		return sseEvent{event: "syslog", payload: SyslogMessage{Message: m, Timestamp: time.Now().UnixMilli()}}, true
	})
}

// OsTrace streams structured os_log trace entries via os_trace_relay as
// Server-Sent Events. Each `ostrace` event carries an OsTraceEntry; a
// `heartbeat` event is emitted on idle. Optional filters (pid, level, subsystem,
// match, exclude) combine with AND semantics.
// @Summary      Stream structured os_log trace (SSE)
// @Description  Streams structured os_log entries as text/event-stream. Events: `ostrace` (OsTraceEntry), `heartbeat`. Filters combine with AND.
// @Tags         general
// @Param        pid        query  int     false  "Filter by process ID"
// @Param        level      query  string  false  "Minimum log level (info, debug, error, ...)"
// @Param        subsystem  query  string  false  "Filter by subsystem"
// @Param        match      query  string  false  "Include only messages matching this substring"
// @Param        exclude    query  string  false  "Exclude messages matching this substring"
// @Produce      text/event-stream
// @Success      200  {object}  OsTraceEntry
// @Router       /ostrace [get]
func OsTrace(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	udid := device.Properties.SerialNumber
	pid := -1
	if pidStr := c.Query("pid"); pidStr != "" {
		var err error
		pid, err = strconv.Atoi(pidStr)
		if err != nil {
			RespondError(c, http.StatusBadRequest, errInvalidPID)
			return
		}
	}
	levelFilter, err := ostrace.ParseLevelFilter(c.Query("level"))
	if err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	clientFilter := ostrace.ClientFilter{
		Levels:    levelFilter.ClientLevels,
		Subsystem: c.Query("subsystem"),
		Match:     c.Query("match"),
		Exclude:   c.Query("exclude"),
	}
	conn, err := ostrace.New(device, pid, levelFilter.MessageFilter, levelFilter.StreamFlags)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer conn.Close()
	golog.Info("ostrace stream started", "module", logModule, "udid", udid, "pid", pid)
	streamSSE(c, udid, func() (sseEvent, bool) {
		entry, err := conn.ReadFilteredEntry(clientFilter)
		if err != nil {
			return sseEvent{}, false
		}
		return sseEvent{event: "ostrace", payload: toOsTraceEntry(entry)}, true
	})
}

// toOsTraceEntry maps ios/ostrace.LogEntry to the spec's OsTraceEntry shape.
func toOsTraceEntry(e ostrace.LogEntry) OsTraceEntry {
	out := OsTraceEntry{
		PID:         e.PID,
		ProcessName: e.ImageName,
		Level:       e.LevelName,
		Message:     e.Message,
	}
	if !e.Timestamp.IsZero() {
		out.Timestamp = e.Timestamp.UnixMilli()
	}
	if e.Label != nil {
		out.Subsystem = e.Label.Subsystem
		out.Category = e.Label.Category
	}
	return out
}

// Listen streams device attach/detach events as Server-Sent Events. Each
// `attachdetach` event carries an AttachDetachEvent; a `heartbeat` event is
// emitted on idle. This stream is host-scoped (not device-scoped).
// @Summary      Stream device attach/detach events (SSE)
// @Description  Streams usbmuxd attach/detach events as text/event-stream. Events: `attachdetach` (AttachDetachEvent), `heartbeat`.
// @Tags         general
// @Produce      text/event-stream
// @Success      200  {object}  AttachDetachEvent
// @Router       /listen [get]
func Listen(c *gin.Context) {
	a, closeFunc, err := ios.Listen()
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer closeFunc()
	golog.Info("listen stream started", "module", logModule)
	streamSSE(c, "", func() (sseEvent, bool) {
		msg, err := a()
		if err != nil {
			return sseEvent{}, false
		}
		return sseEvent{event: "attachdetach", payload: toAttachDetachEvent(msg)}, true
	})
}

// toAttachDetachEvent maps ios.AttachedMessage to the spec's AttachDetachEvent
// shape. `properties` is present on attach events.
func toAttachDetachEvent(m ios.AttachedMessage) AttachDetachEvent {
	ev := AttachDetachEvent{
		DeviceID: m.DeviceID,
		UDID:     m.Properties.SerialNumber,
	}
	switch m.MessageType {
	case "Attached":
		ev.Event = "attached"
		ev.Properties = &DeviceProperties{
			ConnectionSpeed: m.Properties.ConnectionSpeed,
			ConnectionType:  m.Properties.ConnectionType,
			DeviceID:        m.Properties.DeviceID,
			LocationID:      m.Properties.LocationID,
			ProductID:       m.Properties.ProductID,
			SerialNumber:    m.Properties.SerialNumber,
		}
	case "Detached":
		ev.Event = "detached"
	case "Paired":
		ev.Event = "paired"
	default:
		ev.Event = m.MessageType
	}
	return ev
}
