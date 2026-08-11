package com.github.danielpaulus.goios.stream;

/**
 * Base type for a decoded Server-Sent Event emitted by a go-ios SSE endpoint
 * (syslog, notifications, ostrace, listen, sysmontap, job logs).
 *
 * <p>Every concrete event carries the wire {@code event:} name via
 * {@link #eventName()}. Typed payload events additionally expose a strongly
 * typed {@code payload()} accessor; unrecognized events are surfaced as
 * {@link UnknownEvent} rather than being dropped.
 */
public sealed interface SseEvent
        permits SyslogEvent, AppStateEvent, OsTraceEvent, AttachDetachEvent,
                SysmontapEvent, JobLogEvent, HeartbeatEvent, UnknownEvent {

    /** The wire {@code event:} name (e.g. {@code "syslog"}, {@code "heartbeat"}). */
    String eventName();
}
