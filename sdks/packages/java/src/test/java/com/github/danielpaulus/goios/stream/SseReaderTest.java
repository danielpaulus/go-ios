package com.github.danielpaulus.goios.stream;

import com.github.danielpaulus.goios.generated.model.SyslogMessage;
import org.junit.jupiter.api.Test;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.Iterator;
import java.util.List;
import java.util.NoSuchElementException;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicBoolean;

import static org.junit.jupiter.api.Assertions.*;

/** Unit tests for the SSE frame parser over canned event-stream text. */
class SseReaderTest {

    /** Split a canned event-stream body into lines the way ofLines() would. */
    private static Iterator<String> lines(String body) {
        // Note: no trailing empty element even if body ends with \n (matches BufferedReader/ofLines).
        return Arrays.asList(body.split("\n", -1)).iterator();
    }

    private static List<SseEvent> drain(SseReader reader) {
        List<SseEvent> out = new ArrayList<>();
        reader.forEachRemaining(out::add);
        return out;
    }

    @Test
    void parsesMultipleFramesAndSkipsHeartbeatByDefault() {
        String body = """
                event: syslog
                data: {"message":"line one","timestamp":1723200000000}

                event: heartbeat
                data: {}

                event: syslog
                data: {"message":"line two"}

                """;
        try (SseReader r = new SseReader(lines(body), EventDecoder.SYSLOG, false, null)) {
            List<SseEvent> events = drain(r);
            assertEquals(2, events.size(), "heartbeat should be skipped");
            assertInstanceOf(SyslogEvent.class, events.get(0));
            assertEquals("line one", ((SyslogEvent) events.get(0)).payload().getMessage());
            assertEquals("line two", ((SyslogEvent) events.get(1)).payload().getMessage());
        }
    }

    @Test
    void includesHeartbeatWhenRequested() {
        String body = "event: heartbeat\ndata: {}\n\nevent: syslog\ndata: {\"message\":\"x\"}\n\n";
        try (SseReader r = new SseReader(lines(body), EventDecoder.SYSLOG, true, null)) {
            List<SseEvent> events = drain(r);
            assertEquals(2, events.size());
            assertInstanceOf(HeartbeatEvent.class, events.get(0));
            assertEquals("heartbeat", events.get(0).eventName());
            assertInstanceOf(SyslogEvent.class, events.get(1));
        }
    }

    @Test
    void multiLineDataFramesAreJoinedWithNewline() {
        // Two data: lines in one frame -> joined by \n into valid JSON.
        String body = "event: syslog\ndata: {\"message\":\ndata: \"multi\"}\n\n";
        try (SseReader r = new SseReader(lines(body), EventDecoder.SYSLOG, false, null)) {
            List<SseEvent> events = drain(r);
            assertEquals(1, events.size());
            assertEquals("multi", ((SyslogEvent) events.get(0)).payload().getMessage());
        }
    }

    @Test
    void commentLinesAreIgnored() {
        String body = ": this is a keep-alive comment\nevent: syslog\ndata: {\"message\":\"ok\"}\n\n";
        try (SseReader r = new SseReader(lines(body), EventDecoder.SYSLOG, false, null)) {
            List<SseEvent> events = drain(r);
            assertEquals(1, events.size());
            assertEquals("ok", ((SyslogEvent) events.get(0)).payload().getMessage());
        }
    }

    @Test
    void unknownEventIsSurfacedNotDropped() {
        String body = "event: brandnew\ndata: {\"foo\":42}\n\nevent: syslog\ndata: {\"message\":\"y\"}\n\n";
        try (SseReader r = new SseReader(lines(body), EventDecoder.SYSLOG, false, null)) {
            List<SseEvent> events = drain(r);
            assertEquals(2, events.size());
            assertInstanceOf(UnknownEvent.class, events.get(0));
            UnknownEvent u = (UnknownEvent) events.get(0);
            assertEquals("brandnew", u.eventName());
            assertEquals("{\"foo\":42}", u.rawData());
        }
    }

