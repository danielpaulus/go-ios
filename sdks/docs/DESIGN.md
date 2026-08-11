# go-ios SDKs — design & contracts

This document captures the locked decisions and the exact streaming contract that
every downstream phase (TypeScript/Python/Java/C# SDKs, MCP server, and the go-ios
server SSE PR) must implement consistently.

## Source of truth

- **IDL:** TypeSpec 1.14 → **OpenAPI 3.1** (`spec/*.tsp` → `spec/openapi/openapi.yaml`).
  The spec is the *ideal* contract; the go-ios REST server conforms to it. No
  backward-compatibility constraint yet.
- Downstream generators consume the canonical `spec/openapi/openapi.yaml` (3.1).
  A **3.2.0** variant (`spec/openapi/openapi.3.2.0.yaml`) is also emitted because
  3.2 is the only OpenAPI version that carries the typed SSE event schema inline
  (see "Streaming" below).

## Locked toolchain (OSS-first)

Commercial generators (Stainless/Speakeasy/Fern) were all rejected — no free/OSS
tier covers a streaming-heavy Py+TS+MCP project. Stack:

- **TypeScript SDK:** `@hey-api/openapi-ts` (MIT).
- **Python SDK:** `openapi-python-client` (httpx, sync+async).
- **Java SDK:** `openapi-generator` java client (Maven Central).
- **C# SDK:** `openapi-generator` csharp client (NuGet).
- **MCP server:** OSS openapi→mcp with **curated** tools (not naive 1:1).
- **Architecture:** generated client + thin hand-written ergonomic facade with an
  *identical public API across languages*, generator-agnostic (so a commercial
  gen could swap in later).
- **Packaging:** Python uv + PyPI OIDC; TS tsup + changesets + npm provenance;
  Java Maven Central; C# NuGet.

## Authentication

- Bearer auth (`@useAuth(BearerAuth)`) on the whole `/api/v1` group:
  `Authorization: Bearer <GO_IOS_API_KEY>`.
- The server **refuses to start** with no key and no `--disable-auth`.
- When started with **`--disable-auth`**, auth is not enforced and the header may
  be omitted. SDKs should make the token optional but strongly encouraged, and
  MUST send it when present. (`--disable-auth` is a server launch flag, not part
  of the wire contract, so it is only noted here, not modeled as a scheme.)
- The Swagger UI (`/swagger/*`) is unauthenticated and outside `/api/v1` — not
  modeled in the spec.

## Base path & device routing

- Base path: `/api/v1`.
- Device-scoped routes: `/device/{udid}/...`. Middleware resolves the udid:
  unknown udid → **404**, empty udid → **422**.
- `/device/{udid}/apps/*` is serialized per-udid (one concurrent request per
  device) — a server behavior, not expressed in the schema.

## Error model (consistent)

All errors use the `GenericResponse` envelope (`{ message?, error? }`) with proper
status codes. Shared across device routes:

- **401** unauthorized (missing/invalid token when auth enabled)
- **404** device not found (and, for WDA session routes, unknown session)
- **422** empty/invalid udid
- **400** malformed request (missing required query/body)
- **423** device locked (pairing only)
- **500** internal/device error

## Deviations from the current go-ios API (baked into the spec)

The spec fixes these warts; the companion go-ios server PR must conform:

1. **`longtitude` → `longitude`.** `PUT /setlocation` uses `longitude` (correct
   spelling). Server should accept `longitude`, optionally keeping `longtitude`
   as a deprecated alias.
2. **Screenshot media type.** `GET /screenshot` returns `image/png` (was
   mislabeled `application/octet-stream`).
3. **Consistent errors.** All error bodies are `GenericResponse` with the status
   codes above (previously a mix of `gin.H` shapes).
4. **Real SSE.** The four streaming endpoints emit real `text/event-stream`
   frames (today they emit NDJSON or, worse, concatenated JSON with no delimiter
   for `/syslog` and `/listen`). See below.

## Endpoint surface (spec v2 — full daemon parity)

Spec v2 models the **complete** go-ios REST daemon surface (PR #817
`feature/restapi-parity`): **80 operations across 65 paths** under `/api/v1`,
grouped as below. This matches the daemon's registered routes 1:1, with three
deliberate exclusions (see "Excluded routes").

