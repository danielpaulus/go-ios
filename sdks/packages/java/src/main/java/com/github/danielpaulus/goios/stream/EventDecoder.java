package com.github.danielpaulus.goios.stream;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.github.danielpaulus.goios.generated.invoker.JSON;
import com.github.danielpaulus.goios.generated.model.AppStateNotification;
import com.github.danielpaulus.goios.generated.model.CpuUsageSample;
import com.github.danielpaulus.goios.generated.model.JobLogLine;
import com.github.danielpaulus.goios.generated.model.OsTraceEntry;
import com.github.danielpaulus.goios.generated.model.SyslogMessage;

/**
 * Decodes one SSE frame (its {@code event:} name and JSON {@code data:} payload)
 * into a typed {@link SseEvent}. One instance exists per endpoint payload shape;
 * a {@code heartbeat} frame is always decoded to {@link HeartbeatEvent} and an
 * unrecognized name to {@link UnknownEvent} regardless of the endpoint.
 */
@FunctionalInterface
public interface EventDecoder {

    /** Decode {@code data} for the given {@code eventName} into a typed event. */
    SseEvent apply(String eventName, String data);

    /** Shared Jackson mapper matching the generated models' (de)serialization. */
    ObjectMapper MAPPER = new JSON().getMapper();

    static <T> T read(String data, Class<T> type) {
        try {
            return MAPPER.readValue(data == null ? "{}" : data, type);
        } catch (Exception e) {
            throw new IllegalStateException("failed to decode SSE payload: " + e.getMessage(), e);
        }
    }

    /** Decoder for the {@code /syslog} stream. */
    EventDecoder SYSLOG = (name, data) -> switch (name) {
        case "syslog" -> new SyslogEvent(read(data, SyslogMessage.class));
        case "heartbeat" -> new HeartbeatEvent();
        default -> new UnknownEvent(name, data);
    };

    /** Decoder for the {@code /notifications} stream. */
    EventDecoder NOTIFICATIONS = (name, data) -> switch (name) {
        case "appstate" -> new AppStateEvent(read(data, AppStateNotification.class));
        case "heartbeat" -> new HeartbeatEvent();
        default -> new UnknownEvent(name, data);
    };

    /** Decoder for the {@code /ostrace} stream. */
    EventDecoder OSTRACE = (name, data) -> switch (name) {
        case "ostrace" -> new OsTraceEvent(read(data, OsTraceEntry.class));
        case "heartbeat" -> new HeartbeatEvent();
        default -> new UnknownEvent(name, data);
    };

    /** Decoder for the {@code /listen} device attach/detach stream. */
    EventDecoder LISTEN = (name, data) -> switch (name) {
        case "attachdetach" -> new AttachDetachEvent(
                read(data, com.github.danielpaulus.goios.generated.model.AttachDetachEvent.class));
        case "heartbeat" -> new HeartbeatEvent();
        default -> new UnknownEvent(name, data);
    };

    /** Decoder for the {@code /sysmontap} CPU-sample stream. */
    EventDecoder SYSMONTAP = (name, data) -> switch (name) {
        case "sample" -> new SysmontapEvent(read(data, CpuUsageSample.class));
        case "heartbeat" -> new HeartbeatEvent();
        default -> new UnknownEvent(name, data);
    };

    /** Decoder for a {@code /jobs/{id}/logs} stream. */
    EventDecoder JOB_LOGS = (name, data) -> switch (name) {
        case "log" -> new JobLogEvent(read(data, JobLogLine.class));
        case "heartbeat" -> new HeartbeatEvent();
        default -> new UnknownEvent(name, data);
    };
}
