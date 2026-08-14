# @go-ios/sdk examples

Runnable, heavily-commented example scripts for the `@go-ios/sdk` TypeScript
client. They double as **documentation** (read them top-to-bottom to learn the
public API) and as a **pre-release smoke test**: `npm run examples` against a live
daemon + device is a single green/red check.

Every script uses only the real public API (`IosClient`, `client.device(udid)`,
`.apps`, `.ui`, the SSE iterators, binary streams) and is configured entirely
through environment variables — no code edits needed.

## The examples

| Script | What it shows | Needs a device? |
| --- | --- | --- |
| `01-list-devices.ts` | Construct a client, list devices, print them | no |
| `02-device-info.ts` | `device.info()` — name / iOS version / model | yes |
| `03-list-apps.ts` | `device.apps.list()` — installed apps | yes |
| `04-screenshot.ts` | `device.screenshot()` → save `./screenshot.png` | yes |
| `05-stream-syslog.ts` | SSE: `for await` syslog, stop after ~20 events / ~5s | yes |
| `06-ui-automation.ts` | `device.ui` tap + type (advanced, opt-in) | yes + WDA |

`_shared.ts` holds the small env/config helpers shared by all scripts.
`run-all.ts` is the suite runner.

## Configuration (environment variables)

| Variable | Meaning | Default |
| --- | --- | --- |
| `GO_IOS_BASE_URL` | Base URL of the go-ios REST daemon | auto-discovered (`~/.go-ios/rest-api.json`) |
| `GO_IOS_API_KEY` | Bearer API key (**required**) | — |
| `GO_IOS_UDID` | Target device udid (optional → first device) | first device |
| `RUN_UI` | Set to `1` to also run `06-ui-automation` | unset |
| `UI_WDA_URL` | Override the forwarded WDA URL for `06` | SDK default |

If `GO_IOS_API_KEY` is unset, the scripts print a helpful message and exit
non-zero (a smoke test that silently skips auth would be misleading).

## Start a go-ios REST daemon

The examples talk to the go-ios REST server over HTTP with bearer auth. By
default the daemon binds an **ephemeral loopback port** and writes a discovery
file at `~/.go-ios/rest-api.json`; the examples auto-discover it, so you do not
need to know or set the port. Start it with an API key:

```sh
# Terminal 1 — start the daemon with an API key (ephemeral loopback port)
GO_IOS_API_KEY=dev-secret ios --rest --api-key dev-secret
```

To pin a fixed port instead, start the daemon with `--addr :8080` and set
`GO_IOS_BASE_URL=http://localhost:8080` before running the examples.

(See `ios help` / the top-level README for the exact daemon flags on your
version — the key point is that the daemon is bound to an API key and the
examples send that same key as `Authorization: Bearer <key>`.)

## Run the whole suite (the smoke test)

```sh
# Terminal 2
export GO_IOS_API_KEY=dev-secret
# GO_IOS_BASE_URL is optional; unset, the local daemon is auto-discovered.
# export GO_IOS_BASE_URL=http://localhost:8080   # only to pin a fixed/remote daemon
# export GO_IOS_UDID=00008030-...                # only to pin a specific device

npm run examples
```

What the runner does:

- Runs `01`–`05` in order against the configured daemon, sharing one client.
- Prints a per-example `PASS` / `SKIP` / `FAIL` line and a final summary.
- **Exit code:** `0` if everything that ran passed or skipped; `1` if anything
  genuinely failed. So `npm run examples` is a clean green/red pre-release check.

### PASS vs SKIP vs FAIL

- **SKIP** — a device-dependent example that legitimately can't run here (no
  device attached, or the UI backend isn't forwarded). It is **not** a failure;
  the reachable surface (e.g. `01-list-devices`) is still verified. Running the
  suite against a daemon with no device yields all-green (`1 passed, 4 skipped`).
- **FAIL** — a real problem: auth failure, daemon down, an unexpected 5xx, or a
  bug in an example. Flips the exit code to `1`.

## Run a single example

Each script also runs on its own via [`tsx`](https://github.com/privatenumber/tsx)
(a dev dependency of this package):

```sh
GO_IOS_API_KEY=dev-secret npx tsx examples/01-list-devices.ts
GO_IOS_API_KEY=dev-secret npx tsx examples/04-screenshot.ts
```

Run directly, a script exits `0` on success (or a clean SKIP) and `1` on a real
error.

## The UI automation example (`06`)

`06-ui-automation.ts` needs a live UI backend — by default WebDriverAgent (WDA)
reachable over a forwarded port. It is opt-in (`RUN_UI=1`) and self-guarding: it
probes the backend first and SKIPs with a clear message if WDA isn't reachable.

To actually exercise it, start WDA and forward its port first (both are
long-running jobs on the daemon), e.g. from the SDK:

```ts
await client.device(udid).jobs.runwda();                                // start WDA
await client.device(udid).jobs.forward({ hostPort: 8100, targetPort: 8100 }); // forward
```

Then:

```sh
RUN_UI=1 GO_IOS_API_KEY=dev-secret npm run examples
# or standalone:
RUN_UI=1 GO_IOS_API_KEY=dev-secret npx tsx examples/06-ui-automation.ts
```

See the header comment in `06-ui-automation.ts` for the full walkthrough,
including how to override the forwarded WDA URL.

## Typecheck the examples

```sh
npm run examples:build   # tsc --noEmit over examples/ (no daemon needed)
```
