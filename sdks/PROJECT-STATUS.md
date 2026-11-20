# go-ios SDKs + MCP — project status (2026-08-10)

Built autonomously. **Local git repo at `/Users/danielpaulus/privaterepos/go-ios-sdks` — 31 commits on `main`, NOT pushed (no remote created), nothing published.**

## What's here
- **spec/** — TypeSpec source of truth → OpenAPI 3.1 (`spec/openapi/openapi.yaml`). **80 operations / 65 paths**, 1:1 with the daemon's `/api/v1` surface (PR #817). Typed SSE via `@events` unions + `x-sse-events` extension (+ a 3.2 variant with inline itemSchema). API fixes baked in: `longtitude`→`longitude`, screenshot `image/png`, unified `GenericResponse` errors, real SSE.
- **packages/typescript** — `@go-ios/sdk` (hey-api client + facade), 38 tests. ESM+CJS via tsup, changesets, npm-provenance publish workflow.
- **packages/python** — `go-ios-sdk` (openapi-python-client + sync & async facades), 44 tests, mypy+ruff clean. uv + PyPI trusted-publishing (OIDC) workflow.
- **packages/java** — `com.github.danielpaulus:go-ios-sdk` (openapi-generator + facade), 32 tests (javac+JUnit; Maven Central pom). RawHttp for binary/multipart.
- **packages/csharp** — `GoIos.Sdk` (openapi-generator + async facade), 31 tests. NuGet publish workflow.
- **packages/mcp** — `@go-ios/mcp` (official MCP SDK), **31 curated tools**, stdio + Streamable-HTTP, 27 tests.
- **.github/workflows/ci.yml** — validates spec compile + every package build/test. Publishing is separate & gated (tag/dispatch), currently inert.

## Cross-language facade (identical shape)
`IosClient(baseUrl, apiKey)` → `devices.list()` · `device(udid).{info,deviceName,date,battery,diagnostics,mobileGestalt,processes,lockdown,screenshot,activate,pair,reboot,shutdown,erase,devmode,lang,memlimitoff,conditions/enable/disable,images/installImage/unmountImage,profiles/addProfile/removeProfile,resetAccessibility,resetLocation,setLocation}` + sub-clients `apps`, `wda`, `files`, `crashes`, `media`, `settings`, `mdm`, `proxy`, `jobs` (device-scoped) + SSE streams `syslog/notifications/ostrace/listen/sysmontap` · `client.tunnels` (fleet-level). SSE = async iterators (TS/C#), async generators (Python), Iterable/AutoCloseable (Java).

## Toolchain decisions (see docs/DESIGN.md, and scratchpad DECISION.md)
- Source of truth: **TypeSpec 1.0** (spec-first) → OpenAPI 3.1. OpenAPI is still the right IDL in 2026.
- Generators: **OSS-first** — hey-api (TS), openapi-python-client (Py), openapi-generator (Java/C#), official MCP SDK. Commercial rejected: Stainless (signups closed post-Anthropic-acquisition), Fern & Speakeasy (SSE + MCP paywalled / no 2-language free tier).
- Streaming: **SSE** for all log/event streams; WebSocket/mjpeg reserved for future interactive screen-mirror + input.
- MCP tools deliberately curated (not 1:1); omitted erase, raw file writes, MDM/system-config.

## Depends on
The SDKs/MCP target the **full ~80-endpoint daemon = PR #817** (`feature/restapi-parity`). #817 is merge-ready (see go-ios repo). Until it merges, the SDKs describe endpoints that live on that branch.

## Open items for Daniel
1. **Repo placement** — decide: new GitHub repo `go-ios-sdks`, or subtree in go-ios. Then push (I created no remote).
2. **Merge #817** — dismiss stale CodeQL alert #9 (`ncm/ncm.go`, pre-2024, unrelated) + re-run the macOS e2e syslog flake, then it's green.
3. **Publishing** — register `@go-ios` npm org, PyPI project, Maven Central namespace, NuGet id; wire OIDC/trusted-publishing; flip the (currently inert) publish workflows on.
4. Optional: push server-side to emit *real* SSE frames for the two legacy undelimited streams (/syslog, /listen) — spec already models it; small go-ios PR.
