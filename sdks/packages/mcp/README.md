# @go-ios/mcp

A **Model Context Protocol (MCP) server for [go-ios](https://github.com/danielpaulus/go-ios)**.
It lets autonomous agents (Claude Desktop, IDE agents, custom MCP clients) **control iOS
devices** by proxying the go-ios REST daemon behind a small, high-signal set of tools.

It is **not** a naive 1:1 OpenAPI→tool mapping (which produces too many low-quality tools for
LLMs). Instead it exposes a hand-curated tool set with LLM-oriented descriptions, typed input
schemas (zod), structured output, and explicit error surfacing (device-not-found, auth failures).

Beyond observing a device, agents can now **drive** it: the `ui_*` tools tap, swipe, type, and
read the view hierarchy through WebDriverAgent (see [UI automation](#ui-automation-drive-the-device)).

Published on npm as [`@go-ios/mcp`](https://www.npmjs.com/package/@go-ios/mcp) (public).

## Requirements

- Node.js 20+
- A running **go-ios REST daemon** (`ios rest ...`) reachable over HTTP. This server talks to
  that daemon; it does not talk to devices directly.

## Tools

Every device-scoped tool takes a `udid` — get it from `list_devices` first. The set is
curated (not a 1:1 map of the daemon's ~125 endpoints); see
[Deliberately omitted](#deliberately-omitted-tools) for what is left out and why.

### Discovery & info

| Tool | What it does |
| --- | --- |
| `list_devices` | List connected/reachable devices (udid, connection type, address). Start here. |
| `device_info` | Lockdown + hardware/network/instruments info for a udid (open dictionary). |
| `device_health` | Quick reachability check + compact identity summary + battery snapshot; fails fast if offline. |

### Apps

| Tool | What it does |
| --- | --- |
| `list_apps` | Installed apps (bundle id, name, version). |
| `launch_app` | Launch an app by `bundleId`. |
| `terminate_app` | Kill a running app by `bundleId`. |
| `install_app` | Install from a local `.ipa`/`.app` path (uploaded as multipart, ≤200 MB). |
| `uninstall_app` | Uninstall an app by `bundleId`. |

### Screen & logs

| Tool | What it does |
| --- | --- |
| `screenshot` | Capture the screen; returns a PNG **image content block** (base64). |
| `stream_logs` | **Bounded** capture of syslog / os_log; returns a finite recent buffer (see below). |

### Diagnostics & health

| Tool | What it does |
| --- | --- |
| `device_battery` | Battery snapshot: charge level, charging/plugged-in, temperature. |
| `list_processes` | Running processes (pid, name, is-application); `apps_only` to filter. |
| `device_diagnostics` | Low-level IORegistry/diagnostic values (open map). |
| `device_diskspace` | Filesystem usage over AFC: total / free / used bytes, block size. |
| `device_ip` | Resolve the device's MAC / IPv4 / IPv6 by sniffing pcapd (takes a few seconds). |

### Crash logs

| Tool | What it does |
| --- | --- |
| `list_crash_reports` | List crash report file names (optional glob `pattern`). |
| `pull_crash_report` | Read one report's text by `name`, **bounded to 256 KiB** (`truncated` flag). |

### Files (read-only)

| Tool | What it does |
| --- | --- |
| `list_files` | List a device directory in the `app`/`app-group`/`crash`/`temp` domain. **Listing only.** |
| `list_device_files` | List a directory over AFC (media dir, or an app container via `bundleId`). **Listing only.** |
| `read_file` | Read a device file over AFC as text, **bounded to 512 KiB** (`truncated` flag). |

### Location

| Tool | What it does |
| --- | --- |
| `set_location_gpx` | Simulate live location by replaying a local `.gpx` track (uploaded as multipart). |

### Web debugging (WebInspector)

| Tool | What it does |
| --- | --- |
| `list_webinspector_pages` | List inspectable pages (Safari tabs / WKWebViews). Requires Web Inspector enabled. |
| `webinspector_eval` | Evaluate JavaScript in an inspectable page and return the result. Powerful for web debugging. |

### Performance

| Tool | What it does |
| --- | --- |
| `sample_performance` | **Bounded** CPU-usage sampling (sysmontap); finite buffer (see below). |

### Jobs (drive & observe long-running operations)

| Tool | What it does |
| --- | --- |
| `run_wda` | Start the WebDriverAgent runner as a background **job**; returns the job id. |
| `list_jobs` | List a device's jobs (test runs, WDA runners, forwards) with status. |
| `get_job` | Get one job's status/result by `id` (poll for completion). |
| `stop_job` | Stop a running job, or purge a finished one, by `id`. |
| `tail_job_logs` | **Bounded** capture of a job's log output; finite buffer (see below). |

### Pasteboard

| Tool | What it does |
| --- | --- |
| `get_pasteboard` | Read the device clipboard text (`{ present, text }`). |
| `set_pasteboard` | Set the device clipboard text (useful for injecting text into an app). |

### WebDriverAgent session lifecycle

| Tool | What it does |
| --- | --- |
| `create_wda_session` | Start a WebDriverAgent (XCUITest) session (prereq for UI automation). |
| `read_wda_session` | Fetch a running WDA session by `sessionId`. |
| `delete_wda_session` | Stop a WDA session. |

### UI automation (drive the device)

These tools let an agent **act on** the device, not just observe it. They forward to a UI
backend — **WebDriverAgent** (default) or **DeviceKit** — that must already be **running and
port-forwarded**.

| Tool | What it does |
| --- | --- |
| `ui_tap` | Tap at absolute screen coordinates `(x, y)`. |
| `ui_swipe` | Drag from `(x1, y1)` to `(x2, y2)` (scroll / swipe); optional `duration`. |
| `ui_type` | Type text into the focused field (tap it first). |
| `ui_press_button` | Press a hardware button by name (`home`; devicekit also volume). |
| `ui_source` | Return the view hierarchy (XML for WDA) — find element frames to tap. **Bounded to 512 KiB.** |
| `ui_screenshot` | Single screen frame **via the UI backend** (distinct from `screenshot`); PNG image block. |
| `ui_app_launch` | Launch an app by `bundleId` through the UI backend (attaches for automation). |
| `ui_app_terminate` | Terminate an app by `bundleId` through the UI backend. |

**Prerequisite — a forwarded WDA backend.** The UI tools need WebDriverAgent up and reachable:

1. Start the runner: **`run_wda`** (or `create_wda_session`).
2. Forward WDA's device port to a local port (e.g. `8100`) via the go-ios CLI/daemon.
3. Call the `ui_*` tools with **`wdaUrl`** set to that forwarded URL (e.g. `http://localhost:8100`);
   optionally set `backend` (`wda` | `devicekit`) and `timeout` (seconds).

A typical loop: `ui_source` (or `ui_screenshot`) to see the screen → `ui_tap`/`ui_type`/`ui_swipe`
to act → screenshot again to verify. If a UI tool returns **502**, the backend isn't reachable —
recheck `run_wda`, the port forward, and `wdaUrl`. **501** means the chosen backend doesn't support
that action. There is deliberately **no** UI video-stream tool — `ui_screenshot` captures a single
frame instead (large binary streams don't fit a tool response).

### Device management (disruptive)

| Tool | What it does |
| --- | --- |
| `reboot_device` | **DISRUPTIVE** — reboot the device (goes offline ~30–60s; kills apps/sessions/jobs). |
| `shutdown_device` | **DISRUPTIVE** — power off the device (needs physical interaction to boot again). |

### Deliberately omitted tools

Some daemon endpoints are intentionally **not** exposed as tools:

- **`erase`** (`POST /device/{udid}/erase`) — factory-wipes the device. Too destructive and
  irreversible to hand an autonomous agent; omitted on purpose.
- **`sign/*`** (`sign/app`, `sign/certificate`, `sign/provision`) and **`prepare`** — host-local,
  secret-handling (certificates, provisioning) supervision flows. Not agent-appropriate; left to
  the `ios` CLI / a human operator.
- **Raw binary streams** — `pcap` (packet capture), `screenshot/stream` and `ui/stream` (video)
  produce open-ended binary that doesn't fit a tool response. For screen capture, use the
  single-frame `screenshot` / `ui_screenshot`; for logs/CPU, use the **bounded** `stream_logs` /
  `sample_performance` / `tail_job_logs` capture pattern.
- **Raw file writes** (`files/push`, `fsync/push`, `fsync/rm`, `fsync/mkdir`, `crashes` delete) —
  the read-only `list_files` / `list_device_files` / `read_file` / `pull_crash_report` cover the
  safe, bounded read paths an agent needs. Writes/deletes to on-device paths are omitted to avoid
  corrupting app state.
- **Supervised/MDM & system-config endpoints** (profiles, wifi, http-proxy, language, time
  format, wallpaper, dev-mode, conditions, pairing, tunnels, image mount, MDM
  clear-passcode, …) — low agent value and easy to leave a device in a bad state; left to the
  `ios` CLI.

### Bounded captures are not infinite streams

Several go-ios endpoints are Server-Sent Event streams that run forever (`/syslog`,
`/ostrace`, `/sysmontap`, `/jobs/{id}/logs`). An agent tool call is request/response, so the
tools that consume them — **`stream_logs`**, **`sample_performance`**, and **`tail_job_logs`** —
collect events for a **bounded window** and then return them. They share one bounded-capture
implementation:

- The SSE stream is read until the **first** limit is hit: a `duration_seconds` cap or a
  line/sample count cap. Per-tool caps:
  - `stream_logs`: `duration_seconds` (default 5, **max 30**), `max_lines` (default 100, **max 1000**).
  - `sample_performance`: `duration_seconds` (default 5, **max 30**), `max_samples` (default 30, **max 300**).
  - `tail_job_logs`: `duration_seconds` (default 5, **max 30**), `max_lines` (default 200, **max 2000**).
- The underlying HTTP request is then aborted, so the call always returns promptly even if the
  device keeps emitting.
- `heartbeat` frames are dropped from the result but counted (`heartbeats`) so a live-but-idle
  stream is distinguishable from a dead one.
- Each result reports `stoppedBy` (`duration` | `maxLines` | `streamEnd`), `returned`,
  `totalMatched`, and the captured `lines`/`samples`.

`stream_logs` additionally supports `source: "ostrace"` with AND-combined filters (`pid`,
`level`, `subsystem`, `match`, `exclude`, plus a client-side `process` filter); `source:
"syslog"` returns raw lines.

## Configuration (environment)

| Variable | Default | Meaning |
| --- | --- | --- |
| `GO_IOS_BASE_URL` | `http://localhost:8080` | go-ios daemon base URL. |
| `GO_IOS_API_KEY` | *(none)* | Bearer token; sent as `Authorization: Bearer …` when set. |
| `GO_IOS_MCP_TRANSPORT` | `stdio` | `stdio` or `http`. |
| `GO_IOS_MCP_HTTP_PORT` | `3000` | Port for the HTTP transport. |
| `GO_IOS_MCP_HTTP_HOST` | `127.0.0.1` | Host/interface for the HTTP transport. |

CLI flags override env: `--stdio` / `--http`, `--port <n>`, `--host <h>`, `--base-url <url>`.

## Running

### stdio (default — local agent clients)

```bash
GO_IOS_BASE_URL=http://localhost:8080 \
GO_IOS_API_KEY=your-token \
npx @go-ios/mcp
```

### Streamable HTTP (remote clients)

```bash
GO_IOS_BASE_URL=http://localhost:8080 \
GO_IOS_API_KEY=your-token \
npx @go-ios/mcp --http --port 3000
```

The MCP endpoint is served at `POST /mcp` (Streamable HTTP; SSE is the response-streaming mode
within it, per the current MCP spec).

## MCP client config

### Claude Desktop (`claude_desktop_config.json`)

```json
{
  "mcpServers": {
    "go-ios": {
      "command": "npx",
      "args": ["-y", "@go-ios/mcp"],
      "env": {
        "GO_IOS_BASE_URL": "http://localhost:8080",
        "GO_IOS_API_KEY": "your-token"
      }
    }
  }
}
```

### Generic MCP client (HTTP)

Point the client at `http://127.0.0.1:3000/mcp` after starting the server with `--http`.

## Development

```bash
npm install
npm run build      # tsup -> dist/ (ESM, with shebang + bin entry)
npm test           # vitest
npm run typecheck  # tsc --noEmit
```

## Publishing

This package is **public and publishable** on npm as `@go-ios/mcp` (`publishConfig.access:
"public"`, provenance-friendly). It ships as part of the go-ios SDK release
(`.github/workflows/release-sdks.yml`) via npm OIDC trusted publishing — no auth token.
