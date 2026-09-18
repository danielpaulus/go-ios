/**
 * Example 06 — UI automation (ADVANCED, opt-in).
 *
 * Drives on-screen UI through `client.device(udid).ui` (tap + type). Unlike the
 * other examples, this one needs a live UI backend — by default WebDriverAgent
 * (WDA) reachable over a forwarded port. That is NOT set up automatically, so
 * this example is:
 *
 *   - OPT-IN in the suite runner: it only runs when RUN_UI=1.
 *   - Self-guarding: it probes `ui.status()` first and SKIPs gracefully (clear
 *     message, non-failure) if the UI backend is unreachable.
 *
 * ── Prerequisites (do these once, in separate shells, before RUN_UI=1) ────────
 *
 * 1. Start the WDA runner as a job (keeps running):
 *      curl -H "Authorization: Bearer $GO_IOS_API_KEY" \
 *           -X POST "$GO_IOS_BASE_URL/api/v1/device/$GO_IOS_UDID/jobs/runwda"
 *    (or `client.device(udid).jobs.runwda()` — see 06b note below.)
 *
 * 2. Forward the host port WDA listens on to the device, as another job:
 *      client.device(udid).jobs.forward({ hostPort: 8100, targetPort: 8100 })
 *
 * With WDA reachable at http://localhost:8100 (the SDK's default `wdaUrl`), the
 * `ui.*` methods work. If your forward uses a different host port, pass
 * `{ wdaUrl: "http://localhost:<port>" }` to each `ui.*` call.
 *
 * NB: This example itself does NOT start those jobs — it only *uses* the UI
 * surface, so the demo stays focused and safe. The commented `bootstrap()`
 * helper at the bottom shows how you'd start them from the SDK if you want a
 * fully self-contained run.
 *
 * Run it on its own (with WDA already forwarded):
 *   RUN_UI=1 GO_IOS_API_KEY=... npx tsx examples/06-ui-automation.ts
 */
import { IosApiError } from "../src/index";
import {
  SkipError,
  isMain,
  makeClient,
  resolveDevice,
  runAsScript,
  type IosClient,
} from "./_shared";

/** Default forwarded WDA URL the SDK targets; override via UI_WDA_URL if needed. */
const WDA_URL = process.env.UI_WDA_URL?.trim() || undefined;

export async function run(client: IosClient): Promise<void> {
  const { udid } = await resolveDevice(client);
  const ui = client.device(udid).ui;

  // Shared UI options: default backend (wda) + optional forwarded URL override.
  const opts = WDA_URL ? { wdaUrl: WDA_URL } : {};

  // 1) Probe the backend. If WDA isn't forwarded/reachable, this throws — we turn
  //    that into a clean SKIP rather than a hard failure.
  try {
    await ui.status(opts);
  } catch (err) {
    const why =
      err instanceof IosApiError
        ? `UI backend not reachable (status ${err.status}: ${err.message})`
        : `UI backend not reachable (${err instanceof Error ? err.message : String(err)})`;
    throw new SkipError(
      `${why}. Start WDA (jobs.runwda) and forward its port (jobs.forward) first — see the header of this file.`,
    );
  }

  // 2) Query the screen size so we tap somewhere sensible (center-ish).
  const size = await ui.size(opts);
  console.log(`UI backend is up. Window size response:`, size);

  // Pull width/height defensively from the open UiResponse `value`.
  const value = (size as { value?: Record<string, unknown> }).value ?? {};
  const width = numberOr(value["width"], 200);
  const height = numberOr(value["height"], 400);

  // 3) Tap near the center of the screen.
  const x = Math.round(width / 2);
  const y = Math.round(height / 2);
  console.log(`Tapping at (${x}, ${y})…`);
  await ui.tap(x, y, opts);

  // 4) Type some text into whatever field is focused.
  const text = "hello from @go-ios/sdk";
  console.log(`Typing: ${JSON.stringify(text)}`);
  await ui.type(text, opts);

  console.log("UI automation demo complete.");
}

/** Coerce an unknown to a finite number, else a fallback. */
function numberOr(v: unknown, fallback: number): number {
  return typeof v === "number" && Number.isFinite(v) ? v : fallback;
}

if (isMain(import.meta.url)) {
  await runAsScript("06-ui-automation", () => run(makeClient()));
}

/*
 * Optional: fully self-contained bootstrap of the UI backend from the SDK.
 * Uncomment and call before `run()` if you'd rather not set up WDA by hand.
 * (Left commented so the example's happy path stays focused on `ui.*` usage.)
 *
 * async function bootstrap(client: IosClient, udid: string): Promise<void> {
 *   // Start the WebDriverAgent runner (long-running job).
 *   await client.device(udid).jobs.runwda();
 *   // Forward host:8100 -> device:8100 so http://localhost:8100 reaches WDA.
 *   await client.device(udid).jobs.forward({ hostPort: 8100, targetPort: 8100 });
 *   // Give WDA a moment to come up before driving it.
 *   await new Promise((r) => setTimeout(r, 3_000));
 * }
 */
