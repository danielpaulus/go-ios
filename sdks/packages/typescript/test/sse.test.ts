import { describe, expect, it } from "vitest";

import { SseFrameParser, parseSseStream, type SseEvent } from "../src/sse";
import type { SyslogEventMap } from "../src/events";

const enc = new TextEncoder();

/** Build a ReadableStream that emits the given byte chunks in order. */
function streamOf(chunks: Uint8Array[]): ReadableStream<Uint8Array> {
  let i = 0;
  return new ReadableStream({
    pull(controller) {
      if (i < chunks.length) {
        controller.enqueue(chunks[i++]!);
      } else {
        controller.close();
      }
    },
  });
}

async function collect<T>(it: AsyncIterable<T>): Promise<T[]> {
  const out: T[] = [];
  for await (const v of it) out.push(v);
  return out;
}

describe("SseFrameParser", () => {
  it("parses multiple frames from one chunk", () => {
    const p = new SseFrameParser();
    const frames = p.push(
      "event: syslog\ndata: {\"message\":\"a\"}\n\nevent: syslog\ndata: {\"message\":\"b\"}\n\n",
    );
    expect(frames).toEqual([
      { event: "syslog", data: '{"message":"a"}', id: undefined },
      { event: "syslog", data: '{"message":"b"}', id: undefined },
    ]);
  });

  it("buffers a frame split across chunks", () => {
    const p = new SseFrameParser();
    expect(p.push("event: sys")).toEqual([]);
    expect(p.push("log\ndata: {\"message\"")).toEqual([]);
    const frames = p.push(':"hi"}\n\n');
    expect(frames).toEqual([{ event: "syslog", data: '{"message":"hi"}', id: undefined }]);
  });

  it("concatenates multiple data: lines with newlines", () => {
    const p = new SseFrameParser();
    const frames = p.push("event: syslog\ndata: line1\ndata: line2\n\n");
    expect(frames[0]!.data).toBe("line1\nline2");
  });

  it("defaults the event name to 'message' when omitted", () => {
    const p = new SseFrameParser();
    const frames = p.push("data: {}\n\n");
    expect(frames[0]!.event).toBe("message");
  });

  it("handles CRLF line endings", () => {
    const p = new SseFrameParser();
    const frames = p.push("event: syslog\r\ndata: {\"message\":\"x\"}\r\n\r\n");
    expect(frames).toEqual([{ event: "syslog", data: '{"message":"x"}', id: undefined }]);
  });

  it("flushes a trailing unterminated frame", () => {
    const p = new SseFrameParser();
    expect(p.push("event: syslog\ndata: {\"message\":\"last\"}")).toEqual([]);
    expect(p.flush()).toEqual([{ event: "syslog", data: '{"message":"last"}', id: undefined }]);
  });
});

describe("parseSseStream", () => {
  it("yields typed events and JSON-parses data", async () => {
    const stream = streamOf([
      enc.encode('event: syslog\ndata: {"message":"hello","timestamp":1}\n\n'),
    ]);
    const events = await collect(parseSseStream<SyslogEventMap>(stream));
    expect(events).toEqual([{ event: "syslog", data: { message: "hello", timestamp: 1 } }]);
  });

  it("reassembles a frame split across byte chunks", async () => {
    const full = 'event: syslog\ndata: {"message":"split"}\n\n';
    const bytes = enc.encode(full);
    const mid = 15;
    const stream = streamOf([bytes.slice(0, mid), bytes.slice(mid)]);
    const events = await collect(parseSseStream<SyslogEventMap>(stream));
    expect(events).toEqual([{ event: "syslog", data: { message: "split" } }]);
  });

  it("drops heartbeat events but keeps surrounding events", async () => {
    const stream = streamOf([
      enc.encode(
        'event: heartbeat\ndata: {}\n\n' +
          'event: syslog\ndata: {"message":"after-hb"}\n\n' +
          "event: heartbeat\ndata: {}\n\n",
      ),
    ]);
    const events = await collect(parseSseStream<SyslogEventMap>(stream));
    expect(events).toEqual([{ event: "syslog", data: { message: "after-hb" } }]);
  });

  it("surfaces unknown event types rather than dropping them", async () => {
    const stream = streamOf([
      enc.encode('event: future_thing\ndata: {"x":1}\n\n'),
    ]);
    const events = await collect(parseSseStream<SyslogEventMap>(stream));
    expect(events).toEqual([{ event: "future_thing", data: { x: 1 } }]);
  });

  it("falls back to the raw string when data is not JSON", async () => {
    const stream = streamOf([enc.encode("event: syslog\ndata: not-json\n\n")]);
    const events = await collect(parseSseStream<SyslogEventMap>(stream));
    expect(events).toEqual([{ event: "syslog", data: "not-json" }]);
  });

  it("stops when the AbortSignal fires", async () => {
    const controller = new AbortController();
    let pulls = 0;
    const stream = new ReadableStream<Uint8Array>({
      pull(c) {
        pulls++;
        c.enqueue(enc.encode('event: syslog\ndata: {"message":"tick"}\n\n'));
        if (pulls >= 2) controller.abort();
      },
    });
    const received: SseEvent<SyslogEventMap>[] = [];
    for await (const ev of parseSseStream<SyslogEventMap>(stream, {
      signal: controller.signal,
    })) {
      received.push(ev);
      if (received.length >= 2) {
        // abort was requested inside pull; loop should terminate shortly.
      }
    }
    expect(received.length).toBeGreaterThanOrEqual(1);
    expect(controller.signal.aborted).toBe(true);
  });

  it("returns immediately when the signal is already aborted", async () => {
    const controller = new AbortController();
    controller.abort();
    const stream = streamOf([enc.encode('event: syslog\ndata: {"message":"x"}\n\n')]);
    const events = await collect(
      parseSseStream<SyslogEventMap>(stream, { signal: controller.signal }),
    );
    expect(events).toEqual([]);
  });
});
