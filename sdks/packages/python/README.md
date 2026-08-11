# go-ios SDK (Python)

Ergonomic Python client for the [go-ios](https://github.com/danielpaulus/go-ios)
REST API — full coverage of the go-ios daemon surface (80 endpoints): list and
drive iOS devices, read device info/diagnostics, manage apps, files, crashes,
profiles, images, settings, MDM and proxy, run asynchronous jobs (XCUITest /
WebDriverAgent / port forwards), manage RemoteXPC tunnels, and consume all six
live Server-Sent-Event streams as typed async/sync generators.

- **PyPI package:** `go-ios-sdk`
- **Import name:** `go_ios_sdk`
- **Requires:** Python 3.9+
- Sync (`IosClient`) and async (`AsyncIosClient`) clients, built on
  [httpx](https://www.python-httpx.org/).

## Install

```bash
pip install go-ios-sdk
# or
uv add go-ios-sdk
```

## Auth

The go-ios server uses bearer auth (`Authorization: Bearer <GO_IOS_API_KEY>`).
Pass your key as `api_key`; it is sent on every request. A server started with
`--disable-auth` needs no key, but supplying one is still safe and encouraged.

```python
from go_ios_sdk import IosClient

client = IosClient(base_url="http://localhost:60105", api_key="my-key")
```

## Quickstart (sync, unary calls)

```python
from go_ios_sdk import IosClient

with IosClient(base_url="http://localhost:60105", api_key="my-key") as client:
    # Fleet
    for entry in client.devices.list()["deviceList"]:
        print(entry["properties"]["serialNumber"])

    dev = client.device("00008110-0000000000000000")

    print(dev.info())                       # lockdown + instruments values
    png: bytes = dev.screenshot()           # raw PNG bytes
    with open("shot.png", "wb") as f:
        f.write(png)

    dev.set_location(latitude=52.5200, longitude=13.4050)
    dev.reset_location()

    # Apps
    dev.apps.list()
    dev.apps.launch("com.apple.Preferences")
    dev.apps.kill("com.apple.Preferences")
    dev.apps.install("MyApp.ipa")           # path, bytes, or file-like
    dev.apps.uninstall("com.example.MyApp")

    # Conditions / profiles / images
    dev.conditions()
    dev.enable_condition(profile_type_id="...", profile_id="...")
    dev.disable_condition()
    dev.profiles()
    dev.images()
    dev.install_image(auto=True)            # or install_image(b"...bytes...")

    # WebDriverAgent (XCUITest)
    session = dev.wda.create_session({
        "bundleId": "com.facebook.WebDriverAgentRunner.xctrunner",
        "testBundleId": "com.facebook.WebDriverAgentRunner",
        "xcTestConfig": "WebDriverAgentRunner.xctest",
    })
    dev.wda.read_session(session["sessionId"])
    dev.wda.delete_session(session["sessionId"])
```

## Full endpoint surface

The facade groups all 80 daemon operations (same grouping and snake_case names
as the TypeScript / Java / C# SDKs). Sync and async clients are identical in
shape; async methods are coroutines and streams are async generators.

```python
# Fleet & tunnels (client-level)
client.devices.list(); client.devices.udids()      # udids() -> serialNumbers
client.tunnels.list(); client.tunnels.refresh(udid)
client.tunnels.delete(udid); client.tunnels.shutdown_agent()

dev = client.device(udid)

# Device info
dev.info(); dev.device_name()
dev.date(); dev.battery(); dev.diagnostics()
dev.mobilegestalt(keys=["ProductName", "BuildVersion"])
dev.processes(apps=True); dev.lockdown()

# Management
dev.reboot(); dev.shutdown(); dev.erase(confirm=True)
dev.devmode(); dev.set_devmode("enable", enable_post_restart=True)
dev.lang(); dev.set_lang(language="en", locale="en_US")
dev.memlimitoff("MyApp")

# Files & crashes
dev.files.ls("/Documents", domain="app", identifier="com.example.MyApp")
data: bytes = dev.files.pull("/Documents/log.txt", domain="app")
dev.files.push("/Documents/log.txt", b"...bytes...", domain="app")
dev.crashes.list("*"); dev.crashes.remove("MyApp-*", cwd=".")

# Media
wallpaper: bytes = dev.get_wallpaper()
dev.set_wallpaper("wall.png", "supervision.p12", screen="both")
dev.get_icon_layout(); dev.set_icon_layout({...})
dev.get_pasteboard(); dev.set_pasteboard("clipboard text")

# Profiles & images
dev.add_profile("config.mobileconfig", p12="id.p12", password="pw")
dev.remove_profile("com.example.profile")
dev.mounted_images(); dev.unmount_image()

# Settings
dev.assistive_touch(); dev.set_assistive_touch(True)
dev.time_format(); dev.set_time_format(True)          # 24-hour clock
dev.set_wifi("MySSID", password="pw", enc_type="WPA2")
dev.remove_wifi("MySSID")

# MDM (supervised)
dev.security_info("id.p12", password="pw")
dev.fetch_unlock_token("id.p12")
dev.clear_passcode("id.p12", token="<escrow-token>")
dev.clear_screen_time_password("id.p12")

# HTTP proxy (supervised)
dev.set_http_proxy("proxy.local", 8080, user="u", password="p")
dev.remove_http_proxy()

# Asynchronous jobs
job = dev.jobs.runtest({"bundleId": "com.example.MyAppUITests.xctrunner"})
dev.jobs.runwda({"bundleId": "com.facebook.WebDriverAgentRunner.xctrunner"})
dev.jobs.forward({"hostPort": 8100, "targetPort": 8100})
dev.jobs.list(); dev.jobs.get(job["id"]); dev.jobs.delete(job["id"])
for line in dev.jobs.logs(job["id"]):
    print(line.line)
```

## Streaming (SSE)

Each of the six long-lived endpoints is exposed as a generator that yields typed
event objects. Heartbeat keep-alive frames are filtered out by default (pass
`include_heartbeats=True` to receive them). Unknown event types are surfaced as
`UnknownEvent` rather than dropped, for forward compatibility.

### Async (recommended)

```python
import asyncio
from go_ios_sdk import AsyncIosClient

async def main():
    async with AsyncIosClient(base_url="http://localhost:60105", api_key="my-key") as client:
        dev = client.device("00008110-0000000000000000")

        async for event in dev.syslog():
            print(event.message)

        async for event in dev.notifications():
            print(event.bundle_id, event.state)

        async for event in dev.ostrace(pid=123, subsystem="com.apple.network"):
            print(event.process_name, event.message)

        async for event in dev.listen():
            print(event.event, event.udid)

        async for sample in dev.sysmontap():
            print(sample.total_load, sample.system_load, sample.user_load)

        async for line in dev.jobs.logs("job-id"):
            print(line.line)

asyncio.run(main())
```

Streams are also async context managers, and cancelling the consuming task (or
leaving the `async with` block) closes the underlying HTTP connection promptly:

```python
async with dev.syslog() as stream:
    async for event in stream:
        if "boot" in event.message:
            break     # closes the stream on exit
```

### Sync

```python
with IosClient(base_url="http://localhost:60105", api_key="my-key") as client:
    for event in client.device(udid).syslog():
        print(event.message)
```

### Event types

| Method            | Event object            | Key fields                                             |
| ----------------- | ----------------------- | ------------------------------------------------------ |
| `syslog()`        | `SyslogMessage`         | `message`, `timestamp`                                 |
| `notifications()` | `AppStateNotification`  | `bundle_id`, `state`, `timestamp`                      |
| `ostrace(...)`    | `OsTraceEntry`          | `message`, `pid`, `process_name`, `level`, `subsystem` |
| `listen()`        | `AttachDetachEvent`     | `event`, `device_id`, `udid`, `properties`             |
| `sysmontap()`     | `CpuUsageSample`        | `total_load`, `system_load`, `user_load`               |
| `jobs.logs(id)`   | `JobLogLine`            | `line`                                                 |
| (any)             | `Heartbeat`             | `raw` (only when `include_heartbeats=True`)            |
| (any)             | `UnknownEvent`          | `event`, `data`                                        |

Every event also carries a `.raw` dict with the original JSON payload.

## Errors

Non-2xx responses raise `go_ios_sdk.ApiError` (subclass of `GoIosError`) with
`.status_code`, `.message` and `.error` populated from the server's
`GenericResponse` body.

```python
from go_ios_sdk import ApiError

try:
    client.device("unknown").info()
except ApiError as e:
    print(e.status_code, e.message)   # e.g. 404 "device not found"
```

## Development

```bash
uv sync --extra dev      # or: pip install -e ".[dev]"
uv run pytest            # unit tests (SSE parser + facade, mocked transport)
uv run mypy src/go_ios_sdk
uv run ruff check src tests
```

The low-level typed client under `go_ios_sdk._generated` is generated from the
OpenAPI spec (`spec/openapi/openapi.yaml`, 80 operations) with
[openapi-python-client](https://github.com/openapi-generators/openapi-python-client)
`0.26.1` and vendored so the package is self-contained. The public facade in
`go_ios_sdk/` is hand-written and kept API-identical across the TypeScript, Java
and C# SDKs.
