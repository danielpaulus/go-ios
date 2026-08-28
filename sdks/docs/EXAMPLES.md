# go-ios SDK examples

Each SDK (and the MCP server) ships a set of runnable, heavily-commented
**examples** under `sdks/packages/<lang>/examples/`. They serve two purposes:

1. **Docs** — the shortest correct path from "I have the SDK installed" to
   "I called the go-ios REST daemon and did something useful". Every example is
   annotated so it reads as a tutorial.
2. **Pre-release smoke test** — each package has an example *runner* that
   executes the whole set in sequence against a live daemon and exits non-zero
   if anything breaks. This is what the release/verify pipeline drives to prove
   the published clients actually talk to a real device before shipping.

## Per-language index

| SDK / target       | Examples README                                                             | Runner command (from the package dir)          |
| ------------------ | --------------------------------------------------------------------------- | ---------------------------------------------- |
| TypeScript         | [`packages/typescript/examples/README.md`](../packages/typescript/examples/README.md) | `npm run examples`                             |
| Python             | [`packages/python/examples/README.md`](../packages/python/examples/README.md)         | `uv run python examples/run_all.py`            |
| Java               | [`packages/java/examples/README.md`](../packages/java/examples/README.md)             | `bash examples/run.sh`                         |
| C# / .NET          | [`packages/csharp/examples/README.md`](../packages/csharp/examples/README.md)         | `dotnet run --project examples/GoIos.Examples -- run-all` |
| MCP server         | [`packages/mcp/examples/README.md`](../packages/mcp/examples/README.md)               | `npm run examples`                             |

The **MCP `list-tools`** example is special: it is fully **device-free and
daemon-free** (it just spawns the built MCP server over stdio and asserts the
exact curated 44-tool set), so it runs in ordinary CI as a gate. Its
`call-tool` companion auto-skips when no daemon is reachable. See
[Pre-release verification](#pre-release-verification) below.

## Shared environment convention

All example runners read the same environment variables so a single set of
exports drives every language:

| Variable          | Required | Default                              | Meaning                                                         |
| ----------------- | -------- | ------------------------------------ | -------------------------------------------------------------- |
| `GO_IOS_BASE_URL` | no       | auto-discovered (`~/.go-ios/rest-api.json`) | Origin of the go-ios REST daemon. The SDK appends `/api/v1`. Unset → discover the local daemon. |
| `GO_IOS_API_KEY`  | yes\*    | —                                    | Bearer token. \*Not needed if the daemon runs `--disable-auth`. |
| `GO_IOS_UDID`     | no       | first attached device                | Which device to target.                                        |
| `RUN_UI`          | no       | unset (off)                          | Set to `1` to also run the mutating UI-automation example.     |

```bash
export GO_IOS_API_KEY=your-secret          # required (unless --disable-auth)
# GO_IOS_BASE_URL is optional; unset, the local daemon is auto-discovered.
# export GO_IOS_BASE_URL=http://localhost:8080   # only to pin a fixed/remote daemon
export GO_IOS_UDID=00008030-0000...        # optional; first device otherwise
# export RUN_UI=1                          # optional; include UI automation
```

Read-only examples SKIP (rather than fail) when a device isn't attached, so the
runners are safe to invoke even without hardware — but for a real smoke test you
want a device connected.

## Daemon discovery (no hardcoded port)

By default the go-ios REST daemon binds an **ephemeral loopback port**
(`--addr` defaults to `127.0.0.1:0`) and, after binding, writes a discovery file
at `~/.go-ios/rest-api.json` (`$GO_IOS_HOME/rest-api.json` when `GO_IOS_HOME` is
set) containing the real `baseUrl`. The file is removed on graceful shutdown.

Every SDK (TypeScript, Python, Java, C#) resolves its base URL in the same
order — the examples rely on this rather than hardcoding a port:

1. an **explicit** `baseUrl` / `BaseUrl` init option → used verbatim (for remote
   daemons; discovery is skipped);
2. the **`GO_IOS_BASE_URL`** environment variable;
3. the **discovery file** `~/.go-ios/rest-api.json` (its `baseUrl`);
4. otherwise a **clear error** telling you to start the daemon or pass a base URL.

So you normally just start the daemon and run the examples — no port to know:

```bash
ios api --api-key "$GO_IOS_API_KEY"        # ephemeral loopback port; writes ~/.go-ios/rest-api.json
```

To **pin** a fixed port (e.g. to reach a daemon on another host, or expose it),
start it with `--addr` and point the SDK at it:

```bash
ios api --api-key "$GO_IOS_API_KEY" --addr :8080
export GO_IOS_BASE_URL=http://localhost:8080
```

> **Resolved:** the earlier SDK-library-vs-daemon default mismatch (the SDK
> libraries defaulting `baseUrl` to `http://localhost:60105` from the OpenAPI
> spec's `servers` URL, while the daemon listened on `:8080`) no longer exists.
> The SDKs have **no hardcoded default** — they auto-discover the local daemon
> via `~/.go-ios/rest-api.json`. The OpenAPI `servers` URL is now a documentation
> placeholder only. (Daemon side: PR #825 / #821; SDK side: PR #819.)

## Run them all

With the environment exported and a daemon (plus device) reachable, from each
package directory:

```bash
# TypeScript
cd sdks/packages/typescript && npm ci && npm run examples

# Python
cd sdks/packages/python && uv sync --all-extras && uv run python examples/run_all.py

# Java (JDK 17+, no Maven needed)
cd sdks/packages/java && bash examples/run.sh

# C# / .NET 8
cd sdks/packages/csharp && dotnet run --project examples/GoIos.Examples -- run-all

# MCP server (list-tools is device-free; call-tool needs a daemon)
cd sdks/packages/mcp && npm ci && npm run build && npm run examples
```

## Pre-release verification

Two layers of CI drive these examples:

1. **Device-free gate (always on).** `.github/workflows/sdks.yml` runs the MCP
   `list-tools` smoke check (`cd sdks/packages/mcp && npm ci && npm run build &&
   npm run examples`). It needs no device and no daemon, so a broken MCP server
   — a bad tool set, a server that won't start — fails ordinary CI on every
   `sdks/**` change. (`call-tool` auto-skips without daemon credentials.)

2. **Device-dependent verification (farm-gated, dispatch-only).**
   `.github/workflows/verify-sdk-examples.yml` runs *every* SDK's example runner
   against a real go-ios REST daemon + device on the self-hosted farm
   (office01 / ganjalf). It is **`workflow_dispatch`-only and inert today**: the
   full REST daemon is not yet on `main` (it lands with PRs **#817 / #821**), so
   the workflow documents — but does not yet execute — the daemon-start step. It
   activates once the daemon is deployable on the farm. See the header of that
   workflow file for the exact daemon-start command to fill in.