    @Test
    void trailingFrameWithoutBlankLineIsDispatchedAtEndOfStream() {
        String body = "event: syslog\ndata: {\"message\":\"last\"}"; // no terminating blank line
        try (SseReader r = new SseReader(lines(body), EventDecoder.SYSLOG, false, null)) {
            List<SseEvent> events = drain(r);
            assertEquals(1, events.size());
            assertEquals("last", ((SyslogEvent) events.get(0)).payload().getMessage());
        }
    }

    @Test
    void closeInvokesHookAndStopsIteration() {
        AtomicBoolean closed = new AtomicBoolean(false);
        String body = "event: syslog\ndata: {\"message\":\"a\"}\n\nevent: syslog\ndata: {\"message\":\"b\"}\n\n";
        SseReader r = new SseReader(lines(body), EventDecoder.SYSLOG, false, () -> closed.set(true));
        assertTrue(r.hasNext());
        assertEquals("a", ((SyslogEvent) r.next()).payload().getMessage());
        r.close();
        assertTrue(closed.get(), "onClose hook should run");
        assertFalse(r.hasNext(), "closed reader yields no more events");
        assertThrows(NoSuchElementException.class, r::next);
    }

    @Test
    void closeIsIdempotent() {
        int[] count = {0};
        SseReader r = new SseReader(lines(""), EventDecoder.SYSLOG, false, () -> count[0]++);
        r.close();
        r.close();
        assertEquals(1, count[0], "onClose runs exactly once");
    }

    @Test
    void cancelFromAnotherThreadUnblocksABlockingSource() throws Exception {
        // A line source that blocks on hasNext() until the reader is closed,
        // simulating a live-but-idle SSE connection being cancelled.
        CountDownLatch blocking = new CountDownLatch(1);
        AtomicBoolean aborted = new AtomicBoolean(false);
        Iterator<String> blockingLines = new Iterator<>() {
            @Override
            public boolean hasNext() {
                blocking.countDown();
                try {
                    // Block until interrupted by close()'s abort semantics.
                    Thread.sleep(60_000);
                } catch (InterruptedException e) {
                    aborted.set(true);
                    Thread.currentThread().interrupt();
                    throw new RuntimeException("aborted", e);
                }
                return false;
            }

            @Override
            public String next() {
                throw new NoSuchElementException();
            }
        };

        SseReader r = new SseReader(blockingLines, EventDecoder.SYSLOG, false, () -> { });
        Thread consumer = new Thread(() -> {
            // hasNext() blocks in the source until interrupted.
            boolean has = r.hasNext();
            assertFalse(has, "after abort, hasNext returns false");
        });
        consumer.start();
        assertTrue(blocking.await(2, TimeUnit.SECONDS), "consumer reached the blocking source");
        r.close();
        consumer.interrupt(); // mimic transport abort waking the blocked read
        consumer.join(5_000);
        assertFalse(consumer.isAlive(), "consumer thread should finish after close+interrupt");
    }

    @Test
    void decodesEachEndpointPayloadType() {
        assertInstanceOf(AppStateEvent.class,
                EventDecoder.NOTIFICATIONS.apply("appstate",
                        "{\"bundleId\":\"com.apple.Preferences\",\"state\":\"foreground\",\"timestamp\":1}"));
        assertInstanceOf(OsTraceEvent.class,
                EventDecoder.OSTRACE.apply("ostrace", "{\"message\":\"m\",\"pid\":1}"));
        assertInstanceOf(AttachDetachEvent.class,
                EventDecoder.LISTEN.apply("attachdetach", "{\"event\":\"attached\",\"deviceID\":5}"));
        SyslogMessage sm = ((SyslogEvent) EventDecoder.SYSLOG.apply("syslog", "{\"message\":\"hi\"}")).payload();
        assertEquals("hi", sm.getMessage());
    }
}
