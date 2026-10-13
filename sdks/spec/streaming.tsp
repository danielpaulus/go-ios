import "@typespec/http";
import "@typespec/sse";
import "@typespec/events";

using Http;
using Events;
using SSE;

namespace GoIos;

// ===========================================================================
// Server-Sent Events (SSE) event payload models + event unions.
//
// Each streaming endpoint returns `SSEStream<...>` which sets the response
// content-type to `text/event-stream`. Every union is marked `@events`; each
// *named* union variant becomes the SSE `event:` name and its model is the JSON
// serialized into the `data:` frame. A `heartbeat` event (empty object) is
// included on every stream so clients can detect a live-but-idle connection and
// keep-alives are self-describing.
//
// go-ios server contract (companion PR): frames MUST be emitted as
//   event: <variant-name>\n
//   data: <compact-json>\n\n
// with `heartbeat` sent on an idle interval.
// ===========================================================================

// ---------------------------------------------------------------------------
// Shared: heartbeat keep-alive
// ---------------------------------------------------------------------------

/** Periodic keep-alive frame emitted on every stream. Payload is empty. */
model Heartbeat {}

// ---------------------------------------------------------------------------
// /device/{udid}/notifications  — application state changes
// ---------------------------------------------------------------------------

/** An app foreground/background/lifecycle state change. */
model AppStateNotification {
  /** Bundle id of the app whose state changed. */
  bundleId: string;

  /**
   * New application state.
   * Typical values: `foreground`, `background`, `suspended`, `terminated`,
   * `unknown`.
   */
  state: string;

  /** Unix epoch milliseconds when the change was observed. */
  timestamp?: int64;
}

@events
union NotificationEvents {
  /** An application changed lifecycle state. */
  appstate: AppStateNotification,

  /** Keep-alive. */
  heartbeat: Heartbeat,
}

// ---------------------------------------------------------------------------
// /device/{udid}/syslog  — device syslog lines
// ---------------------------------------------------------------------------

/** A single syslog line from the device. */
model SyslogMessage {
  /** The raw log message text. */
  message: string;

  /** Unix epoch milliseconds when the line was emitted, if known. */
  timestamp?: int64;
}

@events
union SyslogEvents {
  /** A syslog line. */
  syslog: SyslogMessage,

  /** Keep-alive. */
  heartbeat: Heartbeat,
}

// ---------------------------------------------------------------------------
// /device/{udid}/ostrace  — structured os_log / oslog trace entries
// ---------------------------------------------------------------------------

/** A structured os_log trace entry. */
model OsTraceEntry {
  /** Process id that emitted the entry. */
  pid?: int32;

  /** Emitting process/executable name. */
  processName?: string;

  /** Log level, e.g. `default`, `info`, `debug`, `error`, `fault`. */
  level?: string;

  /** Subsystem string (e.g. `com.apple.network`). */
  subsystem?: string;

  /** Category within the subsystem. */
  category?: string;

  /** The formatted log message. */
  message: string;

  /** Unix epoch milliseconds when the entry was emitted, if known. */
  timestamp?: int64;
}

@events
union OsTraceEvents {
  /** A structured trace entry. */
  ostrace: OsTraceEntry,

  /** Keep-alive. */
  heartbeat: Heartbeat,
}

// ---------------------------------------------------------------------------
// /device/{udid}/listen  — device attach / detach events
// ---------------------------------------------------------------------------

/** A device was attached to or detached from the host. */
model AttachDetachEvent {
  /**
   * Event kind.
   * `attached` when a device connects, `detached` when it disconnects,
   * `paired` when a pairing record appears.
   */
  event: string;

  /** usbmuxd device id. */
  deviceID?: int32;

  /** The device udid (serial number), when known. */
  udid?: string;

  /** Full device properties, present on `attached`. */
  properties?: DeviceProperties;
}

@events
union ListenEvents {
  /** A device attach/detach/pair event. */
  attachdetach: AttachDetachEvent,

  /** Keep-alive. */
  heartbeat: Heartbeat,
}

// ---------------------------------------------------------------------------
// /device/{udid}/sysmontap  — CPU usage samples
// ---------------------------------------------------------------------------

/** A single sysmontap CPU-usage sample. Open map; sampler keys vary by OS. */
model CpuUsageSample is Record<unknown> {
  /** Total CPU load across all cores (0–100). */
  CPU_TotalLoad?: float64;

  /** System (kernel) CPU load. */
  SystemLoad?: float64;

  /** User CPU load. */
  UserLoad?: float64;
}

@events
union SysmontapEvents {
  /** A CPU-usage sample. */
  sample: CpuUsageSample,

  /** Keep-alive. */
  heartbeat: Heartbeat,
}

// ---------------------------------------------------------------------------
// /device/{udid}/jobs/{id}/logs  — streamed job log output
// ---------------------------------------------------------------------------

/** A single line of a job's log output. */
model JobLogLine {
  /** The raw log line (already newline-terminated in the buffer). */
  line: string;
}

@events
union JobLogEvents {
  /** A job log line. */
  log: JobLogLine,

  /** Keep-alive. */
  heartbeat: Heartbeat,
}
