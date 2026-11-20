/**
 * call-tool — spawn the go-ios MCP server over stdio and actually CALL a tool
 * (`list_devices`) end to end, printing the result.
 *
 * Unlike list-tools (pure introspection), this exercises the full path through
 * the server to the go-ios REST daemon, so it needs a RUNNING daemon reachable
 * at GO_IOS_BASE_URL (default http://localhost:8080), started with
 * `ios rest ...`, plus GO_IOS_API_KEY if the daemon requires auth.
 *
 * If the daemon is not reachable, this exits 0 with a clear SKIP message rather
 * than failing — it's an optional, device/daemon-dependent example.
 *
 * Run it with:  npm run build  &&  GO_IOS_API_KEY=... npx tsx examples/call-tool.ts
 */
import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StdioClientTransport } from "@modelcontextprotocol/sdk/client/stdio.js";
import { existsSync } from "node:fs";
import { SERVER_ENTRY } from "./list-tools.js";

const BASE_URL = (process.env.GO_IOS_BASE_URL ?? "http://localhost:8080").replace(/\/+$/, "");
const API_KEY = process.env.GO_IOS_API_KEY;

/** Quick pre-flight: is the go-ios daemon answering at all? */
async function daemonReachable(): Promise<boolean> {
  try {
    const headers: Record<string, string> = {};
    if (API_KEY) headers.Authorization = `Bearer ${API_KEY}`;
    // /api/v1/list is the device-list endpoint; any HTTP answer means "up".
    const res = await fetch(`${BASE_URL}/api/v1/list`, {
      headers,
      signal: AbortSignal.timeout(2000),
    });
    return res.status < 500 || res.status === 401; // reachable even if auth-gated
  } catch {
    return false;
  }
}

async function main(): Promise<number> {
  if (!existsSync(SERVER_ENTRY)) {
    console.error(`Server bundle not found at ${SERVER_ENTRY}. Run \`npm run build\` first.`);
    return 1;
  }

  if (!(await daemonReachable())) {
    console.log(
      `SKIP call-tool: no go-ios daemon reachable at ${BASE_URL}.\n` +
        `  Start one with \`ios rest --address 0.0.0.0 --port 8080 ...\`, set\n` +
        `  GO_IOS_BASE_URL / GO_IOS_API_KEY, and re-run to exercise list_devices.`,
    );
    return 0;
  }

  // Spawn the server, forwarding the daemon connection env so the tool call works.
  const transport = new StdioClientTransport({
    command: process.execPath,
    args: [SERVER_ENTRY, "--stdio"],
    env: {
      ...(process.env as Record<string, string>),
      GO_IOS_BASE_URL: BASE_URL,
      ...(API_KEY ? { GO_IOS_API_KEY: API_KEY } : {}),
    },
    stderr: "ignore",
  });

  const client = new Client({ name: "go-ios-mcp-examples", version: "0.1.0" });
  await client.connect(transport);
  try {
    console.log(`Calling list_devices against ${BASE_URL} ...\n`);
    const res = await client.callTool({ name: "list_devices", arguments: {} });

    if (res.isError) {
      const text = (res.content as Array<{ text?: string }>)[0]?.text ?? "unknown error";
      console.error(`list_devices returned an error: ${text}`);
      return 1;
    }

    const sc = res.structuredContent as
      | { count: number; devices: Array<Record<string, unknown>> }
      | undefined;
    if (sc) {
      console.log(`Found ${sc.count} device(s):`);
      console.log(JSON.stringify(sc.devices, null, 2));
    } else {
      // Fall back to the text content block.
      console.log((res.content as Array<{ text?: string }>)[0]?.text ?? "(no content)");
    }
    return 0;
  } finally {
    await client.close();
  }
}

main()
  .then((code) => process.exit(code))
  .catch((err) => {
    console.error(`call-tool failed: ${err instanceof Error ? err.message : String(err)}`);
    process.exit(1);
  });
