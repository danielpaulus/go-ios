package com.github.danielpaulus.goios.stream;

import com.github.danielpaulus.goios.generated.model.AppStateNotification;

/** An {@code appstate} notification event. */
public record AppStateEvent(AppStateNotification payload) implements SseEvent {
    @Override
    public String eventName() {
        return "appstate";
    }
}
