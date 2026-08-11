# go-ios SDKs

Official SDKs and an MCP server for the [go-ios](https://github.com/danielpaulus/go-ios)
REST API, all generated from a single **spec-first source of truth**.

The API is authored in [TypeSpec](https://typespec.io) and emitted to **OpenAPI 3.1**.
Every SDK and the MCP server generate from that one OpenAPI document, so the spec —
not any individual client — is canonical. The go-ios REST server conforms to this
spec (it is the *ideal* contract; there is no backward-compatibility constraint yet).

## Targets

| Package               | Language / product | Generator                              | Status      |
| --------------------- | ------------------ | -------------------------------------- | ----------- |
| `packages/typescript` | TypeScript SDK     | `@hey-api/openapi-ts`                  | Phase B2    |
| `packages/python`     | Python SDK         | `openapi-python-client` (httpx)        | Phase B3    |
| `packages/java`       | Java SDK           | `openapi-generator` (java)             | Phase B     |
| `packages/csharp`     | C# SDK             | `openapi-generator` (csharp)           | Phase B     |
| `packages/mcp`        | MCP server         | OSS openapi→mcp, **curated** tools     | Phase B4    |

Five delivery targets: **typescript, python, java, csharp, mcp**.

## Layout

```
go-ios-sdks/
  spec/                      # TypeSpec source of truth (authoritative)
    main.tsp                 # service, auth, base path
    models.tsp               # data models + shared error responses
    routes.tsp               # all 26 operations
    streaming.tsp            # SSE event models + event unions
    tspconfig.yaml           # emits OpenAPI 3.1 (+ 3.2) to spec/openapi/
    openapi/                 # generated OpenAPI (committed)
      openapi.yaml           # canonical OpenAPI 3.1 (downstream generators read this)
      openapi.json           # canonical OpenAPI 3.1, JSON
      openapi.3.1.0.{yaml,json}   # version-suffixed emitter output
      openapi.3.2.0.{yaml,json}   # 3.2 variant — full typed SSE itemSchema
  packages/
    typescript/  python/  java/  csharp/  mcp/   # per-target (placeholders until Phase B)
  scripts/regen.sh           # regenerate everything from the spec
  docs/DESIGN.md             # locked decisions + SSE/streaming contract
  .github/workflows/         # CI: spec compile check
```

## Examples

Every SDK (and the MCP server) ships runnable, heavily-commented examples under
`packages/<lang>/examples/`. They double as tutorials and as the pre-release
smoke test. See [`docs/EXAMPLES.md`](docs/EXAMPLES.md) for the per-language
index, the shared `GO_IOS_BASE_URL` / `GO_IOS_API_KEY` / `GO_IOS_UDID` /
`RUN_UI` convention, and how to run them all.

## Regenerate everything

```bash
scripts/regen.sh
```

Or just the spec:

```bash
cd spec
npm install        # first time only
npx tsp compile .  # emits spec/openapi/openapi.3.1.0.{yaml,json} and 3.2.0
```

`regen.sh` then copies the canonical 3.1 output to the stable `spec/openapi/openapi.{yaml,json}`
path that all downstream generators consume.

## Authentication

All routes under `/api/v1` require `Authorization: Bearer <GO_IOS_API_KEY>`. The
server refuses to start without a key unless launched with `--disable-auth`, in
which case the header may be omitted. See `docs/DESIGN.md`.

## Streaming

`/notifications`, `/syslog`, `/ostrace`, and `/listen` are real Server-Sent Events
(`text/event-stream`) with typed event payloads. The contract and event model
shapes are documented in `docs/DESIGN.md`.
