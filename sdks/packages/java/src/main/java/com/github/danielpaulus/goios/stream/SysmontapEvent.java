package com.github.danielpaulus.goios.stream;

import com.github.danielpaulus.goios.generated.model.CpuUsageSample;

/** A {@code sample} event from the sysmontap stream carrying a CPU-usage sample. */
public record SysmontapEvent(CpuUsageSample payload) implements SseEvent {
    @Override
    public String eventName() {
        return "sample";
    }
}
