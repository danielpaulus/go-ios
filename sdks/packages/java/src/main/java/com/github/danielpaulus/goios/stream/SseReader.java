package com.github.danielpaulus.goios.stream;

import java.util.Iterator;
import java.util.NoSuchElementException;

/**
 * A pull-based Server-Sent Events reader over a line source.
 *
 * <p>Parses text/event-stream frames (blank-line delimited groups of
 * {@code event:} / {@code data:} lines, {@code :}-comment keep-alives ignored),
 * decodes each frame with the supplied {@link EventDecoder}, and yields typed
 * {@link SseEvent}s. Heartbeats are skipped unless {@code includeHeartbeats} is
 * set. Multi-line {@code data:} fields are joined with {@code \n}. A trailing
 * frame not terminated by a blank line is dispatched at end-of-stream.
 *
 * <p>Usable as an {@link Iterator} or in a {@code for-each} loop
 * ({@link Iterable}), and as an {@link AutoCloseable}: closing runs the
 * {@code onClose} hook exactly once (releasing the underlying HTTP response) and
 * stops further iteration, so a live stream can be cancelled mid-flight.
 */
public final class SseReader implements Iterator<SseEvent>, Iterable<SseEvent>, AutoCloseable {

    private final Iterator<String> lines;
    private final EventDecoder decoder;
    private final boolean includeHeartbeats;
    private final Runnable onClose;

    private SseEvent next;
    private boolean closed;
    private boolean onCloseRan;
    private boolean sawAnyLine; // whether we've consumed at least one line (for trailing-frame flush)

    public SseReader(Iterator<String> lines, EventDecoder decoder,
                     boolean includeHeartbeats, Runnable onClose) {
        this.lines = lines;
        this.decoder = decoder;
        this.includeHeartbeats = includeHeartbeats;
        this.onClose = onClose;
    }

    @Override
    public Iterator<SseEvent> iterator() {
        return this;
    }

    @Override
    public boolean hasNext() {
        if (next != null) {
            return true;
        }
        if (closed) {
            return false;
        }
        next = advance();
        return next != null;
    }

    @Override
    public SseEvent next() {
        if (!hasNext()) {
            throw new NoSuchElementException();
        }
        SseEvent ev = next;
        next = null;
        return ev;
    }

    /** Parse and decode the next dispatchable event, or {@code null} at end/close. */
    private SseEvent advance() {
        String eventName = null;
        StringBuilder data = null;
        boolean inFrame = false;

        while (!closed) {
            if (!lines.hasNext()) {
                // End of stream: flush a trailing frame with data but no blank terminator.
                if (inFrame && data != null) {
                    SseEvent ev = dispatch(eventName, data.toString());
                    return ev != null ? ev : null;
                }
                return null;
            }
            String line = lines.next();
            sawAnyLine = true;

            if (line.isEmpty()) {
                if (inFrame) {
                    SseEvent ev = dispatch(eventName, data == null ? "" : data.toString());
                    if (ev != null) {
                        return ev;
                    }
                    // Skipped (e.g. heartbeat): reset and keep scanning.
                    eventName = null;
                    data = null;
                    inFrame = false;
                }
                continue;
            }
            if (line.charAt(0) == ':') {
                // Comment / keep-alive line — ignore.
                continue;
            }
            inFrame = true;
            if (line.startsWith("event:")) {
                eventName = strip(line.substring("event:".length()));
            } else if (line.startsWith("data:")) {
                String chunk = strip(line.substring("data:".length()));
                if (data == null) {
                    data = new StringBuilder(chunk);
                } else {
                    data.append('\n').append(chunk);
                }
            }
            // Other field names (id:, retry:) are ignored.
        }
        return null;
    }

    /** Decode one frame, honoring heartbeat filtering; returns null if filtered out. */
    private SseEvent dispatch(String eventName, String data) {
        String name = eventName == null ? "message" : eventName;
        SseEvent ev = decoder.apply(name, data);
        if (ev instanceof HeartbeatEvent && !includeHeartbeats) {
            return null;
        }
        return ev;
    }

    private static String strip(String s) {
        // A single optional leading space after the colon is part of the SSE format.
        return s.startsWith(" ") ? s.substring(1) : s;
    }

    @Override
    public void close() {
        closed = true;
        next = null;
        if (!onCloseRan) {
            onCloseRan = true;
            if (onClose != null) {
                onClose.run();
            }
        }
    }
}
