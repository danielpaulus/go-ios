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

| Variable          | Required | Default                 | Meaning                                                         |
| ----------------- | -------- | ----------------------- | -------------------------------------------------------------- |
| `GO_IOS_BASE_URL` | no       | `http://localhost:8080` | Origin of the go-ios REST daemon. The SDK appends `/api/v1`.   |
| `GO_IOS_API_KEY`  | yes\*    | —                       | Bearer token. \*Not needed if the daemon runs `--disable-auth`. |
| `GO_IOS_UDID`     | no       | first attached device   | Which device to target.                                        |
| `RUN_UI`          | no       | unset (off)             | Set to `1` to also run the mutating UI-automation example.     |

```bash
export GO_IOS_API_KEY=your-secret          # required (unless --disable-auth)
export GO_IOS_BASE_URL=http://localhost:8080   # optional; this is the default
export GO_IOS_UDID=00008030-0000...        # optional; first device otherwise
# export RUN_UI=1                          # optional; include UI automation
```

Read-only examples SKIP (rather than fail) when a device isn't attached, so the
runners are safe to invoke even without hardware — but for a real smoke test you
want a device connected.

## Canonical daemon port: `8080`

**The examples target `http://localhost:8080`**, which is the go-ios REST
daemon's real default bind address (`--addr` defaults to `:8080` in the daemon's
`restapi/api/server.go`). Start the daemon with:

```bash
ios api --api-key "$GO_IOS_API_KEY"        # serves on http://localhost:8080
```

> **SDK-library-vs-daemon default discrepancy (for Daniel to reconcile):**
> the *SDK client libraries* (TypeScript, Python, Java, C#) currently default
> their `baseUrl` / `BaseUrl` to **`http://localhost:60105`** — that value comes
> from the OpenAPI spec's `servers` URL, **not** from the daemon, which actually
> listens on **`:8080`**. The examples in this tree are deliberately consistent
> with the *daemon* (`:8080`) and always pass an explicit base URL, so they are
> correct regardless. Reconciling the library default (change the spec `servers`
> URL to `:8080`, or change the daemon default to `:60105`) is **out of scope**
> for the examples work and left for a follow-up so the two agree.

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
