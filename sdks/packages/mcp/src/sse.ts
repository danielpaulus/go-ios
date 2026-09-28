/**
 * Server-Sent Events parsing + bounded capture.
 *
 * The go-ios daemon emits real SSE frames (see docs/DESIGN.md):
 *
 *   event: <event-name>\n
 *   data: <compact-json-of-payload>\n
 *   \n
 *
 * A `heartbeat` event ({}) is emitted on idle. Streams never terminate on their
 * own — they run until the client disconnects. Because an MCP tool call is
 * request/response (an agent cannot hold an infinite stream), `captureStream`
 * reads such a stream for a bounded window and returns the collected events.
 */

export interface SseEvent {
  /** The SSE `event:` name (defaults to "message" if the frame omits it). */
  event: string;
  /** Raw `data:` payload (concatenated across multiple data lines). */
  data: string;
}

/**
 * Incremental SSE frame parser. Feed it decoded string chunks; it yields one
 * SseEvent per complete frame (terminated by a blank line). Robust to frames
 * split across chunk boundaries and to CRLF or LF line endings.
 */
export class SseParser {
  private buffer = "";
  private eventName: string | undefined;
  private dataLines: string[] = [];

  /** Push a chunk of text and return any events completed by it. */
  push(chunk: string): SseEvent[] {
    this.buffer += chunk;
    const events: SseEvent[] = [];
    let idx: number;
    // Process complete lines (split on \n; tolerate trailing \r).
    while ((idx = this.buffer.indexOf("\n")) !== -1) {
      let line = this.buffer.slice(0, idx);
      this.buffer = this.buffer.slice(idx + 1);
      if (line.endsWith("\r")) line = line.slice(0, -1);

      if (line === "") {
        // Blank line: dispatch the accumulated frame, if any.
        const evt = this.flush();
        if (evt) events.push(evt);
        continue;
      }
      if (line.startsWith(":")) {
        // Comment line — ignore.
        continue;
      }
      const colon = line.indexOf(":");
      const field = colon === -1 ? line : line.slice(0, colon);
      // Per spec, a single leading space after the colon is stripped.
      let value = colon === -1 ? "" : line.slice(colon + 1);
      if (value.startsWith(" ")) value = value.slice(1);

      if (field === "event") {
        this.eventName = value;
      } else if (field === "data") {
        this.dataLines.push(value);
      }
      // id/retry fields are ignored — not used by the go-ios contract.
    }
    return events;
  }

  private flush(): SseEvent | undefined {
    if (this.eventName === undefined && this.dataLines.length === 0) {
      return undefined;
    }
    const evt: SseEvent = {
      event: this.eventName ?? "message",
      data: this.dataLines.join("\n"),
    };
    this.eventName = undefined;
    this.dataLines = [];
    return evt;
  }
}

/**
 * Open an SSE endpoint and capture a bounded window of events, sharing the
 * auth/URL plumbing of the REST client. Used by every "collect a finite slice
 * of an infinite stream" tool (device logs, CPU samples, job logs) so they all
 * enforce the same duration/line caps and abort the request when the window
 * ends. `fetchFn` performs the actual request (given an AbortSignal); it must
 * throw for non-2xx and return a Response with a body.
 */
export async function captureSse(
  fetchFn: (signal: AbortSignal) => Promise<Response>,
  opts: Omit<CaptureOptions, "signal">,
): Promise<CaptureResult> {
  const controller = new AbortController();
  const res = await fetchFn(controller.signal);
  if (!res.body) {
    throw new Error("SSE endpoint returned no body.");
  }
  return captureStream(res.body, { ...opts, signal: controller.signal });
}

export interface CaptureOptions {
  /** Stop after this many wall-clock milliseconds. */
  durationMs: number;
  /** Stop after collecting this many matching (non-heartbeat) events. */
  maxLines: number;
  /**
   * Event names to keep. Heartbeats are always dropped from the result but are
   * counted for the "heartbeats" stat so an idle-but-live stream is visible.
   */
  keepEvents?: Set<string>;
  /** Optional AbortSignal to cancel the underlying request when the window ends. */
  signal?: AbortSignal;
}

export interface CapturedEvent {
  event: string;
  /** Parsed JSON payload if `data` was valid JSON; otherwise undefined. */
  payload?: unknown;
  /** Raw data string (always present). */
  raw: string;
}

export interface CaptureResult {
  events: CapturedEvent[];
  heartbeats: number;
  /** Why capture stopped. */
  stoppedBy: "duration" | "maxLines" | "streamEnd";
  /** Total non-heartbeat events seen (equals events.length unless truncated). */
  totalMatched: number;
}

/**
 * Consume a ReadableStream of SSE bytes for a bounded window and return the
 * collected events. Stops at whichever bound is hit first: durationMs elapsed,
 * maxLines matching events collected, or the stream ending on its own.
 *
 * This is the "bounded capture, not an infinite stream" guarantee: the returned
 * promise always resolves within ~durationMs even if the device keeps emitting.
 */
export async function captureStream(
  stream: ReadableStream<Uint8Array>,
  opts: CaptureOptions,
): Promise<CaptureResult> {
  const parser = new SseParser();
  const decoder = new TextDecoder();
  const events: CapturedEvent[] = [];
  let heartbeats = 0;
  let totalMatched = 0;
  let stoppedBy: CaptureResult["stoppedBy"] = "streamEnd";
  // Set once the duration cap fires, so the read loop doesn't misreport the
  // resulting cancel() as a natural stream end.
  let durationHit = false;

  const reader = stream.getReader();
  const deadline = Date.now() + opts.durationMs;

  const cancel = () => {
    reader.cancel().catch(() => {});
  };

  // Hard duration cap: fires even if the stream never sends another byte.
  const timer = setTimeout(() => {
    durationHit = true;
    stoppedBy = "duration";
    cancel();
  }, opts.durationMs);
  if (opts.signal) {
    opts.signal.addEventListener("abort", cancel, { once: true });
  }

  const handle = (evt: SseEvent): boolean => {
    if (evt.event === "heartbeat") {
      heartbeats += 1;
      return false;
    }
    if (opts.keepEvents && !opts.keepEvents.has(evt.event)) {
      return false;
    }
    totalMatched += 1;
    if (events.length < opts.maxLines) {
      let payload: unknown;
      try {
        payload = evt.data ? JSON.parse(evt.data) : undefined;
      } catch {
        payload = undefined;
      }
      events.push({ event: evt.event, payload, raw: evt.data });
    }
    // Signal caller to stop once we've filled the buffer.
    return events.length >= opts.maxLines;
  };

  try {
    for (;;) {
      if (Date.now() >= deadline) {
        stoppedBy = "duration";
        break;
      }
      const { done, value } = await reader.read();
      if (done) {
        // A cancel() from the duration timer also surfaces here as done=true;
        // don't overwrite the already-recorded "duration" reason.
        if (!durationHit) stoppedBy = "streamEnd";
        break;
      }
      const chunk = decoder.decode(value, { stream: true });
      let full = false;
      for (const evt of parser.push(chunk)) {
        if (handle(evt)) {
          full = true;
          break;
        }
      }
      if (full) {
        stoppedBy = "maxLines";
        break;
      }
    }
  } catch {
    // A cancel() during read() rejects the pending read; treat as a clean stop.
    // stoppedBy already reflects the reason (duration/abort) at this point.
  } finally {
    clearTimeout(timer);
    if (opts.signal) opts.signal.removeEventListener("abort", cancel);
    cancel();
  }

  return { events, heartbeats, stoppedBy, totalMatched };
}
