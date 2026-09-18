#!/usr/bin/env bash
#
# Regenerate everything downstream from the TypeSpec source of truth.
#
# Today this compiles the spec (spec/*.tsp -> spec/openapi/*). As the per-language
# SDK generators land in Phase B, add their generate steps below (they all consume
# spec/openapi/openapi.yaml — the canonical OpenAPI 3.1 document).
#
# Targets: typescript, python, java, csharp, mcp.
#
# Usage: scripts/regen.sh
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SPEC_DIR="$REPO_ROOT/spec"
OPENAPI_DIR="$SPEC_DIR/openapi"

echo "==> Compiling TypeSpec -> OpenAPI (spec/)"
cd "$SPEC_DIR"
if [ ! -d node_modules ]; then
  echo "    installing spec dependencies (npm install)"
  npm install
fi
npx tsp compile .

# The emitter writes version-suffixed files (openapi.3.1.0.*, openapi.3.2.0.*).
# Publish the canonical OpenAPI 3.1 document under the stable name that every
# downstream generator hard-codes: spec/openapi/openapi.{yaml,json}.
echo "==> Publishing canonical 3.1 spec as openapi.{yaml,json}"
cp "$OPENAPI_DIR/openapi.3.1.0.yaml" "$OPENAPI_DIR/openapi.yaml"
cp "$OPENAPI_DIR/openapi.3.1.0.json" "$OPENAPI_DIR/openapi.json"

echo "==> OpenAPI documents:"
ls -1 "$OPENAPI_DIR"

# ---------------------------------------------------------------------------
# Phase B — SDK generation (placeholders; each phase fills its own block).
# All generators consume $OPENAPI_DIR/openapi.yaml.
# ---------------------------------------------------------------------------
# --- TypeScript (packages/typescript) : @hey-api/openapi-ts ---
#   (Phase B2)
# --- Python (packages/python)         : openapi-python-client ---
#   (Phase B3)
# --- Java (packages/java)             : openapi-generator java client ---
#   (Phase B, enterprise/Appium tier)
# --- C# (packages/csharp)             : openapi-generator csharp client ---
#   (Phase B, enterprise/Appium tier)
# --- MCP server (packages/mcp)        : OSS openapi->mcp, curated tools ---
#   (Phase B4)

echo "==> Done."
