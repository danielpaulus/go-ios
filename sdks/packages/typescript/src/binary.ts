import { IosApiError } from "./errors";
import type { GenericResponse } from "./generated/types.gen";

/**
 * A consumable binary byte stream returned by the go-ios binary streaming
 * endpoints (`ui.stream`, `screenshotStream`, `pcap`). These are **raw byte
 * streams** (`application/octet-stream`, `image/jpeg`, `pcap`) — not SSE — so
 * they get their own helper instead of going through the SSE frame parser.
 *
 * A `BinaryStream` is an `AsyncIterable<Uint8Array>`: iterate it with
 * `for await (const chunk of stream)` to consume the chunks as they arrive.
 * It also exposes the raw {@link ReadableStream} via {@link body} for callers
 * who want to pipe it directly (e.g. into a file or a decoder). Consuming
 * either view consumes the underlying response body once.
 *
 * Cancellation: pass an `AbortSignal` to the originating call, or simply
 * `break` out of the `for await` loop — both cancel the underlying reader and
 * release the connection.
 */
export interface BinaryStream extends AsyncIterable<Uint8Array> {
  /** The raw response body, for callers who want to pipe it directly. */
  readonly body: ReadableStream<Uint8Array>;
  /** The response `Content-Type`, if the server set one. */
  readonly contentType: string | undefined;
}

/**
 * Shared driver for every binary-streaming endpoint: awaits the streaming
 * request, surfaces a non-2xx status as an {@link IosApiError} (parsing the
 * `GenericResponse` error envelope like the unary/SSE paths do), then wraps the
 * raw `Response.body` in a {@link BinaryStream}.
 *
 * The request must have been issued with `parseAs: "stream"` so the fetch
 * client hands back the live body rather than buffering it.
 */
export async function openBinaryStream(
  request: Promise<{ data?: unknown; error?: unknown; response: Response }>,
  signal?: AbortSignal,
): Promise<BinaryStream> {
  const { data, error, response } = await request;
  if (!response.ok || error !== undefined) {
    // The fetch client may already have parsed the error body (consuming it),
    // in which case it's on `error`/`data`; otherwise fall back to reading it.
    let body: unknown = error ?? data;
    if (body === undefined) {
      try {
        body = await response.json();
      } catch {
        /* non-JSON / already-consumed error body */
      }
    }
    const envelope = body as Partial<GenericResponse> | undefined;
    const message =
      envelope?.error ??
      envelope?.message ??
      `go-ios binary stream failed with status ${response.status}`;
    throw new IosApiError(response.status, String(message), body);
  }

  const stream = response.body ?? emptyStream();
  return new ResponseBinaryStream(
    stream,
    response.headers.get("content-type") ?? undefined,
    signal,
  );
}

/** An empty byte stream, used when a 2xx response carries no body. */
function emptyStream(): ReadableStream<Uint8Array> {
  return new ReadableStream<Uint8Array>({
    start(controller) {
      controller.close();
    },
  });
}

/**
 * Wraps a fetch `Response.body` (`ReadableStream<Uint8Array>`) as an
 * async-iterable byte stream with abort support. Kept separate from the SSE
 * parser: this does no framing/decoding, it just relays raw chunks.
 */
class ResponseBinaryStream implements BinaryStream {
  constructor(
    readonly body: ReadableStream<Uint8Array>,
    readonly contentType: string | undefined,
    private readonly signal?: AbortSignal,
  ) {}

  async *[Symbol.asyncIterator](): AsyncGenerator<Uint8Array, void, unknown> {
    const { body, signal } = this;
    if (signal?.aborted) {
      await body.cancel().catch(() => {});
      return;
    }

    const reader = body.getReader();
    const onAbort = () => {
      // Best-effort cancel; ignore rejection (reader may already be released).
      reader.cancel().catch(() => {});
    };
    signal?.addEventListener("abort", onAbort, { once: true });

    try {
      while (true) {
        const { done, value } = await reader.read();
        if (done) return;
        if (value) yield value;
      }
    } finally {
      signal?.removeEventListener("abort", onAbort);
      reader.releaseLock();
    }
  }
}
