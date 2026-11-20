# go-ios-sdk examples

Runnable, heavily-commented example scripts for the Python SDK (`go-ios-sdk`).
They double as documentation and as a **pre-release smoke test** (`run_all.py`)
you can point at a real go-ios daemon with a device attached.

Each script is standalone — read one top to bottom to learn the API it covers.

## Prerequisites

1. **A running go-ios daemon.** Start it from a go-ios checkout / release:

   ```bash
   ios api --host 0.0.0.0 --port 8080 --api-key your-key
   ```

   (Or start it with `--disable-auth` for local experimentation; you still need
   to set `GO_IOS_API_KEY` to any value — see below.)

2. **An iOS device** attached to (or tunneled from) the daemon host. The
   info/apps/screenshot/stream examples need one; the runner records a `SKIP`
   (not a failure) when none is present.

## Configuration (environment variables)

| Variable          | Required | Default                 | Meaning                                            |
| ----------------- | -------- | ----------------------- | -------------------------------------------------- |
| `GO_IOS_API_KEY`  | **yes**  | —                       | Bearer token the daemon was started with           |
| `GO_IOS_BASE_URL` | no       | auto-discovered         | Daemon base URL; unset -> discover the local daemon (`~/.go-ios/rest-api.json`) |
| `GO_IOS_UDID`     | no       | first attached device   | udid (serial) of the device to target              |
| `RUN_UI`          | no       | unset                   | set to `1` to also run `07_ui_automation` in the runner |

Every example exits non-zero with a helpful message if `GO_IOS_API_KEY` is unset.

```bash
export GO_IOS_API_KEY=your-key
# GO_IOS_BASE_URL optional; unset, the local daemon is auto-discovered.
# Set it to target a pinned/remote daemon: export GO_IOS_BASE_URL=http://localhost:8080
export GO_IOS_UDID=00008110-000...             # optional; first device otherwise
```

## Run everything (the smoke test)

```bash
uv run python examples/run_all.py
```

Runs `01`–`06` in order against the daemon and prints a `PASS/SKIP/FAIL`
summary. It **exits non-zero** if any example raises an unexpected exception, and
treats "no device" / "UI backend unreachable" as `SKIP` (exit 0). Add the UI
example with:

```bash
RUN_UI=1 uv run python examples/run_all.py
```

## Run a single example

```bash
uv run python examples/01_list_devices.py
uv run python examples/04_screenshot.py     # writes ./screenshot.png
```

(If you have the package installed you can also run them with plain `python`;
`uv run` just guarantees the SDK and its deps are available.)

## The examples

| Script                  | Shows                                                            |
| ----------------------- | --------------------------------------------------------------- |
| `01_list_devices.py`    | Construct `IosClient`, list attached devices, print udids       |
| `02_device_info.py`     | `device.info()` for the target device                           |
| `03_list_apps.py`       | `device.apps.list()` installed apps                             |
| `04_screenshot.py`      | `device.screenshot()` → `./screenshot.png` (+ size / PNG check) |
| `05_stream_syslog.py`   | SSE: iterate `device.syslog()`, bounded to ~20 events / ~5s     |
| `06_async_stream.py`    | `AsyncIosClient` + `async for` over `sysmontap()`, bounded      |
| `07_ui_automation.py`   | *(optional)* `device.ui` tap + type over a forwarded WDA backend |

`07_ui_automation.py` needs a WebDriverAgent/DeviceKit backend up and forwarded
(bring it up with `device.jobs.runwda(...)` + `device.jobs.forward(...)`); it
skips gracefully if the backend is unreachable.

## Notes

* Streaming examples are deliberately **bounded** (by event count and time) so
  they always terminate — even on an idle device.
* `screenshot.png` is written into this directory when you run `04`.
