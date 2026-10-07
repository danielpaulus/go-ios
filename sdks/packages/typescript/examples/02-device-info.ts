/**
 * Example 02 — Device info.
 *
 * Fetches lockdown / instruments info for the target device via
 * `client.device(udid).info()` (which maps to `GET /device/{udid}/info`).
 *
 * `info()` returns an OPEN dictionary (`DeviceInfo` is `{ [key: string]: unknown }`)
 * because the daemon surfaces whatever lockdown + `instruments:*` keys the device
 * reports. We pull out a few well-known keys (name / iOS version / model) and
 * print them, defensively, since any given key may be absent.
 *
 * Requires a device: GO_IOS_UDID or the first attached device. With no device
 * this SKIPs (see resolveDevice / SkipError).
 *
 * Run it on its own:
 *   GO_IOS_API_KEY=... npx tsx examples/02-device-info.ts
 */
import {
  isMain,
  makeClient,
  resolveDevice,
  runAsScript,
  type IosClient,
} from "./_shared";

/** Read a string-ish value from the open info dict, or undefined. */
function str(info: Record<string, unknown>, key: string): string | undefined {
  const v = info[key];
  return typeof v === "string" ? v : v == null ? undefined : String(v);
}

export async function run(client: IosClient): Promise<void> {
  const { udid } = await resolveDevice(client);
  const device = client.device(udid);

  // `info()` resolves the open DeviceInfo dictionary.
  const info = await device.info();

  // Common lockdown keys. They're not guaranteed, so we fall back gracefully.
  const name = str(info, "DeviceName") ?? "(unknown)";
  const version = str(info, "ProductVersion") ?? "(unknown)";
  const build = str(info, "BuildVersion");
  const model = str(info, "ProductType") ?? "(unknown)";

  console.log(`Device ${udid}:`);
  console.log(`  name:    ${name}`);
  console.log(`  iOS:     ${version}${build ? ` (${build})` : ""}`);
  console.log(`  model:   ${model}`);

  // `deviceName()` is a dedicated typed helper for just the name — handy when
  // that's all you need.
  const shortName = await device.deviceName();
  console.log(`  deviceName(): ${shortName}`);
}

if (isMain(import.meta.url)) {
  await runAsScript("02-device-info", () => run(makeClient()));
}
