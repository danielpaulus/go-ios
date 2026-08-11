# GoIos.Sdk — C#/.NET SDK for go-ios

Ergonomic, async C#/.NET SDK for the [go-ios](https://github.com/danielpaulus/go-ios)
REST API. It covers the **full daemon surface** — device info & diagnostics,
apps, WebDriverAgent, condition inducers, location, developer images, files,
crashes, media (wallpaper/icon-layout/pasteboard), profiles, settings, MDM,
HTTP proxy, background jobs and tunnels — plus six live streams
(syslog, notifications, os_trace, listen, sysmontap, job-logs) as typed
`IAsyncEnumerable<SseEvent>`.

- **Low-level client:** generated from the OpenAPI 3.1 spec with
  [openapi-generator](https://openapi-generator.tech/) (`csharp`, `httpclient`
  library, `net8.0`, nullable, async). Lives under `src/Generated/`.
- **Facade:** a thin hand-written layer (`GoIos` / `GoIos.Sdk`) with an API shape
  shared across the go-ios SDKs.

## Install

```sh
dotnet add package GoIos.Sdk
```

Targets **net8.0**.

## Quickstart

### Unary calls

```csharp
using GoIos;

using var client = new IosClient(new IosClientOptions
{
    BaseUrl = "http://localhost:60105",
    ApiKey  = "your-go-ios-api-key", // optional if the server runs with --disable-auth
});

// List devices
var devices = await client.Devices.ListAsync();
foreach (var d in devices.VarDeviceList) // generated prop name for JSON "deviceList"
    Console.WriteLine(d.Properties.SerialNumber);

var device = client.Device("00008110-000123456789ABCD");
Console.WriteLine(device.Udid); // convenience accessor (properties.serialNumber)

// Info + screenshot (raw PNG bytes)
var info = await device.InfoAsync();
byte[] png = await device.ScreenshotAsync();
await File.WriteAllBytesAsync("shot.png", png);

// Device information
var name    = await device.DeviceNameAsync();
var date    = await device.DateAsync();
var battery = await device.BatteryAsync();
var diag    = await device.DiagnosticsAsync();
var gestalt = await device.MobileGestaltAsync(new[] { "ProductType", "UniqueDeviceID" });
var procs   = await device.ProcessesAsync(apps: true);
var lockdown = await device.LockdownAsync();

// Management
await device.RebootAsync();
await device.ShutdownAsync();
await device.EraseAsync(confirm: true);          // destructive
await device.SetDevmodeAsync("enable", enablePostRestart: true);
await device.SetLangAsync(language: "en", locale: "en_US");
await device.MemLimitOffAsync("MyApp");

// Files (AFC) — binary transfers stream through the raw HTTP pipeline
var listing = await device.Files.LsAsync(domain: "appDocuments", path: "Documents", identifier: "com.example.app");
byte[] pulled = await device.Files.PullAsync("appDocuments", "Documents/log.txt", "com.example.app");
await device.Files.PushAsync("appDocuments", "Documents/out.txt", pulled, "com.example.app");

// Crashes
var crashes = await device.Crashes.ListAsync("*.ips");
await device.Crashes.RemoveAsync("*.ips", cwd: "/tmp/crashes");

// Media
byte[] wallpaper = await device.Media.WallpaperAsync();
await device.Media.SetPasteboardAsync("hello");
var clip = await device.Media.PasteboardAsync();

// Settings
await device.Settings.SetAssistiveTouchAsync(true);
await device.Settings.SetTimeFormatAsync(uses24Hour: true);
await device.Settings.SetWifiAsync("MyNet", password: "hunter2", encType: "WPA2");
await device.Settings.RemoveWifiAsync("MyNet");

// Profiles / images
await device.AddProfileAsync(File.ReadAllBytes("wifi.mobileconfig"));
await device.RemoveProfileAsync("com.example.profile");
var mounted = await device.MountedImagesAsync();
await device.UnmountImageAsync();

// MDM (supervised — pass the supervision .p12)
byte[] p12 = File.ReadAllBytes("supervision.p12");
var security = await device.Mdm.SecurityInfoAsync(p12, password: "pass");
var token    = await device.Mdm.FetchUnlockTokenAsync(p12, "pass");
await device.Mdm.ClearPasscodeAsync(p12, token.Token, "pass");

// HTTP proxy
await device.Proxy.SetHttpProxyAsync("10.0.0.1", "8888", p12);
await device.Proxy.RemoveHttpProxyAsync();

// Background jobs (device-scoped)
var job = await device.Jobs.RunwdaAsync();
var jobs = await device.Jobs.ListAsync();
await device.Jobs.ForwardAsync(hostPort: 8100, targetPort: 8100);
await device.Jobs.DeleteAsync(job.Id);

// Tunnels (global)
var tunnels = await client.Tunnels.ListAsync();
await client.Tunnels.RefreshAsync(device.Udid);
await client.Tunnels.ShutdownAgentAsync();

// Apps
var apps = await device.Apps.ListAsync();
await device.Apps.LaunchAsync("com.apple.Preferences");
await device.Apps.KillAsync("com.apple.Preferences");
await device.Apps.InstallAsync("/path/to/MyApp.ipa");
await device.Apps.UninstallAsync("com.example.myapp");

// Location
await device.SetLocationAsync(latitude: 52.5200, longitude: 13.4050);
await device.ResetLocationAsync();

// Condition inducers
var conditions = await device.ConditionsAsync();
await device.EnableConditionAsync(profileTypeId: "…", profileId: "…");
await device.DisableConditionAsync();

// Developer disk image
var images = await device.ImagesAsync();
await device.InstallImageAsync(auto: true);

// WebDriverAgent (XCUITest)
var session = await device.Wda.CreateSessionAsync(new GoIos.Sdk.Generated.Model.WdaConfig(
    bundleId: "com.facebook.WebDriverAgentRunner.xctrunner",
    testBundleId: "com.facebook.WebDriverAgentRunner",
    xcTestConfig: "WebDriverAgentRunner.xctest"));
await device.Wda.ReadSessionAsync(session.SessionId);
await device.Wda.DeleteSessionAsync(session.SessionId);
```

