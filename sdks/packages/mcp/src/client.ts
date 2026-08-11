/**
 * Thin fetch-based client for the go-ios REST daemon.
 *
 * We deliberately do NOT generate a full client from the OpenAPI spec here: the
 * MCP server only touches a curated handful of endpoints, and hand-writing the
 * calls keeps error surfacing precise (device-not-found, auth failure, etc.) and
 * lets the SSE streaming endpoints share the same auth/URL plumbing.
 */
import type { ServerConfig } from "./config.js";

export const API_PREFIX = "/api/v1";

/**
 * Error thrown for any non-2xx REST response, carrying the HTTP status and the
 * decoded `GenericResponse.error`/`message` when the daemon provided one.
 */
export class GoIosApiError extends Error {
  readonly status: number;
  readonly statusText: string;
  readonly url: string;
  readonly body?: unknown;

  constructor(args: {
    status: number;
    statusText: string;
    url: string;
    body?: unknown;
    message: string;
  }) {
    super(args.message);
    this.name = "GoIosApiError";
    this.status = args.status;
    this.statusText = args.statusText;
    this.url = args.url;
    this.body = args.body;
  }

  /** A short, LLM-friendly explanation of what went wrong. */
  get friendly(): string {
    switch (this.status) {
      case 401:
        return "Authentication failed (401). Set GO_IOS_API_KEY to a valid bearer token, or run the daemon with --disable-auth.";
      case 404:
        return "Device (udid) not found (404). Use list_devices to see connected udids.";
      case 422:
        return "Empty or invalid udid (422). Provide a non-empty udid.";
      case 423:
        return "Device is locked (423). Unlock the device and retry.";
      case 400:
        return `Malformed request (400): ${this.message}`;
      default:
        return `go-ios daemon returned ${this.status} ${this.statusText}: ${this.message}`;
    }
  }
}

export class GoIosClient {
  constructor(private readonly config: ServerConfig) {}

  get baseUrl(): string {
    return this.config.baseUrl;
  }

  /** Authorization/accept headers shared by REST and SSE requests. */
  headers(extra?: Record<string, string>): Record<string, string> {
    const h: Record<string, string> = { ...extra };
    if (this.config.apiKey) {
      h["Authorization"] = `Bearer ${this.config.apiKey}`;
    }
    return h;
  }

  /** Build an absolute URL for an API path (path is joined under /api/v1). */
  url(path: string, query?: Record<string, string | number | undefined>): string {
    const u = new URL(this.config.baseUrl + API_PREFIX + path);
    if (query) {
      for (const [k, v] of Object.entries(query)) {
        if (v !== undefined && v !== null && v !== "") {
          u.searchParams.set(k, String(v));
        }
      }
    }
    return u.toString();
  }

  private async request(
    method: string,
    path: string,
    opts: {
      query?: Record<string, string | number | undefined>;
      body?: unknown;
      accept?: string;
      contentType?: string;
    } = {},
  ): Promise<Response> {
    const url = this.url(path, opts.query);
    const headers = this.headers({ Accept: opts.accept ?? "application/json" });
    let body: string | undefined;
    if (opts.body !== undefined) {
      if (opts.contentType) {
        // Raw body (e.g. text/plain for the pasteboard endpoint).
        headers["Content-Type"] = opts.contentType;
        body = String(opts.body);
      } else {
        headers["Content-Type"] = "application/json";
        body = JSON.stringify(opts.body);
      }
    }

    let res: Response;
    try {
      res = await fetch(url, { method, headers, body });
    } catch (err) {
      throw new GoIosApiError({
        status: 0,
        statusText: "network error",
        url,
        message: `Could not reach the go-ios daemon at ${this.config.baseUrl}: ${
          err instanceof Error ? err.message : String(err)
        }`,
      });
    }

    if (!res.ok) {
      let parsed: unknown;
      let message = res.statusText;
      const text = await res.text().catch(() => "");
      if (text) {
        try {
          parsed = JSON.parse(text);
          const envelope = parsed as { error?: string; message?: string };
          message = envelope.error ?? envelope.message ?? text;
        } catch {
          message = text;
        }
      }
      throw new GoIosApiError({
        status: res.status,
        statusText: res.statusText,
        url,
        body: parsed,
        message,
      });
    }
    return res;
  }

  async getJson<T>(
    path: string,
    query?: Record<string, string | number | undefined>,
  ): Promise<T> {
    const res = await this.request("GET", path, { query });
    return (await res.json()) as T;
  }

  async postJson<T>(
    path: string,
    opts: { query?: Record<string, string | number | undefined>; body?: unknown } = {},
  ): Promise<T> {
    const res = await this.request("POST", path, opts);
    return (await res.json()) as T;
  }

  async deleteJson<T>(
    path: string,
    query?: Record<string, string | number | undefined>,
  ): Promise<T> {
    const res = await this.request("DELETE", path, { query });
    return (await res.json()) as T;
  }

  /**
   * PUT a raw text body (used for the pasteboard endpoint, which takes
   * `text/plain`). Returns the decoded JSON acknowledgement.
   */
  async putText<T>(
    path: string,
    text: string,
    query?: Record<string, string | number | undefined>,
  ): Promise<T> {
    const res = await this.request("PUT", path, {
      query,
      body: text,
      contentType: "text/plain",
    });
    return (await res.json()) as T;
  }

  /** GET raw bytes (used for the PNG screenshot endpoint). */
  async getBytes(path: string): Promise<{ bytes: Uint8Array; contentType: string }> {
    const res = await this.request("GET", path, { accept: "image/png" });
    const buf = new Uint8Array(await res.arrayBuffer());
    return { bytes: buf, contentType: res.headers.get("content-type") ?? "application/octet-stream" };
  }

  /**
   * GET a raw file body as UTF-8 text, bounded to `maxBytes`. Used for pulling
   * crash reports / small device files where returning the content to an agent
   * is useful but must be size-capped. Returns the (possibly truncated) text,
   * the number of bytes read, and whether truncation occurred.
   */
  async getTextBounded(
    path: string,
    query: Record<string, string | number | undefined>,
    maxBytes: number,
  ): Promise<{ text: string; bytes: number; truncated: boolean; contentType: string }> {
    const res = await this.request("GET", path, {
      query,
      accept: "application/octet-stream",
    });
    const contentType = res.headers.get("content-type") ?? "application/octet-stream";
    if (!res.body) {
      const buf = new Uint8Array(await res.arrayBuffer());
      const slice = buf.subarray(0, maxBytes);
      return {
        text: new TextDecoder().decode(slice),
        bytes: slice.length,
        truncated: buf.length > maxBytes,
        contentType,
      };
    }
    const reader = res.body.getReader();
    const chunks: Uint8Array[] = [];
    let total = 0;
    let truncated = false;
    try {
      for (;;) {
        const { done, value } = await reader.read();
        if (done) break;
        if (!value) continue;
        const remaining = maxBytes - total;
        if (value.length >= remaining) {
          chunks.push(value.subarray(0, remaining));
          total += remaining;
          truncated = true;
          break;
        }
        chunks.push(value);
        total += value.length;
      }
    } finally {
      reader.cancel().catch(() => {});
    }
    const merged = new Uint8Array(total);
    let off = 0;
    for (const c of chunks) {
      merged.set(c, off);
      off += c.length;
    }
    return {
      text: new TextDecoder().decode(merged),
      bytes: total,
      truncated,
      contentType,
    };
  }
}
