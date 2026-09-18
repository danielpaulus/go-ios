/**
 * run-all — the go-ios MCP smoke test (npm run examples).
 *
 * 1. list-tools (ALWAYS): spawns the built server over stdio, lists its tools,
 *    and asserts the exact curated set is present. Needs no device/daemon. If
 *    the set is wrong or the server won't start, this exits non-zero.
 * 2. call-tool (CONDITIONAL): only runs when GO_IOS_API_KEY is set AND the
 *    daemon is reachable; otherwise it SKIPs. Exercises list_devices for real.
 *
 * The result is a mostly-device-free MCP smoke test suitable for CI / a
 * pre-release gate.
 */
import { spawn } from "node:child_process";
import { fileURLToPath } from "node:url";
import { dirname, resolve } from "node:path";
import { listTools } from "./list-tools.js";
import { CURATED_TOOL_NAMES } from "../src/tools.js";

const here = dirname(fileURLToPath(import.meta.url));

/** Spawn `tsx <script>` and resolve with its exit code, streaming its output. */
function runScript(script: string, env: NodeJS.ProcessEnv): Promise<number> {
  return new Promise((resolveCode) => {
    const child = spawn(
      process.execPath,
      [resolve(here, "..", "node_modules", "tsx", "dist", "cli.mjs"), resolve(here, script)],
      { stdio: "inherit", env },
    );
    child.on("exit", (code) => resolveCode(code ?? 1));
    child.on("error", (err) => {
      console.error(`failed to run ${script}: ${err.message}`);
      resolveCode(1);
    });
  });
}

async function runListTools(): Promise<boolean> {
  console.log("── list-tools (always) ─────────────────────────────────────────");
  let tools;
  try {
    tools = await listTools();
  } catch (err) {
    console.error(`FAIL: server did not start / list tools: ${err instanceof Error ? err.message : String(err)}`);
    return false;
  }

  const got = new Set(tools.map((t) => t.name));
  const expected = new Set<string>(CURATED_TOOL_NAMES);

  const missing = [...expected].filter((n) => !got.has(n));
  const extra = [...got].filter((n) => !expected.has(n));

  tools.forEach((t, i) => {
    const firstSentence = t.description.split(/(?<=\.)\s/)[0] ?? t.description;
    console.log(`  ${String(i + 1).padStart(2, " ")}. ${t.name} — ${firstSentence}`);
  });
  console.log(`\n  listed ${tools.length} tools (expected ${expected.size}).`);

  if (missing.length || extra.length) {
    if (missing.length) console.error(`  FAIL: missing tools: ${missing.join(", ")}`);
    if (extra.length) console.error(`  FAIL: unexpected tools: ${extra.join(", ")}`);
    return false;
  }
  // Every tool must carry an LLM-oriented description.
  const undocumented = tools.filter((t) => !t.description.trim()).map((t) => t.name);
  if (undocumented.length) {
    console.error(`  FAIL: tools without a description: ${undocumented.join(", ")}`);
    return false;
  }
  console.log("  PASS: curated tool set matches exactly and all tools are documented.");
  return true;
}

async function runCallTool(): Promise<boolean> {
  console.log("\n── call-tool (only with GO_IOS_API_KEY + reachable daemon) ─────");
  if (!process.env.GO_IOS_API_KEY) {
    console.log("  SKIP call-tool: GO_IOS_API_KEY is not set (no daemon credentials).");
    return true;
  }
  // call-tool itself does the reachability check and SKIPs (exit 0) if down.
  const code = await runScript("call-tool.ts", process.env);
  if (code !== 0) {
    console.error(`  FAIL: call-tool exited ${code}.`);
    return false;
  }
  return true;
}

async function main(): Promise<void> {
  const listOk = await runListTools();
  const callOk = await runCallTool();

  console.log("\n────────────────────────────────────────────────────────────────");
  if (listOk && callOk) {
    console.log("examples: OK");
    process.exit(0);
  }
  console.error("examples: FAILED");
  process.exit(1);
}

main().catch((err) => {
  console.error(`run-all failed: ${err instanceof Error ? err.stack : String(err)}`);
  process.exit(1);
});
