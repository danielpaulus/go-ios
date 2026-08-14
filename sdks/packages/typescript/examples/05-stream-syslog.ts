/**
 * Example 05 — Stream syslog (Server-Sent Events).
 *
 * Opens the live syslog SSE stream via `client.device(udid).syslog()` and
 * consumes it with `for await`. This is the SDK's typed-iterator streaming
 * surface: each yielded value is `{ event, data }`; we use the `isSseEvent`
 * type guard to narrow to a `syslog` event and get a typed `data.message`.
 *
 * We stop after ~20 events OR ~5 seconds, whichever comes first, by aborting an
 * `AbortController` whose `signal` we hand to `syslog()`. Aborting cleanly closes
 * the underlying HTTP connection — the `for await` loop ends and the process can
 * exit. We swallow the resulting abort error (it's expected), so a clean,
 * timer-driven stop is a SUCCESS, not a failure. This matters for the smoke test:
 * "05 must cleanly stop".
 *
 * Requires a device (SKIPs with none).
 *
 * Run it on its own:
 *   GO_IOS_API_KEY=... npx tsx examples/05-stream-syslog.ts
 */
import { isSseEvent } from "../src/index";
import {
  isMain,
  makeClient,
  resolveDevice,
  runAsScript,
  type IosClient,
} from "./_shared";

/** Stop conditions: whichever fires first. */
const MAX_EVENTS = 20;
const MAX_MILLIS = 5_000;

export async function run(client: IosClient): Promise<void> {
  const { udid } = await resolveDevice(client);

  const ac = new AbortController();
  // Safety timer: even on a quiet device (few log lines) we stop after 5s.
  const timer = setTimeout(() => ac.abort(), MAX_MILLIS);
  // Don't let the timer keep the event loop alive on its own.
  timer.unref?.();

  let count = 0;
  console.log(`Streaming syslog from ${udid} (up to ${MAX_EVENTS} events / ${MAX_MILLIS / 1000}s)…`);

  try {
    for await (const ev of client.device(udid).syslog({ signal: ac.signal })) {
      // Narrow the open event union to the typed `syslog` payload.
      if (isSseEvent(ev, "syslog")) {
        count += 1;
        const ts = ev.data.timestamp ? new Date(ev.data.timestamp).toISOString() : "";
        // Trim noisy lines so the demo output stays readable.
        const msg = ev.data.message.replace(/\s+$/, "").slice(0, 160);
        console.log(`  [${count}] ${ts} ${msg}`);
        if (count >= MAX_EVENTS) {
          ac.abort(); // reached our event budget — stop cleanly
          break;
        }
      }
    }
  } catch (err) {
    // An abort we triggered ourselves is the expected, clean stop — not an error.
    if (!ac.signal.aborted) throw err;
  } finally {
    clearTimeout(timer);
  }

  console.log(`Stopped after ${count} syslog event(s).`);
}

if (isMain(import.meta.url)) {
  await runAsScript("05-stream-syslog", () => run(makeClient()));
}
