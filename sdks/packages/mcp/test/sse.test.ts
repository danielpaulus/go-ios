import { describe, it, expect } from "vitest";
import { SseParser, captureStream } from "../src/sse.js";

/** Build a ReadableStream that emits the given byte chunks, optionally spaced in time. */
function streamFromChunks(chunks: string[], gapMs = 0): ReadableStream<Uint8Array> {
  const enc = new TextEncoder();
  let i = 0;
  return new ReadableStream<Uint8Array>({
    async pull(controller) {
      if (i >= chunks.length) {
        controller.close();
        return;
      }
      if (gapMs) await new Promise((r) => setTimeout(r, gapMs));
      controller.enqueue(enc.encode(chunks[i++]!));
    },
  });
}

/** A stream that emits `count` syslog frames then blocks forever (simulates an infinite stream). */
function infiniteSyslogStream(count: number): ReadableStream<Uint8Array> {
  const enc = new TextEncoder();
  let i = 0;
  return new ReadableStream<Uint8Array>({
    async pull(controller) {
      if (i < count) {
        i++;
        controller.enqueue(
          enc.encode(
            `event: syslog\ndata: ${JSON.stringify({ message: `line ${i}`, timestamp: i })}\n\n`,
          ),
        );
      } else {
        // Never resolve -> stream is "infinite". captureStream must still return.
        await new Promise(() => {});
      }
    },
  });
}

describe("SseParser", () => {
  it("parses a single frame", () => {
    const p = new SseParser();
    const evts = p.push("event: syslog\ndata: {\"message\":\"hi\"}\n\n");
    expect(evts).toHaveLength(1);
    expect(evts[0]).toEqual({ event: "syslog", data: '{"message":"hi"}' });
  });

  it("handles frames split across chunk boundaries", () => {
    const p = new SseParser();
    expect(p.push("event: syslog\nda")).toHaveLength(0);
    expect(p.push("ta: {\"a\":1}")).toHaveLength(0);
    const evts = p.push("\n\n");
    expect(evts).toHaveLength(1);
    expect(evts[0]!.data).toBe('{"a":1}');
  });

  it("strips a single leading space after the colon and tolerates CRLF", () => {
    const p = new SseParser();
    const evts = p.push("event: heartbeat\r\ndata: {}\r\n\r\n");
    expect(evts[0]).toEqual({ event: "heartbeat", data: "{}" });
  });

  it("ignores comment lines", () => {
    const p = new SseParser();
    const evts = p.push(": keepalive\nevent: syslog\ndata: x\n\n");
    expect(evts).toHaveLength(1);
    expect(evts[0]!.event).toBe("syslog");
  });

  it("concatenates multiple data lines", () => {
    const p = new SseParser();
    const evts = p.push("event: syslog\ndata: a\ndata: b\n\n");
    expect(evts[0]!.data).toBe("a\nb");
  });
});

describe("captureStream bounds", () => {
  it("stops at maxLines and parses JSON payloads", async () => {
    const frames = Array.from({ length: 10 }, (_, n) =>
      `event: syslog\ndata: ${JSON.stringify({ message: `m${n}`, timestamp: n })}\n\n`,
    );
    const res = await captureStream(streamFromChunks(frames), {
      durationMs: 5000,
      maxLines: 3,
      keepEvents: new Set(["syslog"]),
    });
    expect(res.stoppedBy).toBe("maxLines");
    expect(res.events).toHaveLength(3);
    expect(res.events[0]!.payload).toEqual({ message: "m0", timestamp: 0 });
  });

  it("drops heartbeats but counts them", async () => {
    const frames = [
      "event: heartbeat\ndata: {}\n\n",
      'event: syslog\ndata: {"message":"real"}\n\n',
      "event: heartbeat\ndata: {}\n\n",
    ];
    const res = await captureStream(streamFromChunks(frames), {
      durationMs: 5000,
      maxLines: 100,
      keepEvents: new Set(["syslog"]),
    });
    expect(res.stoppedBy).toBe("streamEnd");
    expect(res.heartbeats).toBe(2);
    expect(res.events).toHaveLength(1);
    expect(res.events[0]!.event).toBe("syslog");
  });

  it("returns within the duration cap even for an infinite stream", async () => {
    const start = Date.now();
    const res = await captureStream(infiniteSyslogStream(2), {
      durationMs: 200,
      maxLines: 1000,
      keepEvents: new Set(["syslog"]),
    });
    const elapsed = Date.now() - start;
    expect(res.stoppedBy).toBe("duration");
    // Collected the 2 available lines, then the duration cap fired.
    expect(res.events).toHaveLength(2);
    expect(elapsed).toBeLessThan(1500);
  });

  it("keeps only requested event types", async () => {
    const frames = [
      'event: ostrace\ndata: {"message":"trace"}\n\n',
      'event: syslog\ndata: {"message":"sys"}\n\n',
    ];
    const res = await captureStream(streamFromChunks(frames), {
      durationMs: 5000,
      maxLines: 100,
      keepEvents: new Set(["ostrace"]),
    });
    expect(res.events).toHaveLength(1);
    expect(res.events[0]!.event).toBe("ostrace");
  });
});
