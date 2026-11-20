/**
 * Example 01 — List devices.
 *
 * The most basic thing you can do: construct a client and ask the daemon which
 * devices it can see. This talks to `GET /list` under the hood via
 * `client.devices.list()` and does NOT require a device to be attached (an empty
 * list is a perfectly valid, successful response).
 *
 * Run it on its own:
 *   GO_IOS_API_KEY=... npx tsx examples/01-list-devices.ts
 */
import { isMain, makeClient, runAsScript, type IosClient } from "./_shared";

/**
 * The suite runner imports and calls this `run(client)` so all examples share
 * one client and the results can be aggregated. Each example follows the same
 * shape: an exported `run(client)` plus a direct-execution guard at the bottom.
 */
export async function run(client: IosClient): Promise<void> {
  // `list()` returns a { deviceList } envelope. Each entry carries the device's
  // udid at `properties.serialNumber` — that's the id every device-scoped route
  // keys on.
  const { deviceList } = await client.devices.list();

  console.log(`Found ${deviceList.length} device(s):`);
  for (const entry of deviceList) {
    const udid = entry.properties.serialNumber;
    // `connectionType` (e.g. "USB", "Network") and `address` (for tunneled
    // devices) are optional metadata the daemon includes when known.
    const via = entry.properties.connectionType ?? "unknown";
    const net = entry.address ? ` @ ${entry.address}` : "";
    console.log(`  - ${udid} (via ${via})${net}`);
  }

  if (deviceList.length === 0) {
    // Not an error: the daemon is up and answered, there's just nothing plugged
    // in. The suite runner treats this example as a PASS regardless.
    console.log("(no devices attached — plug one in to see it here)");
  }
}

// Direct execution: `tsx examples/01-list-devices.ts`.
if (isMain(import.meta.url)) {
  await runAsScript("01-list-devices", () => run(makeClient()));
}
