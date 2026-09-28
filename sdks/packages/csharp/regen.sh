#!/usr/bin/env bash
# Regenerate the low-level C# client from the canonical OpenAPI 3.1 spec.
# Requires: Java (for openapi-generator) and npx.
#
# The hand-written facade under src/GoIos.Sdk/ is NOT touched.
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SPEC="$HERE/../../spec/openapi/openapi.yaml"
OUT="$HERE/src/Generated"
GEN_VERSION="7.14.0"

# npx can inherit a NODE_OPTIONS that breaks the CLI wrapper; clear it.
unset NODE_OPTIONS || true

cat > "$HERE/.openapi-generator-config.json" <<'JSON'
{
  "packageName": "GoIos.Sdk.Generated",
  "targetFramework": "net8.0",
  "library": "httpclient",
  "nullableReferenceTypes": true,
  "useDateTimeOffset": true,
  "netCoreProjectFile": true,
  "validatable": false,
  "hideGenerationTimestamp": true,
  "packageVersion": "0.1.0",
  "sourceFolder": "src"
}
JSON

npx --yes "@openapitools/openapi-generator-cli@2.20.0" version-manager set "$GEN_VERSION"
npx --yes "@openapitools/openapi-generator-cli@2.20.0" generate \
  -g csharp \
  -i "$SPEC" \
  -o "$OUT" \
  -c "$HERE/.openapi-generator-config.json" \
  --skip-validate-spec

# Post-fix: openapi-generator (7.14.0) emits a dangling base-class comma for
# anyOf union models whose only extra variant is an empty object (Heartbeat),
# producing "class X : AbstractOpenAPISchema, " which does not compile. Strip it.
find "$OUT/src" -name '*.cs' -print0 \
  | xargs -0 sed -i '' -E 's/(: AbstractOpenAPISchema),[[:space:]]*$/\1/'

# Drop generator scaffolding we do not commit (its own test project + solution).
rm -rf "$OUT/src/GoIos.Sdk.Generated.Test" "$OUT/GoIos.Sdk.Generated.sln" "$OUT/git_push.sh"

echo "Regenerated low-level client into $OUT"