### Streaming (Server-Sent Events)

The six long-lived endpoints are exposed as `IAsyncEnumerable<SseEvent>`. Each
typed event maps to an SSE `event:` name; `HeartbeatEvent` keep-alives are
surfaced so you can tell a live-but-idle stream from a dropped one, and any
unrecognized `event:` name arrives as `UnknownEvent` (never silently dropped).

```csharp
using var cts = new CancellationTokenSource();

await foreach (var e in device.SyslogAsync(cts.Token))
{
    switch (e)
    {
        case SyslogMessageEvent s: Console.WriteLine(s.Message); break;
        case HeartbeatEvent:       /* still alive */             break;
        case UnknownEvent u:       Console.WriteLine($"unknown {u.EventName}: {u.RawData}"); break;
    }
}

// Notifications (app lifecycle):
await foreach (var e in device.NotificationsAsync(cts.Token))
    if (e is AppStateNotificationEvent a) Console.WriteLine($"{a.BundleId} -> {a.State}");

// os_trace with AND-combined filters:
var filters = new OsTraceFilters { Level = "error", Subsystem = "com.apple.network" };
await foreach (var e in device.OsTraceAsync(filters, cts.Token))
    if (e is OsTraceEntryEvent t) Console.WriteLine($"[{t.ProcessName}] {t.Message}");

// Device attach/detach/pair:
await foreach (var e in device.ListenAsync(cts.Token))
    if (e is AttachDetachEventEvent ad) Console.WriteLine($"{ad.Event} {ad.Udid}");

// sysmontap CPU-usage samples (open map — extra sampler keys land in `Extra`):
await foreach (var e in device.SysmontapAsync(cts.Token))
    if (e is CpuUsageSampleEvent s) Console.WriteLine($"CPU {s.CpuTotalLoad}%");

// Live job logs:
await foreach (var e in device.Jobs.LogsAsync(job.Id, cts.Token))
    if (e is JobLogLineEvent l) Console.Write(l.Line);
```

Cancel the `CancellationToken` (or `break`) to stop a stream and release the
connection.

## Authentication

Every `/api/v1` route expects a bearer token
(`Authorization: Bearer <GO_IOS_API_KEY>`). Set `ApiKey` on `IosClientOptions`;
the SDK sends it on every request. It is optional only when the server is
launched with `--disable-auth`, but supplying it whenever you have it is
strongly encouraged.

## Streaming event types

| Endpoint             | Method                | Typed event                | SSE `event:`   |
| -------------------- | --------------------- | -------------------------- | -------------- |
| `/syslog`            | `SyslogAsync`         | `SyslogMessageEvent`       | `syslog`       |
| `/notifications`     | `NotificationsAsync`  | `AppStateNotificationEvent`| `appstate`     |
| `/ostrace`           | `OsTraceAsync`        | `OsTraceEntryEvent`        | `ostrace`      |
| `/listen`            | `ListenAsync`         | `AttachDetachEventEvent`   | `attachdetach` |
| `/sysmontap`         | `SysmontapAsync`      | `CpuUsageSampleEvent`      | `sample`       |
| `/jobs/{id}/logs`    | `Jobs.LogsAsync`      | `JobLogLineEvent`          | `log`          |
| *(all streams)*      | —                     | `HeartbeatEvent`           | `heartbeat`    |
| *(forward-compat)*   | —                     | `UnknownEvent`             | *(any other)*  |

## Notes on coverage

All 80 daemon operations are exposed. Two conveniences are intentionally absent
because the daemon has **no corresponding endpoint**: there is no set-device-name
or set-date route (only `GET /devicename` and `GET /date`), so the SDK offers
`DeviceNameAsync`/`DateAsync` but no setters.

## Regenerating the low-level client

```sh
./regen.sh   # requires Java (openapi-generator) + npx
```

Regenerates `src/Generated/` from `../../spec/openapi/openapi.yaml`. The facade
under `src/GoIos.Sdk/` is hand-written and is not overwritten.
