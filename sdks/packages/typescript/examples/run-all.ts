/**
 * run-all.ts — the examples suite runner AND pre-release smoke test.
 *
 * Runs examples 01–05 in sequence against the configured daemon, sharing one
 * `IosClient`. Example 06 (UI automation) runs only when RUN_UI=1, because it
 * needs a forwarded WebDriverAgent session.
 *
 * Exit code contract (so `npm run examples` is a clean green/red check):
 *   - Every example that ran either PASSED or SKIPPED  -> exit 0 (green)
 *   - Any example threw a non-SkipError                -> exit 1 (red)
 *
 * SKIP vs FAIL:
 *   - A device-dependent example that legitimately can't run (no device
 *     attached, UI backend not forwarded) throws a `SkipError` and is reported
 *     as SKIP — it does NOT fail the suite. This keeps the smoke test honest on
 *     a daemon with no device: the reachable surface is still verified.
 *   - Any other error (auth failure, daemon down, unexpected 5xx, a bug in an
 *     example) is a FAIL and flips the exit code to 1.
 *
 * Missing GO_IOS_API_KEY is handled up front by `makeClient()` (prints help and
 * exits non-zero) before any example runs.
 *
 * Usage:
 *   GO_IOS_API_KEY=... npm run examples
 *   RUN_UI=1 GO_IOS_API_KEY=... npm run examples   # also run 06
 */
import { IosClient } from "../src/index";
import { SkipError, baseUrl, makeClient } from "./_shared";

import { run as listDevices } from "./01-list-devices";
import { run as deviceInfo } from "./02-device-info";
import { run as listApps } from "./03-list-apps";
import { run as screenshot } from "./04-screenshot";
import { run as streamSyslog } from "./05-stream-syslog";
import { run as uiAutomation } from "./06-ui-automation";

interface Example {
  readonly name: string;
  readonly run: (client: IosClient) => Promise<void>;
}

/** The ordered core suite (always runs). */
const CORE: readonly Example[] = [
  { name: "01-list-devices", run: listDevices },
  { name: "02-device-info", run: deviceInfo },
  { name: "03-list-apps", run: listApps },
  { name: "04-screenshot", run: screenshot },
  { name: "05-stream-syslog", run: streamSyslog },
];

/** Opt-in advanced example, only when RUN_UI=1. */
const UI: Example = { name: "06-ui-automation", run: uiAutomation };

type Outcome = "PASS" | "SKIP" | "FAIL";

async function runOne(client: IosClient, ex: Example): Promise<Outcome> {
  console.log(`\n=== ${ex.name} ===`);
  try {
    await ex.run(client);
    console.log(`--- ${ex.name}: PASS`);
    return "PASS";
  } catch (err) {
    if (err instanceof SkipError) {
      console.log(`--- ${ex.name}: SKIP (${err.message})`);
      return "SKIP";
    }
    console.error(
      `--- ${ex.name}: FAIL — ${err instanceof Error ? err.stack ?? err.message : String(err)}`,
    );
    return "FAIL";
  }
}

async function main(): Promise<void> {
  // `makeClient()` enforces GO_IOS_API_KEY (exits non-zero if missing).
  const client = makeClient();
  console.log(`go-ios examples suite → ${baseUrl() ?? "(auto-discovered local daemon)"}`);

  const suite: Example[] = [...CORE];
  if (process.env.RUN_UI === "1") {
    suite.push(UI);
  } else {
    console.log("(RUN_UI!=1 — skipping 06-ui-automation; set RUN_UI=1 to include it)");
  }

  const results: { name: string; outcome: Outcome }[] = [];
  for (const ex of suite) {
    const outcome = await runOne(client, ex);
    results.push({ name: ex.name, outcome });
  }

  // Summary.
  const pass = results.filter((r) => r.outcome === "PASS").length;
  const skip = results.filter((r) => r.outcome === "SKIP").length;
  const fail = results.filter((r) => r.outcome === "FAIL").length;

  console.log("\n=== summary ===");
  for (const r of results) console.log(`  ${r.outcome.padEnd(4)} ${r.name}`);
  console.log(`  ${pass} passed, ${skip} skipped, ${fail} failed`);

  // Red only if something genuinely failed; SKIPs are fine.
  process.exit(fail > 0 ? 1 : 0);
}

await main();
