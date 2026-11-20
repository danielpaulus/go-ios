#!/usr/bin/env bash
# Compile and run the go-ios Java SDK examples WITHOUT Maven, using javac and the
# same dependency classpath as scripts/verify.sh. This is the pre-release smoke
# test: it builds the SDK (committed generated client + hand-written facade), then
# compiles and runs the examples' RunAllExamples driver against a live daemon.
#
# Requirements: JDK 17+, and a running go-ios daemon reachable at $GO_IOS_BASE_URL
# (default http://localhost:8080) with $GO_IOS_API_KEY set.
#
# Usage:
#   export GO_IOS_API_KEY=...          # required
#   export GO_IOS_BASE_URL=...         # optional (default http://localhost:8080)
#   export GO_IOS_UDID=...             # optional (default: first attached device)
#   export RUN_UI=1                    # optional: also run the UI-automation example
#   bash examples/run.sh
#
# Pass --compile-only to build the examples without launching them (used by CI /
# scripts/verify.sh-style compile checks where no daemon is available).
set -euo pipefail

COMPILE_ONLY=0
if [[ "${1:-}" == "--compile-only" ]]; then
  COMPILE_ONLY=1
fi

# Package root (sdks/packages/java): parent of this examples/ directory.
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
EX="${HERE}/examples"
LIB="${HERE}/.tools/lib"
M="https://repo1.maven.org/maven2"

# Keep these versions in lock-step with scripts/verify.sh.
JACKSON="2.18.2"
HTTPCORE="4.4.16"
HTTPCLIENT="4.5.14"

# Runtime dependencies of the SDK facade + generated client (no JUnit needed here).
deps=(
  "com/fasterxml/jackson/core/jackson-databind/${JACKSON}/jackson-databind-${JACKSON}.jar"
  "com/fasterxml/jackson/core/jackson-core/${JACKSON}/jackson-core-${JACKSON}.jar"
  "com/fasterxml/jackson/core/jackson-annotations/${JACKSON}/jackson-annotations-${JACKSON}.jar"
  "com/fasterxml/jackson/datatype/jackson-datatype-jsr310/${JACKSON}/jackson-datatype-jsr310-${JACKSON}.jar"
  "org/apache/httpcomponents/httpmime/${HTTPCLIENT}/httpmime-${HTTPCLIENT}.jar"
  "org/apache/httpcomponents/httpclient/${HTTPCLIENT}/httpclient-${HTTPCLIENT}.jar"
  "org/apache/httpcomponents/httpcore/${HTTPCORE}/httpcore-${HTTPCORE}.jar"
  "jakarta/annotation/jakarta.annotation-api/3.0.0/jakarta.annotation-api-3.0.0.jar"
)

mkdir -p "${LIB}"
for d in "${deps[@]}"; do
  f="${LIB}/$(basename "$d")"
  if [[ ! -f "$f" ]]; then
    echo "Downloading $(basename "$d")..."
    curl -sSL -o "$f" "${M}/${d}"
  fi
done

CP="$(printf '%s:' "${LIB}"/*.jar)"

echo "Compiling generated client + facade (javac --release 17)..."
rm -rf "${HERE}/target/classes"
mkdir -p "${HERE}/target/classes"
mkdir -p "${HERE}/.tools"
find "${HERE}/generated/src/main/java" "${HERE}/src/main/java" -name '*.java' \
  > "${HERE}/.tools/sources.txt"
javac --release 17 -cp "${CP}" -d "${HERE}/target/classes" @"${HERE}/.tools/sources.txt"

echo "Compiling examples (javac --release 17)..."
rm -rf "${EX}/target/classes"
mkdir -p "${EX}/target/classes"
# RunAllExamples.java lives at the examples/ root; the example classes live under
# examples/src/. Compile both trees against the SDK classes we just built.
find "${EX}/src" -name '*.java' > "${HERE}/.tools/example-sources.txt"
echo "${EX}/RunAllExamples.java" >> "${HERE}/.tools/example-sources.txt"
javac --release 17 -cp "${CP}${HERE}/target/classes" -d "${EX}/target/classes" \
  @"${HERE}/.tools/example-sources.txt"

if [[ "${COMPILE_ONLY}" == "1" ]]; then
  echo "Compile-only: examples compiled successfully."
  exit 0
fi

echo "Running RunAllExamples..."
java -cp "${EX}/target/classes:${HERE}/target/classes:${CP}" RunAllExamples
