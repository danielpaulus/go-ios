/**
 * Example 04 — Screenshot.
 *
 * Captures a PNG screenshot of the target device via
 * `client.device(udid).screenshot()` (`GET /device/{udid}/screenshot`), which
 * resolves to a `Blob` of image/png bytes, then writes it to `./screenshot.png`
 * and prints the byte size.
 *
 * Requires a device (SKIPs with none).
 *
 * Run it on its own:
 *   GO_IOS_API_KEY=... npx tsx examples/04-screenshot.ts
 */
import { writeFile } from "node:fs/promises";
import { resolve } from "node:path";

import {
  isMain,
  makeClient,
  resolveDevice,
  runAsScript,
  type IosClient,
} from "./_shared";

/** Where the PNG is written (cwd-relative, per the task's `./screenshot.png`). */
const OUTPUT = resolve(process.cwd(), "screenshot.png");

export async function run(client: IosClient): Promise<void> {
  const { udid } = await resolveDevice(client);

  // `screenshot()` returns a Blob (image/png). Convert to bytes to both measure
  // and persist it.
  const png = await client.device(udid).screenshot();
  const bytes = new Uint8Array(await png.arrayBuffer());

  await writeFile(OUTPUT, bytes);

  console.log(`Captured screenshot from ${udid}`);
  console.log(`  wrote ${bytes.byteLength} bytes -> ${OUTPUT}`);
}

if (isMain(import.meta.url)) {
  await runAsScript("04-screenshot", () => run(makeClient()));
}
