# @go-ios/sdk

TypeScript SDK for the [go-ios](https://github.com/danielpaulus/go-ios) REST
server. Provides a typed, ergonomic client covering the **full daemon surface**
(80 operations) — device info & diagnostics, management (reboot/shutdown/erase/
devmode/lang), app management, on-device files & crash reports, media
(wallpaper/icon-layout/pasteboard), settings (AssistiveTouch/time format/wifi),
supervised MDM, HTTP proxy, configuration profiles, developer disk images,
WebDriverAgent sessions, condition inducers, async jobs (runtest/runwda/forward)
and device tunnels — plus six long-lived **Server-Sent Events** streams exposed
as typed async iterators.

- **ESM + CJS** dual build, ships `.d.ts` types, `sideEffects: false`, Node 20+.
- Low-level client generated from the canonical OpenAPI 3.1 spec with
  [`@hey-api/openapi-ts`](https://github.com/hey-api/openapi-ts); a thin
  hand-written facade gives the public API below (identical shape across the
  Python / Java / C# SDKs).

## Install

```sh
npm install @go-ios/sdk
```

## Quickstart

```ts
import { IosClient, isSseEvent } from "@go-ios/sdk";

const client = new IosClient({
  baseUrl: "http://localhost:60105",
  apiKey: process.env.GO_IOS_API_KEY, // optional if the server runs --disable-auth
});

// List devices
const { deviceList } = await client.devices.list();
const udid = deviceList[0].properties.serialNumber;

// Device-scoped operations
const device = client.device(udid);
const info = await device.info();
const png = await device.screenshot(); // Blob of image/png bytes

// Apps
await device.apps.launch("com.apple.Preferences");
const apps = await device.apps.list();

// WebDriverAgent
const session = await device.wda.createSession({
  bundleId: "com.facebook.WebDriverAgentRunner.xctrunner",
  testBundleId: "com.facebook.WebDriverAgentRunner",
  xcTestConfig: "WebDriverAgentRunner.xctest",
});
await device.wda.deleteSession(session.sessionId);
```

## Streaming (SSE)

The `syslog`, `notifications`, `ostrace`, `listen`, `sysmontap` (device-scoped)
and `jobs.logs(id)` endpoints return typed async iterables. Each yielded value is `{ event, data }`. Use the `isSseEvent`
type guard to narrow to a known event and get a typed `data` (a bare
`ev.event === "syslog"` check can't narrow the open union). Heartbeat frames are
consumed internally and never yielded; unknown event names are surfaced (not
dropped) as `{ event: string, data: unknown }` for forward compatibility.

```ts
// Stream syslog until you break out
for await (const ev of client.device(udid).syslog()) {
  if (isSseEvent(ev, "syslog")) {
    console.log(ev.data.timestamp, ev.data.message);
  }
}

// Filtered os_log trace with cancellation
const ac = new AbortController();
setTimeout(() => ac.abort(), 5000);
try {
  for await (const ev of client.device(udid).ostrace(
    { level: "error", subsystem: "com.apple.network" },
    { signal: ac.signal },
  )) {
    if (isSseEvent(ev, "ostrace")) console.log(ev.data.processName, ev.data.message);
  }
} catch (err) {
  if (!ac.signal.aborted) throw err;
}

// Device attach/detach
for await (const ev of client.device(udid).listen({ signal: ac.signal })) {
  if (isSseEvent(ev, "attachdetach")) console.log(ev.data.event, ev.data.udid);
}

// CPU-usage samples (sysmontap)
for await (const ev of client.device(udid).sysmontap({ signal: ac.signal })) {
  if (isSseEvent(ev, "sample")) console.log(ev.data.CPU_TotalLoad);
}

// Start a test run, then follow its logs
const job = await client.device(udid).jobs.runtest({ bundleId: "com.example.MyAppUITests" });
for await (const ev of client.device(udid).jobs.logs(job.id)) {
  if (isSseEvent(ev, "log")) process.stdout.write(ev.data.line);
}
```

You can also use the low-level SSE helpers directly against any
`text/event-stream` `Response.body`:

```ts
import { parseSseStream, SseFrameParser, isSseEvent } from "@go-ios/sdk";
```

## Authentication

The server uses bearer auth on `/api/v1`. Pass `apiKey` to the constructor and it
is sent as `Authorization: Bearer <apiKey>` on every request. When the server is
started with `--disable-auth` the key may be omitted, but supplying it is
harmless and recommended.

## Errors

Any non-2xx response throws an `IosApiError` carrying `.status` (HTTP status) and
`.body` (the parsed `GenericResponse` envelope when present):

```ts
import { IosApiError } from "@go-ios/sdk";

try {
  await client.device("unknown-udid").info();
} catch (err) {
  if (err instanceof IosApiError && err.status === 404) {
    console.error("device not found:", err.message);
  }
}
```

## Public API

- `new IosClient({ baseUrl, apiKey?, fetch? })`
- `client.devices.list()`
- `deviceUdid(entry)` — convenience accessor for `entry.properties.serialNumber`
- `client.device(udid)` →
  - info → `info()`, `deviceName()`, `date()`, `battery()`, `diagnostics()`,
    `mobileGestalt(keys)`, `processes()`, `lockdown()`, `screenshot()`
  - management → `reboot()`, `shutdown()`, `erase(confirm)`, `devmode()`,
    `setDevmode(action, enablePostRestart?)`, `lang()`, `setLang(language?, locale?)`,
    `memlimitoff(process)`, `activate()`, `pair(opts)`, `resetAccessibility()`
  - location → `resetLocation()`, `setLocation(latitude, longitude)`
  - conditions → `conditions()`, `enableCondition(profileTypeId, profileId)`, `disableCondition()`
  - developer disk images → `images()`, `mountedImages()`, `installImage(opts)`, `unmountImage()`
  - configuration profiles → `profiles()`, `addProfile(opts)`, `removeProfile(name)`
  - `apps` → `list()`, `launch(bundleId)`, `kill(bundleId)`, `install(ipa)`, `uninstall(bundleId)`
  - `wda` → `createSession(config)`, `readSession(id)`, `deleteSession(id)`
  - `files` → `ls(scope, path?)`, `pull(scope, remote)`, `push(scope, remote, data)`
  - `crashes` → `list(pattern?)`, `remove(pattern, cwd?)`
  - `media` → `getWallpaper()`, `setWallpaper(opts)`, `getIconLayout()`, `setIconLayout(layout)`,
    `getPasteboard()`, `setPasteboard(text)`
  - `settings` → `assistiveTouch()`, `setAssistiveTouch(enabled)`, `timeFormat()`,
    `setTimeFormat(uses24Hour)`, `setWifi(ssid, password?, encType?)`, `removeWifi(ssid)`
  - `mdm` → `securityInfo(identity)`, `fetchUnlockToken(identity)`,
    `clearPasscode(identity, token)`, `clearScreenTimePassword(identity)`
  - `proxy` → `setHttpProxy(opts)`, `removeHttpProxy()`
  - `jobs` → `runtest(cfg)`, `runwda(cfg?)`, `forward(cfg)`, `list()`, `get(id)`,
    `logs(id)` (SSE), `delete(id)`
  - streams → `syslog()`, `notifications()`, `ostrace(filters?)`, `listen()`, `sysmontap()`
- `client.tunnels` → `list()`, `delete(udid)`, `refresh(udid)`, `shutdownAgent()`

## Development

```sh
npm install
npm run generate   # regenerate src/generated from ../../spec/openapi/openapi.yaml
npm run typecheck
npm test
npm run build
```

The generated client under `src/generated/` is committed so the package is
self-contained; re-run `npm run generate` when the spec changes. The package name
is **`@go-ios/sdk`** (scoped, published with npm provenance).
