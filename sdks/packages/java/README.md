# go-ios SDK (Java)

Java 17 SDK for the [go-ios](https://github.com/danielpaulus/go-ios) REST API. It
covers the **full 125-operation daemon surface**: device lifecycle, apps,
WebDriverAgent sessions, device info & network/disk diagnostics, device
management, house-arrest files & crashes, AFC file transfer (`fsync`), media,
profiles, settings, accessibility (VoiceOver / Zoom / audit / element snapshot),
MDM, the HTTP proxy, UI automation, the Safari Web Inspector, host-scoped app
signing & device preparation, the tunnel agent, async jobs, typed Server-Sent
Event streams (syslog, notifications, os_trace, listen, sysmontap, job logs), and
raw binary streams (UI video, MJPEG screenshots, pcap).

- **Generated low-level client:** [openapi-generator](https://openapi-generator.tech/)
  `7.11.0`, `java` generator, **`native`** HTTP library (`java.net.http`),
  consuming `spec/openapi/openapi.yaml` (OpenAPI 3.1). Generated sources live in
  [`generated/`](generated/) and are committed; regenerate with
  [`scripts/generate.sh`](scripts/generate.sh).
- **Ergonomic facade:** a thin hand-written layer (`com.github.danielpaulus.goios`)
  driving `java.net.http` directly, with a public API aligned to the other go-ios
  SDKs (TypeScript / Python / C#).
- **Streaming:** two seams — a typed SSE reader for `text/event-stream`
  endpoints, and a raw `BinaryStream` (an `InputStream`) for the `x-stream:
  binary` endpoints. Both are `AutoCloseable` and cancel the underlying HTTP
  connection on close.

## Install

Maven:

```xml
<dependency>
  <groupId>com.github.danielpaulus</groupId>
  <artifactId>go-ios-sdk</artifactId>
  <version>0.1.0</version>
</dependency>
```

Gradle:

```groovy
implementation("com.github.danielpaulus:go-ios-sdk:0.1.0")
```

## Connecting — daemon discovery

`baseUrl` is **optional**. The go-ios REST daemon now binds an **ephemeral
loopback port by default** and writes a discovery file at `<home>/rest-api.json`;
the SDK reads it so you don't have to know the port. When you build a client with
no explicit `baseUrl`, the endpoint is resolved in this order:

1. an explicit `.baseUrl(...)` on the builder (use this for remote daemons);
2. the `GO_IOS_BASE_URL` env var;
3. **discovery** — the `baseUrl` field of `<home>/rest-api.json`, where `<home>`
   is `GO_IOS_HOME` if set and non-empty, else `~/.go-ios`;
4. otherwise `build()` throws `IosDiscoveryException` naming the expected path
   (start the go-ios REST API, or pass a `baseUrl`).

```java
// Auto-discover a locally running daemon (no baseUrl needed):
try (IosClient client = IosClient.builder()
        .apiKey(System.getenv("GO_IOS_API_KEY"))
        .build()) {
    // ...
}
```

To pin the daemon to a fixed port instead, start it with `--addr :8080` and/or
set `GO_IOS_BASE_URL` (or pass `.baseUrl("http://localhost:8080")`).

## Authentication

All `/api/v1` routes require `Authorization: Bearer <GO_IOS_API_KEY>`. Pass the
key via the builder (or set `GO_IOS_API_KEY` and read it yourself). It is
optional (a server started with `--disable-auth` accepts requests without it)
but strongly encouraged; the SDK sends it whenever it is set. The API key is
**not** read from the discovery file.

```java
IosClient client = IosClient.builder()
        .baseUrl("http://localhost:8080")          // optional; /api/v1 is appended automatically
        .apiKey(System.getenv("GO_IOS_API_KEY"))   // optional but recommended
        .build();
```

## Quickstart — unary calls

```java
import com.github.danielpaulus.goios.*;
import com.github.danielpaulus.goios.generated.model.*;

try (IosClient client = IosClient.builder()
        .apiKey(System.getenv("GO_IOS_API_KEY"))   // baseUrl auto-discovered
        .build()) {

    // Fleet
    for (DeviceEntry d : client.devices().list()) {
        System.out.println(Devices.udid(d));
    }

    // One device
    Device device = client.device("00008110-0011...");
    Object info      = device.info();
    byte[] png       = device.screenshot();               // image/png bytes
    device.setLocation(37.3349, -122.0090);               // correctly-spelled longitude
    device.resetLocation();

    // Apps
    device.apps().install(Files.readAllBytes(Path.of("app.ipa")));
    device.apps().launch("com.apple.Preferences");
    for (AppInfo app : device.apps().list()) {
        System.out.println(app.getCfBundleIdentifier());
    }
    device.apps().kill("com.apple.Preferences");

    // WebDriverAgent
    WdaConfig cfg = new WdaConfig()
            .bundleId("com.facebook.WebDriverAgentRunner.xctrunner");
    WdaSession session = device.wda().createSession(cfg);
    device.wda().getSession(session.getSessionId());
    device.wda().deleteSession(session.getSessionId());
}
```

Non-2xx responses throw `IosApiException` carrying the HTTP status and the
decoded `GenericResponse` error envelope:

```java
try {
    client.device("does-not-exist").info();
} catch (IosApiException e) {
    if (e.statusCode() == 404) {
        System.out.println("unknown device: " + e.errorBody().getError());
    }
}
```

## Quickstart — streaming (SSE)

Each SSE method returns an `SseReader`, which is both an `Iterable<SseEvent>` and
an `AutoCloseable`. Use pattern matching to branch on the typed event, and
try-with-resources to guarantee the underlying HTTP stream is cancelled:

```java
import com.github.danielpaulus.goios.stream.*;

try (SseReader stream = device.syslog()) {
    for (SseEvent ev : stream) {
        if (ev instanceof SyslogEvent s) {
            System.out.println(s.payload().getMessage());
        }
        if (someStopCondition) {
            break; // closing the try-with-resources aborts the stream
        }
    }
}
```

Available SSE streams and their typed events (heartbeats are parsed and skipped
by default; pass `true` to include them):

| Method                                              | Event type          | Payload                 |
| --------------------------------------------------- | ------------------- | ----------------------- |
| `device.syslog()`                                   | `SyslogEvent`       | `SyslogMessage`         |
| `device.notifications()`                            | `AppStateEvent`     | `AppStateNotification`  |
| `device.ostrace(pid, level, subsystem, m, x, hb)`   | `OsTraceEvent`      | `OsTraceEntry`          |
| `device.listen()`                                   | `AttachDetachEvent` | attach/detach payload   |
| `device.sysmontap()`                                | `SysmontapEvent`    | `CpuUsageSample`        |
| `device.jobs().logs(jobId)`                         | `JobLogEvent`       | `JobLogLine`            |

Any unrecognized `event:` name is surfaced as `UnknownEvent` (never dropped) for
forward-compatibility. `device.ostrace()` (no args) streams unfiltered.

## Quickstart — binary streams

The `x-stream: binary` endpoints (UI video, MJPEG screenshots, live pcap) are
**not** SSE — they return an opaque byte stream. The SDK exposes them as a
`BinaryStream`, a plain `InputStream` the caller reads directly. Closing it
releases (cancels) the HTTP connection, so a long-lived capture can be stopped
at any time:

```java
import com.github.danielpaulus.goios.stream.BinaryStream;

// Live pcap capture piped to a file (stop after `timeout` seconds server-side).
try (BinaryStream pcap = device.pcap(30);
     var out = Files.newOutputStream(Path.of("capture.pcap"))) {
    pcap.transferTo(out);
}

// MJPEG screenshot stream / UI video stream.
try (BinaryStream video = device.ui().stream()) {
    System.out.println(video.contentType());   // e.g. multipart/x-mixed-replace
    byte[] frameBytes = video.readNBytes(64 * 1024);
}

try (BinaryStream shots = device.screenshotStream(80 /* quality */)) { /* ... */ }
```

## Full API surface

`client.device(udid)` returns a `Device`; grouped operations are reached through
sub-facades. Host-scoped (device-free) operations hang off the client.

```java
Device d = client.device(udid);

// Device info & diagnostics
d.info(); d.deviceName(); d.date(); d.battery(); d.batteryRegistry();
d.diagnostics(); d.diskSpace(); d.ip(); d.rsd();
d.mobileGestalt(List.of("ProductType")); d.processes(null);
d.lockdown(); d.lockdown("com.apple.mobile.battery");   // no-arg or domain-scoped

// Device management
d.activate(); d.reboot(); d.shutdown(); d.erase(true);
d.devMode(); d.setDevMode("enable", true);
d.lang(); d.setLang("en", "en_US"); d.memlimitoff("backboardd");

// Location
d.setLocation(37.3349, -122.0090); d.resetLocation();
d.setLocationGpx(Files.readAllBytes(Path.of("track.gpx")));   // multipart

// Accessibility
d.ax();                              // focused element snapshot
d.axAudit(60);                       // run the a11y audit (timeout seconds)
d.voiceOver(); d.setVoiceOver(true);
d.zoom(); d.setZoom(true);
d.resetAccessibility();

// Developer image
d.images(); d.mountImage(bytes); d.mountImageAuto(basedir);
d.mountedImages(); d.unmountImage();

// Profiles & conditions
d.profiles(); d.addProfile(mobileconfig, p12, pass); d.removeProfile(name);
d.conditions(); d.enableCondition(profileTypeId, profileId); d.disableCondition();

// House-arrest files & crashes
d.files().ls("app", "com.x", "/Documents");
byte[] f = d.files().pull("app", "com.x", "/Documents/log.txt");
d.files().push("app", "com.x", "/Documents/out.txt", bytes);
d.crashes().list(); d.crashes().remove("*.crash"); d.crashes().remove("*.crash", cwd);

// AFC file transfer (fsync); pass a bundleId to scope to an app container
d.fsync().ls("/Documents", "com.x"); d.fsync().tree("/Documents", null);
byte[] b = d.fsync().pull("/Documents/a.txt", null);
d.fsync().push("/Documents/x.bin", bytes, null);
d.fsync().mkdir("/Documents/new", null); d.fsync().rm("/Documents/old", null, true);
d.cloudConfig();

// Media
byte[] wp = d.media().wallpaper();
d.media().setWallpaper(image, p12, pass, "home");   // supervised multipart
d.media().iconLayout(); d.media().setIconLayout(layout);
d.media().pasteboard(); d.media().setPasteboard("copied");

// Settings
d.settings().assistiveTouch(); d.settings().setAssistiveTouch(true);
d.settings().timeFormat(); d.settings().setTimeFormat(true);
d.settings().setWifi("ssid", "pw", "WPA2"); d.settings().removeWifi("ssid");

// MDM (supervised; each takes a .p12 identity)
d.mdm().securityInfo(p12, pass);
d.mdm().fetchUnlockToken(p12, pass);
d.mdm().clearPasscode(p12, pass, token);
d.mdm().clearScreenTimePassword(p12, pass);

// HTTP proxy & pairing / prepare (supervised, multipart)
d.setHttpProxy(host, port, user, pass, p12, p12Pass); d.removeHttpProxy();
d.pair(true, p12, supervisionPassword);
d.prepare(cert, p12password, List.of("Passcode", "Siri"), orgname, "en_US", "en");

// UI automation (backend/wdaUrl/timeout via Ui.Options; convenience overloads use defaults)
d.ui().tap(100, 200);
d.ui().swipe(10, 10, 300, 300, 0.5, null);
d.ui().longPress(50, 50);
d.ui().type("hello");
d.ui().button("home");
byte[] uiShot = d.ui().screenshot();
d.ui().source(); d.ui().size(); d.ui().status();
d.ui().orientation(); d.ui().setOrientation("LANDSCAPE");
d.ui().appLaunch(bundleId); d.ui().appTerminate(bundleId); d.ui().appForeground();
d.ui().api(rawBackendBody, new Ui.Options("devicekit", null, 60));

// Safari Web Inspector
d.webinspector().pages();
d.webinspector().launch("https://example.com", null);
d.webinspector().eval("document.title", pageId, null);

// Async jobs (device-scoped)
Job job = d.jobs().runWda(new RunTestRequest());
d.jobs().runTest(req); d.jobs().forward(8080, 9090);
d.jobs().list(); d.jobs().get(job.getId()); d.jobs().delete(job.getId());
try (SseReader logs = d.jobs().logs(job.getId())) { /* stream */ }

// Tunnel agent (fleet-level)
client.tunnels().list();
client.tunnels().refresh(udid);
client.tunnels().delete(udid);
client.tunnels().shutdownAgent();

// Host-scoped app signing (device-free)
byte[] signedIpa = client.sign().app(ipa, p12, profile, p12pass, bundleId);
byte[] p12Cert   = client.sign().certificate(ascKeyP8, keyId, issuerId, false, p12pass);
ProvisioningResult prov = client.sign().provision(
        ascKeyP8, keyId, issuerId, bundleId, udid,
        null, null, null, null, false, p12pass);

// Host-scoped preparation helpers
client.prepare().createCert();     // self-signed supervision cert + key
client.prepare().skipOptions();    // setup panes that prepare can skip
```

## Examples

Runnable, heavily commented example programs live in [`examples/`](examples/) —
each a standalone `main`, configured via `GO_IOS_BASE_URL` / `GO_IOS_API_KEY` /
`GO_IOS_UDID`. They double as documentation and as a pre-release smoke test:
[`examples/RunAllExamples.java`](examples/RunAllExamples.java) runs listing
devices, device info, apps, a screenshot and an SSE syslog stream in sequence
(plus an optional UI-automation example gated on `RUN_UI=1`) and exits non-zero
if any core step fails. Compile and run them without Maven:

```bash
export GO_IOS_API_KEY=...            # required
bash examples/run.sh                 # compile + run all examples
bash examples/run.sh --compile-only  # compile only; no daemon needed
```

See [`examples/README.md`](examples/README.md) for the full list and setup.

## Build & test

Maven (recommended):

```bash
mvn -q package                  # compile facade + committed generated sources, run tests
mvn -q -DskipTests package      # compile only
```

Without Maven (JDK 17+ only), a helper compiles with `javac --release 17` and
runs the suite via the JUnit Platform Console Standalone launcher (dependency
jars are downloaded once into `.tools/lib/`, gitignored):

```bash
./scripts/verify.sh
```

Regenerate the low-level client from the spec:

```bash
./scripts/generate.sh           # pins openapi-generator-cli 7.11.0
```

## Publishing (maintainers)

Configured for **Maven Central** via the Sonatype Central Publisher Portal under
the `release` profile (`central-publishing-maven-plugin` + `maven-gpg-plugin`).
No credentials are stored here. To publish a release you would supply a
`central` server entry in `~/.m2/settings.xml` and a GPG signing key, then run
`mvn -Prelease clean deploy`.
