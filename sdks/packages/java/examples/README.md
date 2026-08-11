# go-ios Java SDK — examples

Runnable, heavily commented example programs for the go-ios Java SDK. Each class
has its own `main`, is configured entirely through environment variables, and
demonstrates exactly one feature so it can be read as documentation. Together
they also serve as a **pre-release smoke test** (see [`run.sh`](run.sh)):
[`RunAllExamples`](RunAllExamples.java) runs them in sequence and exits non-zero
if any core example throws.

## The examples

| # | Class | What it shows |
| - | ----- | ------------- |
| 1 | [`ListDevicesExample`](src/com/github/danielpaulus/goios/examples/ListDevicesExample.java) | Build an `IosClient`, list attached devices (`GET /list`). |
| 2 | [`DeviceInfoExample`](src/com/github/danielpaulus/goios/examples/DeviceInfoExample.java) | Read a device's info (`GET /device/{udid}/info`). |
| 3 | [`ListAppsExample`](src/com/github/danielpaulus/goios/examples/ListAppsExample.java) | List installed apps (`GET /device/{udid}/apps/`). |
| 4 | [`ScreenshotExample`](src/com/github/danielpaulus/goios/examples/ScreenshotExample.java) | Capture a PNG screenshot to `./screenshot.png`. |
| 5 | [`StreamSyslogExample`](src/com/github/danielpaulus/goios/examples/StreamSyslogExample.java) | Stream syslog over SSE (`SseReader`), stop after ~20 events or ~5 s. |
| 6 | [`UiAutomationExample`](src/com/github/danielpaulus/goios/examples/UiAutomationExample.java) | **Optional.** UI tap + type via WebDriverAgent; only when `RUN_UI=1`. |

Examples 2–6 need a device: when none is attached they print `SKIP` and return
normally, so the suite still passes on a device-less daemon.

## 1. Start the daemon

The examples talk to a running go-ios REST daemon. Start one with an API key (or
with `--disable-auth`, in which case any non-empty `GO_IOS_API_KEY` works):

```bash
# From the go-ios repo root; --api-key can be any secret you choose.
ios api --api-key "$GO_IOS_API_KEY"          # serves on http://localhost:8080
```

## 2. Configure the environment

| Variable | Required | Default | Meaning |
| -------- | -------- | ------- | ------- |
| `GO_IOS_API_KEY` | **yes** | — | Bearer token sent on every request. Missing → the example prints help and exits 1. |
| `GO_IOS_BASE_URL` | no | `http://localhost:8080` | Daemon origin (the SDK appends `/api/v1`). |
| `GO_IOS_UDID` | no | first attached device | Target device udid. |
| `RUN_UI` | no | unset | Set to `1` to also run the UI-automation example. |

```bash
export GO_IOS_API_KEY=dev
export GO_IOS_BASE_URL=http://localhost:8080   # optional
# export GO_IOS_UDID=00008110-0011...          # optional
```

## 3. Compile and run

No Maven required — [`run.sh`](run.sh) mirrors [`../scripts/verify.sh`](../scripts/verify.sh):
it downloads the dependency jars once into `../.tools/lib/` (gitignored),
compiles the SDK (committed generated client + hand-written facade) and the
examples with `javac --release 17`, then runs `RunAllExamples`.

```bash
# From sdks/packages/java/
bash examples/run.sh                 # compile + run all examples (the pre-release check)
RUN_UI=1 bash examples/run.sh        # also run the optional UI example
bash examples/run.sh --compile-only  # compile only; no daemon needed
```

Requires **JDK 17+**. `run.sh` exits non-zero if any core example (1–5) throws,
making it suitable as a CI / pre-release gate.

### Running a single example

After a compile (`bash examples/run.sh --compile-only`), run any one directly:

```bash
# The classpath is: examples classes, SDK classes, and the downloaded jars.
CP="examples/target/classes:target/classes:$(printf '%s:' .tools/lib/*.jar)"
java -cp "$CP" com.github.danielpaulus.goios.examples.ScreenshotExample
```
