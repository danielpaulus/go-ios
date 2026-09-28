# @go-ios/mcp examples

These examples double as **docs** and as a **pre-release smoke test** for the
go-ios MCP server. They live next to the package so `npm run examples` verifies
the built server before it ships.

The MCP server does **not** talk to devices directly — it proxies a running
**go-ios REST daemon** (started with `ios rest ...`). Listing tools needs
neither a device nor a daemon; calling a tool does.

## Contents

| Path | What it is |
| --- | --- |
| [`client-config/`](client-config/) | Ready-to-paste MCP client configs (Claude Desktop + a generic client), heavily commented. |
| [`list-tools.ts`](list-tools.ts) | Spawns the server over **stdio**, initializes an MCP client, calls `tools/list`, and prints every tool. **No device/daemon needed** — pure introspection. |
| [`call-tool.ts`](call-tool.ts) | Spawns the server and calls `list_devices` for real. **Needs a running daemon**; SKIPs cleanly if it can't reach one. |
| [`run-all.ts`](run-all.ts) | The `npm run examples` runner: always runs `list-tools` (asserts the exact curated tool set), runs `call-tool` only when `GO_IOS_API_KEY` is set + daemon reachable. |

## Running the MCP server

The examples spawn the server for you, but you can also run it directly.

### stdio (default — local agent clients)

```bash
GO_IOS_BASE_URL=http://localhost:8080 \
GO_IOS_API_KEY=your-token \
npx -y @go-ios/mcp
```

### Streamable HTTP (remote clients)

```bash
GO_IOS_BASE_URL=http://localhost:8080 \
GO_IOS_API_KEY=your-token \
npx -y @go-ios/mcp --http --port 3000
```

The MCP endpoint is then served at `POST http://127.0.0.1:3000/mcp`.

## Client configs

See [`client-config/`](client-config/):

- **`claude-desktop.json`** — paste-ready. Merge the `mcpServers.go-ios` entry
  into your real `claude_desktop_config.json` (strict JSON, no comments), set
  `GO_IOS_BASE_URL` / `GO_IOS_API_KEY`, then fully restart Claude Desktop.
  Config file locations:
  - macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
  - Windows: `%APPDATA%\Claude\claude_desktop_config.json`
- **`claude-desktop.jsonc`** — the same config, annotated with comments
  explaining every field (docs only; don't paste comments into the real file).
- **`generic-mcp-client.jsonc`** — annotated stdio **and** Streamable-HTTP blocks
  for any other MCP client; adapt the field names to your client's schema.

## The smoke test: `npm run examples`

From `sdks/packages/mcp/`:

```bash
npm install
npm run examples   # builds, then runs the examples runner
```

What it does:

1. **`list-tools` (always).** Spawns the freshly-built server over stdio,
   initializes an MCP client, and lists the tools. It **asserts the exact
   curated tool set** (currently 44 tools) is present and that every tool has a
   description. If the server won't start or the set is wrong, the runner exits
   non-zero — making this a genuine, device-free pre-release check.
2. **`call-tool` (conditional).** Only runs when `GO_IOS_API_KEY` is set **and**
   the daemon at `GO_IOS_BASE_URL` (default `http://localhost:8080`) is
   reachable. It calls `list_devices` end to end and prints the devices.
   Otherwise it **SKIPs** (still exit 0).

Run individual examples directly (after `npm run build`):

```bash
npx tsx examples/list-tools.ts
GO_IOS_API_KEY=your-token npx tsx examples/call-tool.ts
```

Typecheck the examples:

```bash
npx tsc --noEmit -p examples/tsconfig.json
```
