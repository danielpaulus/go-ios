#!/usr/bin/env bash
# Regenerate the low-level go-ios Java client from the canonical OpenAPI 3.1 spec.
#
# Pinned to openapi-generator-cli 7.11.0 (java generator, native HTTP library).
# Generated sources land in packages/java/generated/ and are committed.
set -euo pipefail

OPENAPI_GENERATOR_VERSION="7.11.0"
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
JAR="${HERE}/.tools/openapi-generator-cli.jar"
JAR_URL="https://repo1.maven.org/maven2/org/openapitools/openapi-generator-cli/${OPENAPI_GENERATOR_VERSION}/openapi-generator-cli-${OPENAPI_GENERATOR_VERSION}.jar"

if [[ ! -f "${JAR}" ]]; then
  echo "Downloading openapi-generator-cli ${OPENAPI_GENERATOR_VERSION}..."
  mkdir -p "${HERE}/.tools"
  curl -sSL -o "${JAR}" "${JAR_URL}"
fi

echo "Cleaning generated/ ..."
rm -rf "${HERE}/generated"

echo "Generating client (java / native, Java 17)..."
cd "${HERE}"
java -jar "${JAR}" generate -c "${HERE}/openapi-generator-config.yaml"

echo "Done. Generated sources under ${HERE}/generated/src/main/java"
