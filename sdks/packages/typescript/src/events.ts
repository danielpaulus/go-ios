import type {
  AppStateNotification,
  AttachDetachEvent,
  CpuUsageSample,
  JobLogLine,
  OsTraceEntry,
  SyslogMessage,
} from "./generated/types.gen";

/**
 * Typed SSE event maps, one per streaming endpoint. Each maps the SSE `event:`
 * name to its `data:` payload model, mirroring the `x-sse-events` vendor
 * extension in the OpenAPI 3.1 spec (and the inline `itemSchema` of the 3.2
 * variant). `heartbeat` is intentionally omitted here — the SSE parser consumes
 * heartbeats and never surfaces them.
 *
 * These maps drive the typed union returned by each streaming facade method
 * (`SseEvent<Map>`), so consumers can narrow on `event`.
 */

/** `GET /device/{udid}/syslog` — `x-sse-events.events`: `{ syslog: SyslogMessage }`. */
export interface SyslogEventMap {
  syslog: SyslogMessage;
}

/** `GET /device/{udid}/notifications` — `{ appstate: AppStateNotification }`. */
export interface NotificationEventMap {
  appstate: AppStateNotification;
}

/** `GET /device/{udid}/ostrace` — `{ ostrace: OsTraceEntry }`. */
export interface OsTraceEventMap {
  ostrace: OsTraceEntry;
}

/** `GET /device/{udid}/listen` — `{ attachdetach: AttachDetachEvent }`. */
export interface ListenEventMap {
  attachdetach: AttachDetachEvent;
}

/** `GET /device/{udid}/sysmontap` — `x-sse-events`: `{ sample: CpuUsageSample }`. */
export interface SysmontapEventMap {
  sample: CpuUsageSample;
}

/** `GET /device/{udid}/jobs/{id}/logs` — `x-sse-events`: `{ log: JobLogLine }`. */
export interface JobLogEventMap {
  log: JobLogLine;
}
