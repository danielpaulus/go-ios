# GoIos.Sdk examples

Runnable, heavily-commented examples for the go-ios C#/.NET SDK. They double as
**documentation** (each example is standalone, annotated source) and as a
**pre-release smoke test** (`run.sh` exercises the read-only surface of a live
go-ios daemon and fails on any exception).

The project ([`GoIos.Examples`](./GoIos.Examples)) references the SDK facade
project directly, so it always compiles against the exact public API the package
ships.

## Prerequisites

- **.NET 8 SDK** (`dotnet` on your `PATH`).
- A running **go-ios REST daemon**. Start it and note the API key it prints:

  ```sh
  ios api --udid <udid>     # prints the base URL and the API key on startup
  ```

  (Or run it with `--disable-auth` for local experiments — see below.)
- For most examples, **an iOS device attached** to the daemon's host. Steps that
  need a device print `SKIP` instead of failing when none is available.
- For `ui-automation` only: a **forwarded WebDriverAgent** (`ios runwda`). Without
  it that example prints `SKIP`.

## Configuration

All configuration is via environment variables:

| Variable          | Default                  | Meaning                                                        |
| ----------------- | ------------------------ | -------------------------------------------------------------- |
| `GO_IOS_BASE_URL` | `http://localhost:8080`  | Base URL of the go-ios REST daemon.                            |
| `GO_IOS_API_KEY`  | *(required)*             | Bearer token the daemon prints on startup.                     |
| `GO_IOS_UDID`     | *(first device)*         | Target device udid. When unset, the first device is used.      |
| `RUN_UI`          | *(off)*                  | Set to `1` to also run the mutating `ui-automation` example.   |

If `GO_IOS_API_KEY` is missing, the examples print how to fix it and exit `2`.
If the daemon was started with `--disable-auth`, set `GO_IOS_API_KEY` to any
non-empty placeholder to acknowledge that.

## Running

```sh
export GO_IOS_BASE_URL=http://localhost:8080
export GO_IOS_API_KEY=<the-key-the-daemon-printed>
# export GO_IOS_UDID=00008110-000123456789ABCD   # optional

# Run examples 1-5 in sequence (the pre-release smoke test):
./run.sh
# equivalent to:
dotnet run --project GoIos.Examples -- run-all

# Run one example by name:
dotnet run --project GoIos.Examples -- list-devices
dotnet run --project GoIos.Examples -- device-info
dotnet run --project GoIos.Examples -- list-apps
dotnet run --project GoIos.Examples -- screenshot
dotnet run --project GoIos.Examples -- stream-syslog

# Include the mutating UI example in run-all:
RUN_UI=1 ./run.sh
# ...or run it directly:
dotnet run --project GoIos.Examples -- ui-automation
```

## Examples

| # | Command         | What it shows                                                                 |
| - | --------------- | ----------------------------------------------------------------------------- |
| 1 | `list-devices`  | Build `IosClient`, `Devices.ListAsync()`, print udid/connection per device.   |
| 2 | `device-info`   | `Device(udid).InfoAsync()` — lockdown + `instruments:*` values.               |
| 3 | `list-apps`     | `device.Apps.ListAsync()` — installed applications.                           |
| 4 | `screenshot`    | `device.ScreenshotAsync()` → `./screenshot.png`; prints the byte size.        |
| 5 | `stream-syslog` | `await foreach` over `device.SyslogAsync()` — ~20 events / ~5s, then stops.   |
| 6 | `ui-automation` | `device.Ui` tap + type (needs forwarded WDA; SKIPs if unreachable). Opt-in.   |

## Exit codes

| Code | Meaning                                                        |
| ---- | ------------------------------------------------------------- |
| `0`  | Every selected example succeeded or `SKIP`ped cleanly.        |
| `1`  | An example threw (hard failure).                              |
| `2`  | Configuration error (missing `GO_IOS_API_KEY`, unknown cmd). |

`SKIP` (no device attached, WDA not forwarded, …) is **not** a failure — the run
still exits `0`, which keeps `run.sh` usable as a smoke test even on a host with
no device currently attached.

## Building only

To verify the examples compile without a daemon:

```sh
dotnet build -c Release GoIos.Examples/GoIos.Examples.csproj
```
