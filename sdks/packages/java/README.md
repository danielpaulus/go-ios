# go-ios SDK (Java)

Java 17 SDK for the [go-ios](https://github.com/danielpaulus/go-ios) REST API. It
covers the **full 80-operation daemon surface**: device lifecycle, apps,
WebDriverAgent sessions, device info, device management, files & crashes, media,
profiles, settings, MDM, the HTTP proxy, the tunnel agent, async jobs, and typed
Server-Sent Event streams (syslog, notifications, os_trace, listen, sysmontap,
and job logs).

- **Generated low-level client:** [openapi-generator](https://openapi-generator.tech/)
  `7.11.0`, `java` generator, **`native`** HTTP library (`java.net.http`),
  consuming `spec/openapi/openapi.yaml` (OpenAPI 3.1). Generated sources live in
  [`generated/`](generated/) and are committed; regenerate with
  [`scripts/generate.sh`](scripts/generate.sh).
- **Ergonomic facade:** a thin hand-written layer (`com.github.danielpaulus.goios`)
  with a public API aligned to the other go-ios SDKs.
- **Streaming:** a reusable SSE reader over `java.net.http` that parses
  `text/event-stream` frames, JSON-decodes each `data:` payload into a typed
  event, skips heartbeats, surfaces unknown events, and is cancellable
  (`AutoCloseable`). See [`docs/DESIGN.md`](../../docs/DESIGN.md) for the wire
  contract.

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

## Authentication

All `/api/v1` routes require `Authorization: Bearer <GO_IOS_API_KEY>`. Pass the
key via the builder. It is optional (a server started with `--disable-auth`
accepts requests without it) but strongly encouraged; the SDK sends it whenever
it is set.

```java
IosClient client = IosClient.builder()
        .baseUrl("http://localhost:8080")          // /api/v1 is appended automatically
        .apiKey(System.getenv("GO_IOS_API_KEY"))   // optional but recommended
        .build();
```

## Quickstart — unary calls

```java
import com.github.danielpaulus.goios.*;
import com.github.danielpaulus.goios.generated.model.*;

try (IosClient client = IosClient.builder()
        .baseUrl("http://localhost:8080")
        .apiKey(System.getenv("GO_IOS_API_KEY"))
        .build()) {

    // Fleet
    for (DeviceEntry d : client.devices().list()) {
        System.out.println(d.getProperties().getSerialNumber());
    }

    // One device
    Device device = client.device("00008110-0011...");
    Object info      = device.info();
    byte[] png       = device.screenshot();               // image/png bytes
    device.setLocation(37.3349, -122.0090);               // correctly-spelled longitude
    device.resetLocation();

    // Apps
    device.apps().install("app.ipa", Files.readAllBytes(Path.of("app.ipa")));
    device.apps().launch("com.apple.Preferences");
    for (AppInfo app : device.apps().list()) {
        System.out.println(app.getCfBundleIdentifier());
    }
    device.apps().kill("com.apple.Preferences");

    // WebDriverAgent
    WdaConfig cfg = new WdaConfig()
            .bundleId("com.facebook.WebDriverAgentRunner.xctrunner")
            .testBundleId("com.facebook.WebDriverAgentRunner.xctrunner")
            .xcTestConfig("WebDriverAgentRunner.xctest");
    WdaSession session = device.wda().createSession(cfg);
    device.wda().readSession(session.getSessionId());
    device.wda().deleteSession(session.getSessionId());
}
```

Non-2xx responses throw {@code IosApiException} carrying the HTTP status and the
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

Each streaming method returns an `SseReader`, which is both an
`Iterable<SseEvent>` and an `AutoCloseable`. Use pattern matching to branch on
the typed event, and try-with-resources to guarantee the underlying HTTP stream
is cancelled:

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

Available streams and their typed events (heartbeats are parsed and skipped):

| Method                                         | Event type          | Payload                 |
| ---------------------------------------------- | ------------------- | ----------------------- |
| `device.syslog()`                              | `SyslogEvent`       | `SyslogMessage`         |
| `device.notifications()`                       | `AppStateEvent`     | `AppStateNotification`  |
| `device.ostrace(pid, level, subsystem, m, x)`  | `OsTraceEvent`      | `OsTraceEntry`          |
| `device.listen()`                              | `AttachDetachEvent` | attach/detach payload   |
| `device.sysmontap()`                           | `SysmontapEvent`    | `CpuUsageSample`        |
| `device.jobs().logs(jobId)`                    | `JobLogEvent`       | `JobLogLine`            |

Any unrecognized `event:` name is surfaced as `UnknownEvent` (never dropped) for
forward-compatibility. `device.ostrace()` (no args) streams unfiltered; the
filtered overload AND-combines any non-null filters.

## Full API surface

The facade groups all 80 operations idiomatically. `client.device(udid)` returns
a `Device`; a few groups are reached through sub-facades.

```java
Device d = client.device(udid);

// Device info (read-only)
d.deviceName(); d.date(); d.battery(); d.diagnostics();
d.mobileGestalt(List.of("ProductType")); d.processes(); d.lockdown();

// Device management
d.reboot(); d.shutdown(); d.erase(true);
d.devmode(); d.setDevmode("enable", true);
d.lang(); d.setLang("en", "en_US"); d.memlimitoff("backboardd");

// Developer image
d.images(); d.installImage(bytes); d.installImageAuto(basedir);
d.mountedImages(); d.unmountImage();

// Profiles
d.profiles(); d.addProfile(mobileconfig, p12, pass); d.removeProfile(name);

// Files & crashes
d.files().ls("app", "com.x", "/Documents");
byte[] f = d.files().pull("app", "com.x", "/Documents/log.txt");
d.files().push("app", "com.x", "/Documents/out.txt", bytes);
d.crashes().list(); d.crashes().remove("*.crash"); d.crashes().remove("*.crash", cwd);

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

// HTTP proxy (supervised)
d.proxy().setHttpProxy(host, port, user, pass, p12, p12Pass);
d.proxy().removeHttpProxy();

// Async jobs (device-scoped)
Job job = d.jobs().runwda(new RunTestRequest());
d.jobs().runtest(req); d.jobs().forward(8080, 9090);
d.jobs().list(); d.jobs().get(job.getId()); d.jobs().delete(job.getId());
try (SseReader logs = d.jobs().logs(job.getId())) { /* stream */ }

// Tunnel agent (fleet-level, not device-scoped)
client.tunnels().list();
client.tunnels().refresh(udid);
client.tunnels().delete(udid);
client.tunnels().shutdownAgent();

// Convenience: udid from a DeviceEntry (properties.serialNumber)
String udid = Devices.udid(client.devices().list().get(0));
```

## Build & test

Maven (recommended):

```bash
mvn -q package                  # compile facade + committed generated sources, run tests
mvn -q -DskipTests package      # compile only
```

Without Maven (JDK 17+ only), a helper compiles with `javac` and runs the suite
via the JUnit Platform Console Standalone launcher (dependency jars are
downloaded once into `.tools/lib/`, gitignored):

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
