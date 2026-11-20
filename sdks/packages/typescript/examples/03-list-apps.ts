/**
 * Example 03 — List installed apps.
 *
 * Lists the applications installed on the target device via
 * `client.device(udid).apps.list()` (`GET /device/{udid}/apps`).
 *
 * Each entry is an `AppInfo` — an open Info.plist-derived map; the SDK surfaces
 * the common keys (`CFBundleIdentifier`, `CFBundleName`, version, etc). We print
 * a compact table and a total count.
 *
 * Requires a device (SKIPs with none).
 *
 * Run it on its own:
 *   GO_IOS_API_KEY=... npx tsx examples/03-list-apps.ts
 */
import {
  isMain,
  makeClient,
  resolveDevice,
  runAsScript,
  type IosClient,
} from "./_shared";

export async function run(client: IosClient): Promise<void> {
  const { udid } = await resolveDevice(client);
  const apps = await client.device(udid).apps.list();

  console.log(`Device ${udid} has ${apps.length} installed app(s):`);

  // Sort by bundle id for stable, readable output.
  const sorted = [...apps].sort((a, b) =>
    (a.CFBundleIdentifier ?? "").localeCompare(b.CFBundleIdentifier ?? ""),
  );

  // Print the first 25 so the output stays scannable on big devices.
  const shown = sorted.slice(0, 25);
  for (const app of shown) {
    const id = app.CFBundleIdentifier ?? "(no bundle id)";
    const label = app.CFBundleName ?? app.CFBundleExecutable ?? "";
    const version = app.CFBundleShortVersionString ?? "";
    console.log(`  - ${id}  ${label}${version ? ` v${version}` : ""}`);
  }
  if (sorted.length > shown.length) {
    console.log(`  … and ${sorted.length - shown.length} more`);
  }
}

if (isMain(import.meta.url)) {
  await runAsScript("03-list-apps", () => run(makeClient()));
}