| Group                | Routes | Notes |
| -------------------- | ------ | ----- |
| Global               | `GET /list` | device listing |
| Tunnel agent         | `GET /tunnels`, `DELETE /tunnels/{udid}`, `POST /tunnels/{udid}/refresh`, `POST /tunnel-agent/shutdown` | not device-scoped; talk to the tunnel agent by udid. `502 BadGateway` when the agent is unreachable. |
| Device lifecycle     | `activate`, `info`, `screenshot`, `setlocation`, `resetlocation`, `resetaccessibility`, `pair` | (in `routes.tsp`) |
| Developer image      | `GET/PUT/DELETE image`, `GET image/list` | mount (raw body or `?auto=true`), list, unmount |
| Conditions           | `GET conditions`, `PUT enable-condition`, `POST disable-condition` | condition inducers |
| WebDriverAgent       | `POST wda/session`, `GET/DELETE wda/session/{sessionId}` | interactive session (distinct from the `runwda` job) |
| Apps                 | `GET apps/`, `POST apps/launch|kill|install|uninstall` | serialized per-udid |
| Device info (RO)     | `devicename`, `date`, `battery`, `diagnostics`, `mobilegestalt`, `processes`, `lockdown` | `routes-deviceinfo.tsp` |
| Device management    | `reboot`, `shutdown`, `erase` (`?confirm=true`), `GET/POST devmode`, `GET/PUT lang`, `memlimitoff` | `routes-devicemgmt.tsp` |
| Files & crashes      | `GET files`, `GET files/pull` (octet-stream), `POST files/push` (raw body), `GET/DELETE crashes` | `routes-files.tsp`; `domain` = `app|app-group|crash|temp` |
| Media                | `GET/PUT wallpaper`, `GET/PUT icon-layout`, `GET/PUT pasteboard` | `routes-media.tsp`; wallpaper is supervised (multipart) |
| Config / profiles    | `POST profiles`, `DELETE profiles/{name}` | `routes-config.tsp` (`GET profiles`, `GET/PUT image` are in `routes.tsp`) |
| Settings             | `GET/PUT assistivetouch`, `GET/PUT timeformat`, `PUT/DELETE wifi` | `routes-settings.tsp` |
| Monitoring           | `GET sysmontap` (SSE) | `routes-monitoring.tsp`; `pcap` not exposed yet |
| MDM (supervised)     | `POST mdm/security-info|fetch-unlock-token|clear-passcode|clear-screen-time-password` | `routes-mdm.tsp`; each takes a `p12` multipart identity |
| Proxy (supervised)   | `PUT/DELETE httpproxy` | `routes-proxy.tsp` |
| Async jobs           | `POST jobs/runtest|runwda|forward`, `GET jobs`, `GET jobs/{id}`, `GET jobs/{id}/logs` (SSE), `DELETE jobs/{id}` | `routes-jobs.tsp` |

### Async jobs subsystem

Long-running operations (test runs, the WDA runner, port forwards) run in the
background as **jobs**. `POST /jobs/{runtest,runwda,forward}` returns **202** with
a `Job` (`id`, `kind`, `udid`, `status`, `startedAt`, `finishedAt?`, `error?`,
`result?`; `status` ∈ `running|succeeded|failed|stopped`). Poll `GET /jobs/{id}`,
list with `GET /jobs`, stream output with `GET /jobs/{id}/logs` (SSE, see below),
and `DELETE /jobs/{id}` to stop a running job or purge a terminal one. Unknown
job ids yield **404**.

Note: `POST /wda/session` (interactive WDA session, `routes.tsp`) and
`POST /jobs/runwda` (async WDA runner job) are distinct surfaces that both launch
WebDriverAgent — the former returns a `WdaSession`, the latter a `Job`.

### Excluded routes

Three routes registered by the server are intentionally **not** modeled:
`GET /healthz` and `GET /readyz` (unauthenticated probes outside `/api/v1`) and
`GET /swagger/*any` (the Swagger UI). None are part of the JSON wire contract.

## Streaming contract (Server-Sent Events)

The six long-lived endpoints are modeled with `@typespec/sse`'s `SSEStream<T>`,
which sets the response content-type to **`text/event-stream`**. Each stream's
`T` is an `@events` union: **each named union variant becomes the SSE `event:`
name**, and the variant's model is the JSON payload of that event's `data:` frame.

### Wire framing (server PR must emit exactly this)

```
event: <event-name>\n
data: <compact-json-of-payload>\n
\n
```

- One event per frame, terminated by a blank line.
- `data:` is compact (single-line) JSON of the payload model.
- A `heartbeat` event (empty JSON object `{}`) is sent on an idle interval on
  **every** stream, so clients can distinguish a live-but-idle connection from a
  dropped one and keep-alives are self-describing.
- There is no terminal event; streams run until the client disconnects or the
  device goes away.

### Endpoints and their event unions

