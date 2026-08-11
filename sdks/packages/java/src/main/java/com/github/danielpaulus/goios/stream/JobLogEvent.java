package com.github.danielpaulus.goios.stream;

import com.github.danielpaulus.goios.generated.model.JobLogLine;

/** A {@code log} event from a job-logs stream carrying one output line. */
public record JobLogEvent(JobLogLine payload) implements SseEvent {
    @Override
    public String eventName() {
        return "log";
    }
}
