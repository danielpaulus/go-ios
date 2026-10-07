# Releasing the go-ios SDKs

All five SDKs (`typescript`, `python`, `java`, `csharp`, `mcp`) are generated
from **one** OpenAPI spec (`sdks/spec/openapi/openapi.yaml`), so they ship in
**lockstep**: a single version number is stamped into every package manifest and
published to every registry in one run.

Releasing is **dispatch-only** — exactly like the CLI's `release.yml`. Merging a
PR never publishes anything. The pipeline lives in
[`.github/workflows/release-sdks.yml`](../../.github/workflows/release-sdks.yml)
and is completely separate from the CLI release (`release.yml`).

## Lockstep versioning

One `version` input drives all five packages. `sdks/scripts/set-version.sh`
stamps it into:

| Package    | Manifest                                             | Registry id                       |
| ---------- | ---------------------------------------------------- | --------------------------------- |
| typescript | `sdks/packages/typescript/package.json`              | npm `@go-ios/sdk`                  |
| mcp        | `sdks/packages/mcp/package.json`                     | (private — not published to npm)   |
| python     | `sdks/packages/python/pyproject.toml`                | PyPI `go-ios-sdk`                  |
| java       | `sdks/packages/java/pom.xml`                          | Maven Central `com.github.danielpaulus:go-ios-sdk` |
| csharp     | `sdks/packages/csharp/src/GoIos.Sdk/GoIos.Sdk.csproj` | NuGet `GoIos.Sdk`                 |

The script fails loudly if any manifest is not updated, so a partial version
bump can never reach the publish step. `packages/mcp` is `"private": true`; it is
versioned and built/tested for parity but not published to npm.

## How to cut a release

1. **Dry run first (always).** From the Actions tab run **Release-SDKs** with
   `dry_run = true` (the default), or:

   ```
   gh workflow run release-sdks.yml -f version=0.1.0 -f dry_run=true
   ```

   This stamps the version, builds and tests all five SDKs, then does a **real
   dry-run of every publish** — `npm publish --dry-run`, `twine check`,
   `mvn verify`, `dotnet pack` — with **no upload** and **no git tag/release**.

2. **Real release.** Once the dry run is green, run it again with
   `dry_run = false`:

   ```
   gh workflow run release-sdks.yml -f version=0.1.0 -f dry_run=false
   ```

   This rebuilds/retests, then each ecosystem's publish job uploads **only if
   that registry is armed** (see prerequisites below). After all publish jobs
   succeed it creates the git tag `sdk-v<version>` and a GitHub release.

### Safety model

- **Dispatch only.** No tag/push trigger — nothing here runs off a merge.
  `dry_run` defaults to `true`.
- **Build before upload.** All five build/test jobs must pass before any publish
  job starts, so a bad build ships nothing.
- **Every upload is gated twice:** (1) `if: ${{ !inputs.dry_run }}` — a dry run
  does a real dry-run instead; (2) a registry-armed guard — even a real run
  **self-skips** an upload (with a `::warning::` in the log) when that registry
  isn't configured yet. This makes the first real run safe before any registry
  exists.
- **Tag + GitHub release** are created only after every publish job succeeds and
  it was not a dry run — the single repo-mutating stage, last.

## Registry prerequisites (maintainer must set these up before real publishing)

Until each of these is configured, that ecosystem's real publish **self-skips**
with a warning; the run still succeeds. Arm them one at a time and re-run.

### npm (`@go-ios/sdk`)

- Create the **`@go-ios` npm org** and the `@go-ios/sdk` package.
- Register **OIDC trusted publishing** for the package on npmjs.com, pointing at
  this repo + the `release-sdks.yml` workflow.
- **No token.** Auth is OIDC only (`id-token: write`, `NPM_CONFIG_PROVENANCE`).
  Do **not** add `NODE_AUTH_TOKEN` or an `.npmrc` token line — even an empty one
  breaks OIDC (per AGENTS.md). The job self-skips until an OIDC token is issued.

### PyPI (`go-ios-sdk`)

- Create the **PyPI project** `go-ios-sdk`.
- Add a **Trusted Publisher** for this repo + `release-sdks.yml` (environment
  `pypi`).
- Set the repo secret **`PYPI_TRUSTED_PUBLISHER_CONFIGURED`** to any non-empty
  value — that's the arm flag the workflow gates on. (Publishing itself is
  tokenless OIDC via `pypa/gh-action-pypi-publish`.)

### Maven Central (`com.github.danielpaulus:go-ios-sdk`)

- Register the namespace on the **Sonatype Central Portal** — either
  `com.github.danielpaulus` (verify via GitHub) or an `io.github.*` namespace.
- Generate a **GPG key** and publish the public key to a keyserver.
- Add these repo secrets:
  - **`MAVEN_GPG_PRIVATE_KEY`** — ASCII-armored private key.
  - **`MAVEN_GPG_PASSPHRASE`** — its passphrase.
  - **`CENTRAL_TOKEN_USERNAME`** / **`CENTRAL_TOKEN_PASSWORD`** — Central Portal
    user token.
- The workflow deploys via the `release` profile in `pom.xml`
  (`central-publishing-maven-plugin` + `maven-gpg-plugin`). It self-skips unless
  all three of the key + token username + token password secrets are present.

### NuGet (`GoIos.Sdk`)

- Reserve the **`GoIos.Sdk`** package id on nuget.org.
- Create an API key scoped to that package and add it as the repo secret
  **`NUGET_API_KEY`**. The push self-skips until it exists.

## Notes

- The `sdk-v*` tag prefix is used so SDK releases never collide with the CLI's
  own `v*` release tags.
- To change what's published, edit `release-sdks.yml`; to change versions, pass a
  different `version` input — never hand-edit the version in the manifests for a
  release (the pipeline owns that via `set-version.sh`).