| Endpoint                          | Event union         | `event:` name → payload model                 |
| --------------------------------- | ------------------- | --------------------------------------------- |
| `GET /device/{udid}/notifications`| `NotificationEvents`| `appstate` → `AppStateNotification`; `heartbeat` → `Heartbeat` |
| `GET /device/{udid}/syslog`       | `SyslogEvents`      | `syslog` → `SyslogMessage`; `heartbeat` → `Heartbeat` |
| `GET /device/{udid}/ostrace`      | `OsTraceEvents`     | `ostrace` → `OsTraceEntry`; `heartbeat` → `Heartbeat` |
| `GET /device/{udid}/listen`       | `ListenEvents`      | `attachdetach` → `AttachDetachEvent`; `heartbeat` → `Heartbeat` |
| `GET /device/{udid}/sysmontap`    | `SysmontapEvents`   | `sample` → `CpuUsageSample`; `heartbeat` → `Heartbeat` |
| `GET /device/{udid}/jobs/{id}/logs`| `JobLogEvents`     | `log` → `JobLogLine`; `heartbeat` → `Heartbeat` |

`/ostrace` also accepts optional AND-combined query filters:
`pid`, `level`, `subsystem`, `match`, `exclude`.

`/sysmontap` streams CPU-usage samples; `/jobs/{id}/logs` replays the job's
buffered log history first, then streams live lines until the job ends.

### Event payload shapes (JSON)

```jsonc
// AppStateNotification (event: appstate)
{ "bundleId": "com.apple.Preferences", "state": "foreground", "timestamp": 1723200000000 }

// SyslogMessage (event: syslog)
{ "message": "…raw syslog line…", "timestamp": 1723200000000 }

// OsTraceEntry (event: ostrace)
{ "pid": 123, "processName": "SpringBoard", "level": "info",
  "subsystem": "com.apple.network", "category": "boringssl",
  "message": "…", "timestamp": 1723200000000 }

// AttachDetachEvent (event: attachdetach)
{ "event": "attached", "deviceID": 5, "udid": "00008110-…",
  "properties": { "serialNumber": "00008110-…", "connectionType": "USB", … } }

// CpuUsageSample (event: sample) — open map; sampler keys vary by OS
{ "CPU_TotalLoad": 42.5, "SystemLoad": 12.0, "UserLoad": 30.5 }

// JobLogLine (event: log)
{ "line": "…one line of job output…" }

// Heartbeat (event: heartbeat)
{}
```

`state` (AppStateNotification) is one of `foreground`, `background`, `suspended`,
`terminated`, `unknown`. `AttachDetachEvent.event` is `attached`, `detached`, or
`paired` (`properties` present on `attached`). `OsTraceEntry.level` is one of
`default`, `info`, `debug`, `error`, `fault`.

### How SDKs should expose SSE

- **TS:** async iterator, `for await (const ev of client.streamSyslog(udid))`.
- **Python:** async generator + context manager.
- **Java / C#:** hand-written SSE reader exposing an iterable/observable of typed
  events.
- Each event is dispatched by its `event:` name to the matching typed payload.
  Unknown event names should be surfaced (not dropped) for forward-compat.

### OpenAPI representation note (important for generators)

On **OpenAPI 3.1** the SSE responses render with content-type `text/event-stream`
and a bare `schema: { type: string }` — 3.1 cannot express the per-event typed
body inline (the emitter drops `itemSchema` with a warning). To keep the typed
contract machine-readable we provide it two ways:

1. **`x-sse-events` vendor extension** on every SSE operation in the 3.1 doc:
   `{ schema: "<UnionName>", events: { "<event-name>": "<PayloadModelName>", … } }`.
   All payload/union models are still fully defined under `components/schemas`.
2. **`spec/openapi/openapi.3.2.0.yaml`** — the 3.2 variant carries the full typed
   `itemSchema` inline (a `oneOf` keyed by `event` const, with `data` as a JSON
   `contentSchema` `$ref` to the payload model).

Generators/facades should read the event map from `x-sse-events` (or the 3.2 file)
and hand-write the typed dispatch; do not rely on the bare 3.1 SSE response schema.

## Binary & parity gaps (not in v0.1)

- `GET /screenshot` returns `image/png` bytes (the only binary endpoint in v0.1).
- `PUT /device/{udid}/image` accepts a raw image body (`application/octet-stream`,
  up to 2 GiB) or auto-resolves via `?auto=true&basedir=…`.
- **Not yet in v0.1 (parity gap):** mjpeg screen mirroring and pcap packet
  capture are binary streams that don't fit SSE/JSON framing. They will be
  chunked-HTTP binary streams with a hand-written async-iterator helper per SDK
  (one documented convention) when added. `/healthz` and `/readyz` are dead code
  in go-ios and are intentionally excluded.
