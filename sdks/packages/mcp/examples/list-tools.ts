/**
 * list-tools — spawn the go-ios MCP server over stdio, initialize an MCP
 * client against it, call `tools/list`, and print every tool's name and
 * description.
 *
 * This is the always-runnable pre-release smoke test: it introspects the server
 * and needs NO device and NO running go-ios daemon (it never calls a tool, it
 * only lists them). If it prints the curated tool set and exits 0, the server
 * builds, starts, and exposes its tools correctly.
 *
 * Run it with:  npm run build  &&  npx tsx examples/list-tools.ts
 * (or via the runner:  npm run examples)
 */
import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StdioClientTransport } from "@modelcontextprotocol/sdk/client/stdio.js";
import { fileURLToPath } from "node:url";
import { dirname, resolve } from "node:path";
import { existsSync } from "node:fs";

const here = dirname(fileURLToPath(import.meta.url));
/** The built server entry point (produced by `npm run build`). */
export const SERVER_ENTRY = resolve(here, "..", "dist", "index.js");

export interface ListedTool {
  name: string;
  description: string;
}

/**
 * Spawn the built MCP server over stdio, initialize, and return its tool list.
 * The caller owns nothing — the child process is started and torn down here.
 */
export async function listTools(): Promise<ListedTool[]> {
  if (!existsSync(SERVER_ENTRY)) {
    throw new Error(
      `Server bundle not found at ${SERVER_ENTRY}. Run \`npm run build\` first.`,
    );
  }

  // Spawn the server exactly as an MCP client would: `node dist/index.js` over
  // stdio. No GO_IOS_* env is needed — listing tools never touches the daemon.
  const transport = new StdioClientTransport({
    command: process.execPath, // the current node binary
    args: [SERVER_ENTRY, "--stdio"],
    // Silence the server's startup line on our stderr for clean output.
    stderr: "ignore",
  });

  const client = new Client({ name: "go-ios-mcp-examples", version: "0.1.0" });
  await client.connect(transport); // performs the MCP initialize handshake
  try {
    const { tools } = await client.listTools();
    return tools
      .map((t) => ({ name: t.name, description: t.description ?? "" }))
      .sort((a, b) => a.name.localeCompare(b.name));
  } finally {
    await client.close();
  }
}

/** Print the tool list as a readable, numbered catalog. */
function printTools(tools: ListedTool[]): void {
  console.log(`go-ios MCP server exposes ${tools.length} tools:\n`);
  tools.forEach((t, i) => {
    const n = String(i + 1).padStart(2, " ");
    // Descriptions are long (LLM-oriented); show the first sentence for scanning.
    const firstSentence = t.description.split(/(?<=\.)\s/)[0] ?? t.description;
    console.log(`${n}. ${t.name}`);
    console.log(`    ${firstSentence}`);
  });
}

// Run directly (not when imported by the runner).
const invokedDirectly =
  process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url);

if (invokedDirectly) {
  listTools()
    .then((tools) => {
      printTools(tools);
      process.exit(0);
    })
    .catch((err) => {
      console.error(`list-tools failed: ${err instanceof Error ? err.message : String(err)}`);
      process.exit(1);
    });
}
