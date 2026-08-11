# Changelog

All notable user-facing and internal changes to go-ios are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

Releases are cut by dispatching the **Release-Go-iOS** workflow (Actions tab →
Run workflow, or `gh workflow run release.yml -f release_notes=... -f bump=patch`).
That workflow computes the next version, writes a new `## [x.y.z] - DATE` section
here from the notes you provide, and uses the same text as the GitHub release
body — so this file is updated automatically and never needs a manual PR.
Use `## [Unreleased]` to jot down notable changes between releases if you like.

## [Unreleased]

## [1.3.1] - 2026-08-11

## Fixes

- **syslog:** `ios syslog` now decodes BSD vis escapes, so non-ASCII log text (e.g. CJK/Chinese) renders correctly instead of appearing as `\M-` escape sequences (#687).
- **image mount:** `ios image auto` resolves an already-downloaded developer disk image from the base directory before making any network request, so it works fully offline; DDI selection is now deterministic and cached image paths are contained within the base directory (#644).
- **fsync / house_arrest:** falls back to `VendDocuments` when `VendContainer` is denied, so `ios fsync --app=<bundleID>` now works on App Store apps that previously failed with `InstallationLookupFailed` (#593).
- **tunnel:** tunnel/RSD TCP dials are now bounded by a 15s timeout, so an unreachable device fails fast instead of hanging for ~135s (#764).
- **tunnel (Windows/macOS):** interface-setup errors now include the actual `netsh`/`ifconfig` output, the command that was run, and its exit status, plus an elevation hint and a suggestion to use `--userspace` (#545).
- **CLI dispatch:** `ios webinspector launch <url>` now runs the webinspector launcher instead of being misrouted to `ios launch`; related command-dispatch collisions were also hardened (#769).

## Improvements

- **help:** `ios help pcap` now includes practical usage examples, and the help catalog gained an `examples` field. Thanks to @sssppp-cooled for the contribution (#816).

## Chores

- Removed stray `proxy.mobileconfig` files from the repository root so they no longer leak into release assets (#731).

## [1.3.0] - 2026-08-10

## Security & stability
- **Two rounds of hardening against malformed/hostile input** (#791, #792): eliminated panics, unbounded allocations, and hangs across the untrusted-input surface — plist/usbmux framing, XPC, AFC, the DTX codec, testmanagerd, and NSKeyedArchiver (including an uncatchable stack overflow from cyclic archives).
- **REST API authentication (BREAKING).** The REST API no longer runs open by default: set `GO_IOS_API_KEY=<token>` (enforced as a bearer token on `/api/v1`) or pass `--disable-auth` to run without it. If neither is provided, the server refuses to start.
- REST API: rejected `basedir` path traversal, capped image uploads, and fixed a single-request server kill (`/notifications`) and a 100% CPU streaming busy-loop.
- Enabled TLS verification for the Apple TSS endpoint; contained AFC pull path traversal.
- Dependency/toolchain CVE fixes: go-pkcs12 → v0.7.2, toolchain → go1.26.5, x/crypto → v0.52.0, x/net → v0.55.0, and a quic-go CVE upgrade.

## Features
- **MDM** subcommands for passcode and Screen Time management, a read-only security-info query, and unlock-token fetch (#778).
- **Pasteboard** service control — get/set the device clipboard (#773).
- WDA/XCUITest now run on iOS 14–16 via an in-memory test config (#790).

## Fixes
- pair: report locked devices instead of false success; fix a PairSupervised panic when the device is passcode-locked.
- rsd: fail fast when a service is missing from the RSD list; export `RsdPortForService`; hint at outdated DDIs (#787).
- imagemounter: validate plist replies and the UnmountImage reply; fix silent personalized-image mount failures (#788).

## [1.2.1] - 2026-08-03

* fix(imagemounter): derive EPRO/ESEC from RestoreRequestRules for TSS [#781](https://github.com/danielpaulus/go-ios/pull/781)

## [1.2.0] - 2026-06-08

## Highlights

- **New `ios ui run (wda | devicekit)`** — bring up a WebDriverAgent or DeviceKit UI-automation runner and forward its port, the run counterpart to `ui download`/`ui install` (#761).
- **Accessibility audit** — `ios ax audit` runs the on-device accessibility audit on iOS 14–18, with structured output (#618).
- **WebInspector / CDP** — new WebInspector service with a Chrome DevTools Protocol bridge, interactive controls, and a JS shell (#744).
- **Apple signing + UI automation commands** — sign and drive WDA/DeviceKit runners directly from the CLI.
- **`runtest` / `runxctest` now emit test results as JSON** to stdout, always (#573).

## Tunnel (iOS 17+)

- Per-device tunnel stop and refresh (#738).
- Userspace tunnel: IPv6 framing fix and TLS-PSK transport for iOS 18.2+, plus robustness and performance work (#748).
- Limit automatic tunnel lookup to RSD commands (#753); attach tunnel info to `devicestate` / `resetlocation` / `setlocationgpx` (#756).
- Fix a usbmux socket leak in `TunnelManager` — skip network devices and back off failed ones (#682).
- More reliable single-device reconnect handling (#691).

## Fixes

- `forward`: stop logging normal connection teardown as errors (#754, #639) and make teardown race-free (#762).
- `instruments`: retry transient device launch failures (NSError code 2), fixing flaky `ios launch` (#763).
- Windows: stop the TUN event loop spinning when a device disconnects (#690).
- Replace a panic with an error return in the archive path (#705).
- Warn that `--pair-record-path=default` is TCC-blocked on macOS 26+ (#747).
- Better error messages explaining why instruments/devmode fail on unsupported devices (#742).

## Internal

- Go 1.26 across all modules; device command dispatch refactored into per-domain handlers; expanded real-device e2e coverage (pre-iOS17, WebInspector, signing, accessibility) with the data-race detector enabled in unit CI.

## 🙏 Thanks to our contributors

This release was made possible by a fantastic group of contributors — thank you all:

- **@sakhisheikh** (Sakhi Mansoor) — built out the accessibility APIs (toggle caption text, first/last element, AX queries) that underpin `ios ax audit` (#618). 🎉
- **@aluedeke** (Andreas Lüdeke) — sharp-eyed fix to stop `forward` logging normal connection teardown as errors, with a clean half-duplex refactor (#754).
- **@lizhizhuanshu** (Ponder) — caught and fixed a Windows goroutine that busy-spun forever when a device disconnects (#690). A year-old fix, finally landed — worth the wait!
- **@vbragaru** — diagnosed and fixed a real usbmux socket leak in `TunnelManager`, complete with overnight `lsof` evidence (#682). Excellent debugging.
- **@briankrznarich** (Brian Krznarich) — spotted spurious errors logged when a forward connection is cleanly closed (#639). Precise root-cause, patiently carried across go-ios's logrus→golog migration.
- **@dmdmdm-nz** — made `runtest`/`runxctest` emit machine-readable JSON results, so test output is finally pipeable (#573, #574).
- **@mvanhorn** (Matt Van Horn) — replaced a panic with a proper error return in the archive path, making the library safer to embed (#705).

…and **@danielpaulus** for the WebInspector/CDP bridge, `ios ui run`, the tunnel and signing work, and shepherding it all in. 🚀

Several of these PRs had been waiting a while and were rebased onto current `main` with the contributors' original authorship preserved. Thank you for your patience and your excellent work — go-ios is better because of you.

## [1.1.0] - 2026-06-04

## What's new

### Performance instrumentation (#365, @kissfu)
New `instruments`-backed services for live device telemetry: **network statistics**, **battery**, **`sysmontap`** system monitoring, and **FPS / OpenGL graphics** metrics. Useful for profiling app performance and resource usage directly from the CLI/library, with configurable network timeouts.

### Wallpaper & icon layout (#714, @aluedeke)
New `wallpaper` and `icon-layout` commands to read the device's home-screen wallpaper and springboard icon arrangement.

### Wi-Fi profile management (#692, @gbalduzzi)
Add and remove Wi-Fi connection profiles on a device. Inputs are validated up front, and the command clearly reports when the device must be **supervised** for the operation to succeed rather than failing obscurely.

### Device shutdown (#693, @gbalduzzi)
New support for shutting a device down programmatically.

### `ConnectionType` in device details (#698, @Harrilee)
`ios list` now reports each device's `ConnectionType` (**USB** or **Network**), so devices reachable over the network are distinguishable from USB-attached ones. Covered by the e2e device suite.

### Configurable tunnel-info API host (#737, @aluedeke)
The tunnel-info HTTP API's bind host is now configurable instead of being hard-coded, making it easier to run the tunnel daemon in containerized or remote setups.

### REST API: `resetaccessibility` endpoint (#637, @iSevenDays)
New endpoint to reset a device's iOS accessibility settings (including font size and related options) back to defaults.

## Improvements & fixes

### More robust usbmuxd socket resolution (#577, @Ylarod)
`GetUsbmuxdSocket` / `USBMUXD_SOCKET_ADDRESS` handling was rewritten: an explicit scheme (`unix://`, `tcp://`) is honored as-is (and case-insensitively), a bare `host:port` is treated as TCP, and a bare path is treated as a unix socket — without the previous panic on unscheme'd `host:port` values.

### Pluggable logging — logrus → slog (#736, @danielpaulus)
go-ios no longer depends on **logrus**. All library logging now flows through a thin `ios/golog` slog seam with consistent, filterable attributes (`module`, `udid`, and instance identifiers). Library embedders can route go-ios's logs into their own handler with `ios.SetLogger(*slog.Logger)`; if you do nothing, standard `slog.Default()` behavior applies.

> ⚠️ **Embedder-facing change:** the logrus dependency has been removed and `debugproxy`'s logger-typed signatures changed accordingly. CLI users are unaffected.

## Internal
- Removed legacy device integration tests now superseded by the gated `e2e` suite (#735, @danielpaulus)
- Added a contributor `AGENTS.md` documenting build/test, real-device CI, the logging convention, and the dispatch-only release process (#734, @danielpaulus)
- CI fixes for changelog insertion and npm-propagation wait during publish verification (#733, @danielpaulus)

Thanks to everyone who contributed to this release: @kissfu, @aluedeke, @gbalduzzi, @Harrilee, @iSevenDays, and @Ylarod. 🎉

## [1.0.218] - 2026-06-03

### ✨ New features & improvements
- **Redesigned help output** — `ios --help` now shows a clean, sectioned layout (global options + a full command table with one-line descriptions) instead of the raw docopt usage block. `ios help <command>`, `ios <command> --help`, and `ios <command> -h` are all equivalent, including nested subcommands like `ios help tunnel start`. (#716)

### 🐛 Fixes
- **REST API `KillApp`** — fixed kill-by-bundle-ID, which stopped matching after `CFBundleIdentifier`/`CFBundleExecutable` became method calls on the app model. (#729)

## [1.0.217] - 2026-06-03

First release published to npm since `1.0.213` — npm publishing broke when npm
revoked legacy tokens, and is now restored on secure OIDC trusted publishing.
This release also makes the npm package install correctly on Windows for the
first time.

### ✨ New features & improvements
- **`ostrace --follow`** — persistent log streaming that survives process restarts. Auto-reconnects when the target process exits/restarts: with `--process=<name>` it re-resolves the new PID, with an explicit `--pid` it exits cleanly when that PID ends. (#719)
- **`ostrace` plain-text output** — `--nojson` now prints structured, human-readable lines `[timestamp] PID:##### <Level> [subsystem:category] message` instead of raw JSON, with ANSI color coding (Info=cyan, Debug=gray, Error=red, Fault=bold red). Colors are emitted only when stdout is a TTY, so pipes and files stay clean. (#718)
- **More reliable Developer Disk Image (DDI) mounting on iOS 17+** — fixes `identity-not-found` failures on newer chips (e.g. A19 Pro) and adds a manifest fast-path that skips the nonce + Apple TSS round-trip on re-mounts when a valid personalization manifest already exists. (#720, #723)

### 🐛 Fixes
- **Windows npm install** — `npm i -g go-ios` previously installed a binary that wasn't on `PATH` (the postinstall placed it in a non-PATH `bin` subfolder), so `ios` was uncallable on Windows. The binary is now installed to the correct location. (#730)
- **macOS 26 binary launch** — bump the Go toolchain to 1.24.13 so produced binaries carry an `LC_UUID` load command; without it, macOS 26's dyld rejects them with `abort trap`. (#723)

### 🔧 Build, CI & internals
- **npm publishing migrated to OIDC trusted publishing** — no more long-lived `NODE_AUTH_TOKEN`; releases authenticate per-run via GitHub OIDC, with provenance attestations. (#727)
- **Cross-platform install verification** — release and canary pipelines now install the freshly published package on Windows, Linux and macOS and run the binary, asserting the version matches. (#728, #730)
- **Canary release pipeline** — manual-dispatch workflow publishing a throwaway `go-ios-canary` package to validate the full release flow without shipping. (#727)
- **Releases now gated behind the `make-release` label.** (#726)
- **Real-device e2e test suite** on the self-hosted macOS/Linux runners, gated to run only after unit tests pass, and runnable on fork PRs via a maintainer `/test-devices` comment. (#723, #725, #726)
