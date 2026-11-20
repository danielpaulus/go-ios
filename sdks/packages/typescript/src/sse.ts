/**
 * A single decoded Server-Sent Events frame.
 *
 * `event` is the SSE `event:` field (defaults to `"message"` per the SSE spec
 * when the server omits it). `data` is the concatenated `data:` field(s) with the
 * trailing newline stripped, still as a raw string — JSON parsing happens one
 * layer up in {@link parseSseStream} so this parser stays transport-only.
 */
export interface SseFrame {
  event: string;
  data: string;
  id?: string;
}

/**
 * Incremental SSE frame parser.
 *
 * Feed it arbitrary string chunks (already UTF-8 decoded) via {@link push}; it
 * buffers across chunk boundaries and yields whole frames only once their
 * terminating blank line has arrived. This makes it robust to a `text/event-stream`
 * body that splits a single frame across multiple network reads.
 *
 * Framing per the WHATWG SSE spec / DESIGN.md:
 *   event: <name>\n
 *   data: <payload>\n
 *   \n
 */
export class SseFrameParser {
  private buffer = "";

  /** Feed a decoded chunk; returns any frames completed by this chunk. */
  push(chunk: string): SseFrame[] {
    // Normalize CRLF / CR to LF so field/line splitting is uniform.
    this.buffer += chunk.replace(/\r\n?/g, "\n");
    const frames: SseFrame[] = [];
    let sep: number;
    // Frames are separated by a blank line (i.e. "\n\n").
    while ((sep = this.buffer.indexOf("\n\n")) !== -1) {
      const raw = this.buffer.slice(0, sep);
      this.buffer = this.buffer.slice(sep + 2);
      const frame = this.decodeFrame(raw);
      if (frame) frames.push(frame);
    }
    return frames;
  }

  /**
   * Flush any trailing frame not terminated by a blank line (e.g. the stream
   * ended cleanly right after a frame with no final newline). Most well-behaved
   * servers terminate every frame, so this is usually a no-op.
   */
  flush(): SseFrame[] {
    const rest = this.buffer.trim();
    this.buffer = "";
    if (!rest) return [];
    const frame = this.decodeFrame(rest);
    return frame ? [frame] : [];
  }

  private decodeFrame(raw: string): SseFrame | undefined {
    let event = "message";
    const dataLines: string[] = [];
    let id: string | undefined;
    let sawField = false;

    for (const line of raw.split("\n")) {
      if (line === "" || line.startsWith(":")) {
        // Blank line inside a frame can't happen (that's the separator) and
        // lines starting with ':' are comments — skip both.
        continue;
      }
      const colon = line.indexOf(":");
      const field = colon === -1 ? line : line.slice(0, colon);
      // Per spec a single leading space after the colon is stripped.
      let value = colon === -1 ? "" : line.slice(colon + 1);
      if (value.startsWith(" ")) value = value.slice(1);

      switch (field) {
        case "event":
          event = value;
          sawField = true;
          break;
        case "data":
          dataLines.push(value);
          sawField = true;
          break;
        case "id":
          id = value;
          sawField = true;
          break;
        default:
          // Unknown field ("retry" etc.) — ignored, but still a real frame.
          sawField = true;
      }
    }

    if (!sawField) return undefined;
    return { event, data: dataLines.join("\n"), id };
  }
}

/**
 * A typed SSE event surfaced by {@link parseSseStream}.
 *
 * - For events named in `eventMap`, `data` is the JSON-parsed payload of that
 *   variant's model (the caller narrows via `event`).
 * - For any other (unknown / forward-compat) event, `data` is the parsed JSON
 *   value (or the raw string if it wasn't valid JSON). Unknown events are
 *   surfaced rather than dropped, per the DESIGN.md contract.
 */
export type KnownSseEvent<TMap> = {
  [K in keyof TMap]: { event: K; data: TMap[K] };
}[keyof TMap];

/**
 * The forward-compat branch of {@link SseEvent}: an event whose name is not in
 * the map (a newer server event this SDK doesn't yet model). Surfaced rather
 * than dropped so callers can log/handle it.
 */
export interface UnknownSseEvent {
  event: string;
  data: unknown;
}

export type SseEvent<TMap> = KnownSseEvent<TMap> | UnknownSseEvent;

/**
 * Type guard that narrows an {@link SseEvent} to a specific known event by name.
 * TypeScript cannot narrow a union with an open `string` member on a bare
 * `ev.event === "name"` check, so use this to get the typed `data`:
 *
 * ```ts
 * if (isSseEvent(ev, "syslog")) console.log(ev.data.message); // data: SyslogMessage
 * ```
 */
export function isSseEvent<TMap, K extends keyof TMap & string>(
  ev: SseEvent<TMap>,
  name: K,
): ev is Extract<KnownSseEvent<TMap>, { event: K }> {
  return ev.event === name;
}

/**
 * Turns a `text/event-stream` `Response.body` into an async iterable of typed
 * events.
 *
 * - `heartbeat` frames are consumed to keep the connection alive but never
 *   yielded (they carry an empty `{}` payload).
 * - Each non-heartbeat frame's `data` is `JSON.parse`d; malformed JSON falls
 *   back to the raw string so a single bad frame can't kill the stream.
 * - Cancellation: pass an `AbortSignal`; aborting cancels the underlying reader
 *   and the iterator returns. Breaking out of the `for await` also releases the
 *   reader (via the async iterator's `return`).
 */
export async function* parseSseStream<TMap>(
  body: ReadableStream<Uint8Array>,
  options: { signal?: AbortSignal; eventNames?: ReadonlySet<string> } = {},
): AsyncGenerator<SseEvent<TMap>, void, unknown> {
  const { signal } = options;
  if (signal?.aborted) return;

  const reader = body.getReader();
  const decoder = new TextDecoder();
  const parser = new SseFrameParser();

  const onAbort = () => {
    // Best-effort cancel; ignore rejection (reader may already be released).
    reader.cancel().catch(() => {});
  };
  signal?.addEventListener("abort", onAbort, { once: true });

  try {
    while (true) {
      const { done, value } = await reader.read();
      if (done) {
        for (const frame of parser.flush()) {
          const ev = toEvent<TMap>(frame);
          if (ev) yield ev;
        }
        return;
      }
      const frames = parser.push(decoder.decode(value, { stream: true }));
      for (const frame of frames) {
        const ev = toEvent<TMap>(frame);
        if (ev) yield ev;
      }
    }
  } finally {
    signal?.removeEventListener("abort", onAbort);
    reader.releaseLock();
  }
}

function toEvent<TMap>(
  frame: SseFrame,
): SseEvent<TMap> | undefined {
  if (frame.event === "heartbeat") return undefined;
  let data: unknown = frame.data;
  if (frame.data.length > 0) {
    try {
      data = JSON.parse(frame.data);
    } catch {
      // Keep the raw string on parse failure rather than dropping the frame.
      data = frame.data;
    }
  }
  return { event: frame.event, data } as SseEvent<TMap>;
}
