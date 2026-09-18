package com.github.danielpaulus.goios.examples;

import com.github.danielpaulus.goios.Device;
import com.github.danielpaulus.goios.IosClient;
import com.github.danielpaulus.goios.stream.SseEvent;
import com.github.danielpaulus.goios.stream.SseReader;
import com.github.danielpaulus.goios.stream.SyslogEvent;

/**
 * Example 5 — stream the device syslog over Server-Sent Events.
 *
 * <p>{@code device.syslog()} opens {@code GET /device/{udid}/syslog} as an
 * {@link SseReader}, which is both an {@code Iterable<SseEvent>} and an
 * {@code AutoCloseable}. We iterate the typed events, printing each decoded
 * {@link SyslogEvent}, and stop after roughly {@value #MAX_EVENTS} events or
 * {@value #MAX_MILLIS} ms — whichever comes first — so the example terminates on
 * its own. Closing the reader (via try-with-resources) cancels the underlying
 * HTTP stream immediately.
 *
 * <p>Heartbeats are parsed and skipped by default, so the loop only sees real
 * payload events. Any unrecognized event name would arrive as an
 * {@code UnknownEvent} rather than being dropped.
 *
 * <p>Skips gracefully when no device is attached.
 */
public final class StreamSyslogExample {

    private static final int MAX_EVENTS = 20;
    private static final long MAX_MILLIS = 5_000;

    private StreamSyslogExample() {
    }

    public static void main(String[] args) {
        Env.requireApiKey();

        try (IosClient client = Env.client()) {
            String udid = Env.resolveUdid(client);
            if (udid == null) {
                System.out.println("SKIP StreamSyslogExample: no device attached.");
                return;
            }

            Device device = client.device(udid);
            System.out.printf("Streaming syslog from %s (up to %d events or %d ms) ...%n",
                    udid, MAX_EVENTS, MAX_MILLIS);

            long deadline = System.currentTimeMillis() + MAX_MILLIS;
            int count = 0;

            // try-with-resources guarantees the stream is cancelled when we break out.
            try (SseReader syslog = device.syslog()) {
                for (SseEvent ev : syslog) {
                    // Pattern-match to the typed event to reach the decoded payload.
                    if (ev instanceof SyslogEvent s) {
                        System.out.println("  " + s.payload().getMessage());
                    } else {
                        System.out.println("  [" + ev.eventName() + "]");
                    }
                    count++;
                    if (count >= MAX_EVENTS || System.currentTimeMillis() >= deadline) {
                        break; // leaving the loop closes the reader and aborts the stream
                    }
                }
            }

            System.out.println("Received " + count + " syslog event(s).");
        }
    }
}
