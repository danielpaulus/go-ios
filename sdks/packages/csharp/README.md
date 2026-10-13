# GoIos.Sdk — C#/.NET SDK for go-ios

Ergonomic, async C#/.NET SDK for the [go-ios](https://github.com/danielpaulus/go-ios)
REST API. It covers the **full daemon surface** (125 operations) — device info &
diagnostics (disk space, IP, RSD, battery registry), apps, WebDriverAgent, UI
automation, WebInspector, accessibility (VoiceOver / Zoom / AX audit), condition
inducers, location (incl. GPX), developer images, files (AFC + fsync), crashes,
media (wallpaper/icon-layout/pasteboard), profiles, settings, MDM, HTTP proxy,
device preparation & supervision, host-side code signing, background jobs and
tunnels — plus six live SSE streams (syslog, notifications, os_trace, listen,
sysmontap, job-logs) as typed `IAsyncEnumerable<SseEvent>` and three raw binary
streams (UI video, MJPEG screenshots, pcap) as `Stream`.

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

## Connecting (daemon discovery)

By default the go-ios REST daemon binds an **ephemeral, loopback-only** port and
writes a discovery file at `<home>/rest-api.json` after it starts. The SDK reads
that file so `new IosClient()` "just works" against a locally running daemon —
no port to hardcode.

`BaseUrl` is optional. When it is not set, the base URL is resolved in this order:

1. **explicit `Options.BaseUrl`** — used verbatim (for remote daemons); discovery
   is skipped;
