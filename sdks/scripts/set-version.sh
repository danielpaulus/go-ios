#!/usr/bin/env bash
# Stamp a single SDK version into every package manifest (lockstep versioning).
#
# All five SDKs are generated from one OpenAPI spec, so they share one version.
# This script writes $1 into each manifest with an anchored, in-place edit and
# fails loudly if any manifest was not actually updated — a partial version bump
# must never reach the publish step.
#
# Usage: sdks/scripts/set-version.sh <version>   e.g. set-version.sh 0.1.0
set -euo pipefail

VERSION="${1:?usage: set-version.sh <version> (e.g. 0.1.0)}"

# Reject anything that isn't a plain semver x.y.z(-prerelease)? — the value is
# interpolated into sed/perl patterns and manifests, so keep it strict.
if [[ ! "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+([-.+][0-9A-Za-z.-]+)?$ ]]; then
  echo "::error::invalid version '$VERSION' (expected semver like 0.1.0)" >&2
  exit 1
fi

# Resolve paths relative to this script so it works from any CWD.
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"  # -> sdks/
PKG="${HERE}/packages"

fail() { echo "::error::$*" >&2; exit 1; }

echo "Stamping SDK version ${VERSION} into all package manifests..."

# --- npm packages (typescript + mcp) --------------------------------------
# `npm version --no-git-tag-version` rewrites "version" robustly (keeps JSON
# formatting via npm's own writer). Both the public SDK and the (private) MCP
# server are versioned so the whole set stays in lockstep.
for np in typescript mcp; do
  ( cd "${PKG}/${np}" && npm version "${VERSION}" --no-git-tag-version --allow-same-version >/dev/null )
  grep -Eq "\"version\": \"${VERSION}\"" "${PKG}/${np}/package.json" \
    || fail "version not updated in packages/${np}/package.json"
  echo "  ok packages/${np}/package.json"
done

# --- Python (pyproject.toml) ----------------------------------------------
# Anchored to the [project] table's top-level `version = "..."` line.
PY="${PKG}/python/pyproject.toml"
perl -0pi -e 's/^version = "[^"]*"/version = "'"${VERSION}"'"/m' "${PY}"
grep -Eq "^version = \"${VERSION}\"" "${PY}" || fail "version not updated in ${PY}"
echo "  ok packages/python/pyproject.toml"

# --- Java (pom.xml) --------------------------------------------------------
# Only the project's own <version> (line ~9) sits immediately after
# <artifactId>go-ios-sdk</artifactId>. Anchor on that pair so dependency
# <version> tags are never touched. Maven Central rejects -SNAPSHOT releases, so
# write the bare version (no -SNAPSHOT suffix).
POM="${PKG}/java/pom.xml"
perl -0pi -e 's{(<artifactId>go-ios-sdk</artifactId>\s*<version>)[^<]*(</version>)}{${1}'"${VERSION}"'${2}}' "${POM}"
grep -Eq "<artifactId>go-ios-sdk</artifactId>[[:space:]]*<version>${VERSION}</version>" \
  <(tr '\n' ' ' < "${POM}") || fail "version not updated in ${POM}"
echo "  ok packages/java/pom.xml"

# --- C# (packable csproj) --------------------------------------------------
# Only GoIos.Sdk.csproj is IsPackable=true; stamp its <Version> element.
CSPROJ="${PKG}/csharp/src/GoIos.Sdk/GoIos.Sdk.csproj"
perl -0pi -e 's{<Version>[^<]*</Version>}{<Version>'"${VERSION}"'</Version>}' "${CSPROJ}"
grep -Eq "<Version>${VERSION}</Version>" "${CSPROJ}" || fail "version not updated in ${CSPROJ}"
echo "  ok packages/csharp/src/GoIos.Sdk/GoIos.Sdk.csproj"

echo "All manifests stamped to ${VERSION}."
