package com.github.danielpaulus.goios.stream;

/**
 * A keep-alive {@code heartbeat} event. Skipped by default; surfaced only when a
 * stream is opened with {@code includeHeartbeats == true}.
 */
public record HeartbeatEvent() implements SseEvent {
    @Override
    public String eventName() {
        return "heartbeat";
    }
}
