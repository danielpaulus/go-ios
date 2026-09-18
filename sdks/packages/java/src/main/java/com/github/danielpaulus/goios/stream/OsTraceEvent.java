package com.github.danielpaulus.goios.stream;

import com.github.danielpaulus.goios.generated.model.OsTraceEntry;

/** An {@code ostrace} event carrying a decoded {@link OsTraceEntry}. */
public record OsTraceEvent(OsTraceEntry payload) implements SseEvent {
    @Override
    public String eventName() {
        return "ostrace";
    }
}
