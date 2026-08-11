package com.github.danielpaulus.goios.stream;

/**
 * An event whose {@code event:} name is not recognized by the SDK. The raw
 * name and JSON data are preserved so callers can handle forward-compatible
 * event types the SDK does not yet model.
 */
public record UnknownEvent(String eventName, String rawData) implements SseEvent {
}
