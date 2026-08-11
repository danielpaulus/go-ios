/**
 * Shared helpers for the @go-ios/sdk examples.
 *
 * Every example is a standalone script that is configured entirely through
 * environment variables, so it can be run against any running go-ios REST
 * daemon without editing code:
 *
 *   GO_IOS_BASE_URL   Base URL of the daemon. Default: http://localhost:8080
 *   GO_IOS_API_KEY    Bearer API key. REQUIRED (the examples refuse to run
 *                     without one, so you never accidentally hit an unsecured
 *                     daemon). Start the daemon with the same key.
 *   GO_IOS_UDID       Target device udid. OPTIONAL — when unset, the examples
 *                     fall back to the first device returned by `list`.
 *
 * These helpers deliberately live in the examples folder (not the SDK) so the
 * examples read as real, copy-pasteable usage of the public `@go-ios/sdk` API.
 *
 * NOTE ON IMPORTS: the examples import the SDK from its source (`../src/index`)
 * so they typecheck and run straight from the repo with `tsx`, no build step
 * required. In your own project you would instead:
 *
 *   import { IosClient } from "@go-ios/sdk";
 */
import { IosClient, type DeviceEntry } from "../src/index";

// Re-export the client type so each example can import it from one place.
export { IosClient } from "../src/index";
export type { DeviceEntry } from "../src/index";

/** Default daemon URL, matching the task's contract. */
export const DEFAULT_BASE_URL = "http://localhost:8080";

/**
 * A small typed error the runner recognizes as "this step could not run in this
 * environment, but that is not a failure" (e.g. no device attached). Throwing it
 * from an example makes the runner print a SKIP instead of a hard failure.
 */
export class SkipError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "SkipError";
  }
}

/**
 * True when the given `import.meta.url` is the process entrypoint, i.e. the file
 * was run directly (`tsx examples/01-list-devices.ts`) rather than imported by
 * the suite runner. Used to guard each example's direct-run block.
 */
export function isMain(moduleUrl: string): boolean {
  return process.argv[1] !== undefined && moduleUrl === `file://${process.argv[1]}`;
}

/** Resolve the base URL from the environment (or the default). */
export function baseUrl(): string {
  return process.env.GO_IOS_BASE_URL?.trim() || DEFAULT_BASE_URL;
}

/**
 * Read the API key. If it is missing we print a helpful message and exit
 * non-zero: an example is useless (and a smoke test is misleading) if it silently
 * skips auth. This is the "on missing GO_IOS_API_KEY, exit non-zero" contract.
 */
export function requireApiKey(): string {
  const key = process.env.GO_IOS_API_KEY?.trim();
  if (!key) {
    console.error(
      [
        "ERROR: GO_IOS_API_KEY is not set.",
        "",
        "Start a go-ios REST daemon with an API key, e.g.:",
        "  GO_IOS_API_KEY=dev-secret ios --rest --api-key dev-secret",
        "",
        "then export the same key before running an example:",
        "  export GO_IOS_API_KEY=dev-secret",
      ].join("\n"),
    );
    process.exit(1);
  }
  return key;
}

/** Construct a configured {@link IosClient} from the environment. */
export function makeClient(): IosClient {
  return new IosClient({ baseUrl: baseUrl(), apiKey: requireApiKey() });
}

/**
 * Resolve the target device.
 *
 * Uses GO_IOS_UDID when set; otherwise lists devices and returns the first one.
 * When no device is attached it throws a {@link SkipError} so device-dependent
 * examples SKIP (rather than fail) on a daemon with no device.
 */
export async function resolveDevice(client: IosClient): Promise<{
  udid: string;
  entry?: DeviceEntry;
}> {
  const envUdid = process.env.GO_IOS_UDID?.trim();
  if (envUdid) return { udid: envUdid };

  const { deviceList } = await client.devices.list();
  const first = deviceList[0];
  if (!first) {
    throw new SkipError(
      "no device attached (device list is empty) and GO_IOS_UDID is not set",
    );
  }
  return { udid: first.properties.serialNumber, entry: first };
}

/**
 * Small runner used by each example's `main()` so a script run directly
 * (`tsx examples/01-list-devices.ts`) exits with the right code:
 *   - clean success        -> exit 0
 *   - SkipError             -> print SKIP, exit 0 (nothing was wrong)
 *   - any other error       -> print it, exit 1
 *
 * The suite runner (`run-all.ts`) does NOT use this — it imports each example's
 * `run(client)` directly so it can aggregate results — but running a single file
 * on its own still behaves sensibly.
 */
export async function runAsScript(
  name: string,
  fn: () => Promise<void>,
): Promise<void> {
  try {
    await fn();
    process.exit(0);
  } catch (err) {
    if (err instanceof SkipError) {
      console.log(`SKIP (${name}): ${err.message}`);
      process.exit(0);
    }
    console.error(`FAIL (${name}):`, err instanceof Error ? err.message : err);
    process.exit(1);
  }
}
