package com.github.danielpaulus.goios.stream;

/** An {@code attachdetach} device-listen event (attach/detach/pair). */
public record AttachDetachEvent(com.github.danielpaulus.goios.generated.model.AttachDetachEvent payload)
        implements SseEvent {
    @Override
    public String eventName() {
        return "attachdetach";
    }
}
