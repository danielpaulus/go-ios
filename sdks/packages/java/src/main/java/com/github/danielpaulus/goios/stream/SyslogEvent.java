package com.github.danielpaulus.goios.stream;

import com.github.danielpaulus.goios.generated.model.SyslogMessage;

/** A {@code syslog} event carrying a decoded {@link SyslogMessage}. */
public record SyslogEvent(SyslogMessage payload) implements SseEvent {
    @Override
    public String eventName() {
        return "syslog";
    }
}