2. **`GO_IOS_BASE_URL`** environment variable;
3. **discovery** — read `baseUrl` from `<home>/rest-api.json`;
4. **none found** → a `DaemonNotFoundException` is thrown with a clear message
   ("no local go-ios REST daemon found at `<path>`; start it … or pass an explicit
   BaseUrl").

The **home** directory is `GO_IOS_HOME` (if set and non-empty), otherwise
`~/.go-ios` (the user profile dir; `%USERPROFILE%\.go-ios` on Windows). To pin the
daemon to a fixed port instead of the ephemeral default, run it with `--addr :8080`
and/or set `GO_IOS_BASE_URL`.

The `ApiKey` is independent of discovery: set it on `Options.ApiKey` (see
[Authentication](#authentication)). It is **not** read from the discovery file.

You can also call the discovery helper directly:

```csharp
string baseUrl = GoIos.Discovery.DiscoverBaseUrl(); // throws DaemonNotFoundException if none
```

## Quickstart

### Unary calls

```csharp
using GoIos;

// No BaseUrl: the SDK auto-discovers a local go-ios REST daemon (see below).
using var client = new IosClient(new IosClientOptions
{
    ApiKey = "your-go-ios-api-key", // optional if the server runs with --disable-auth
});

// ...or, equivalently, with no options at all:
using var discovered = new IosClient();

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
var lockdown = await device.LockdownAsync();                         // or LockdownAsync(domain: "com.apple.mobile.battery")

// Diagnostics / network
var disk    = await device.DiskSpaceAsync();
var ip      = await device.IpAsync();
var rsd     = await device.RsdAsync();
var batReg  = await device.BatteryRegistryAsync();

// Accessibility
await device.SetVoiceOverAsync(true);
await device.SetZoomAsync(false);
var axIssues = await device.AxAuditAsync(timeout: 60);
var axTree   = await device.AxAsync();
await device.SetLocationGpxAsync(File.ReadAllBytes("route.gpx"));
var cloudCfg = await device.CloudConfigAsync();

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

// fsync (ios fsync ...) — path + optional bundleId scope
var tree = await device.Fsync.TreeAsync(path: "/Documents", bundleId: "com.example.app");
byte[] f = await device.Fsync.PullAsync("/Documents/log.txt", bundleId: "com.example.app");
await device.Fsync.PushAsync("/Documents/out.txt", f, bundleId: "com.example.app");
await device.Fsync.MkdirAsync("/Documents/sub", bundleId: "com.example.app");
await device.Fsync.RmAsync("/Documents/old", recursive: true, bundleId: "com.example.app");

// UI automation (WDA / DeviceKit) — optional backend/wdaUrl/timeout via Options
var ui = device.Ui;
await ui.TapAsync(100, 200);
await ui.SwipeAsync(10, 400, 10, 100, duration: 0.5);
await ui.LongPressAsync(100, 200, duration: 1.0);
await ui.TypeAsync("hello");
await ui.ButtonAsync("home");
byte[] uiShot = await ui.ScreenshotAsync();
string source = await ui.SourceAsync();
var size = await ui.SizeAsync();
await ui.SetOrientationAsync("landscape");
await ui.AppLaunchAsync("com.apple.Preferences", new UiClient.Options { Backend = "devicekit", Timeout = 30 });
var raw = await ui.ApiAsync(method: "GET", path: "/status");

// WebInspector (remote web debugging)
var pages = await device.WebInspector.PagesAsync();
await device.WebInspector.LaunchAsync(url: "https://example.com");
var eval = await device.WebInspector.EvalAsync("document.title", page: pages[0]["id"]?.ToString());

// Device preparation (multipart; supply a supervision cert to supervise)
var prep = await device.PrepareAsync(cert: File.ReadAllBytes("supervision.p12"),
                                     p12Password: "pass", skip: new[] { "Siri" }, orgName: "Acme");

// Host-side code signing / preparation (device-free)
var skipOpts = await client.Prepare.SkipOptionsAsync();
var superCert = await client.Prepare.CreateCertAsync();
byte[] p12Cert = await client.Sign.CertificateAsync(File.ReadAllBytes("AuthKey.p8"), "KEYID", "ISSUER");
byte[] signedIpa = await client.Sign.AppAsync(File.ReadAllBytes("app.ipa"),
                                              File.ReadAllBytes("id.p12"),
                                              File.ReadAllBytes("app.mobileprovision"));

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

### Binary streams (raw bytes, not SSE)

Three endpoints emit an open-ended stream of raw bytes rather than SSE frames:
live UI video (MJPEG / H.264), MJPEG screenshots, and a libpcap capture. Each
returns a `BinaryStream` — a read-only `Stream` opened with
`HttpCompletionOption.ResponseHeadersRead` so bytes are pulled off the socket as
they arrive. Reads honor the `CancellationToken`; dispose the stream to stop the
capture and release the connection. `ContentType` exposes the negotiated media
type.

```csharp
using var cts = new CancellationTokenSource();

// pcap → pipe straight to a file (or into wireshark/tshark)
await using (var pcap = await device.PcapAsync(timeout: 30, cancellationToken: cts.Token))
await using (var file = File.Create("capture.pcap"))
    await pcap.CopyToAsync(file, cts.Token);

// MJPEG screenshot stream
await using var shots = await device.ScreenshotStreamAsync(quality: 80, cancellationToken: cts.Token);

// UI video (MJPEG default; codec: "h264" needs the devicekit backend)
await using var video = await device.Ui.StreamAsync(
    new UiClient.Options { Backend = "devicekit" }, codec: "h264", cancellationToken: cts.Token);
Console.WriteLine(video.ContentType);
```

## Examples

Runnable, heavily-commented examples live in [`examples/`](./examples). They
double as documentation and as a pre-release smoke test: `examples/run.sh`
(`dotnet run --project examples/GoIos.Examples -- run-all`) drives the read-only
surface of a live daemon (list devices, device info, list apps, screenshot,
stream syslog) and exits non-zero on any failure. Steps that need a device — or
a forwarded WebDriverAgent for the optional UI example — print `SKIP` instead of
failing. See [`examples/README.md`](./examples/README.md) for setup and the full
command list.

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

All 125 daemon operations are exposed. Two conveniences are intentionally absent
because the daemon has **no corresponding endpoint**: there is no set-device-name
or set-date route (only `GET /devicename` and `GET /date`), so the SDK offers
`DeviceNameAsync`/`DateAsync` but no setters.

Endpoints whose response is an open, schema-less JSON object (RSD services,
cloud config, AX snapshot/audit, WebInspector pages, and the UI backend
passthrough responses) are surfaced as `IReadOnlyDictionary<string, object?>`
(or a list of them) so no data is lost to a fixed DTO.

## Regenerating the low-level client

```sh
./regen.sh   # requires Java (openapi-generator) + npx
```

Regenerates `src/Generated/` from `../../spec/openapi/openapi.yaml`. The facade
under `src/GoIos.Sdk/` is hand-written and is not overwritten.
