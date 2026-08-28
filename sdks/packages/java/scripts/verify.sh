#!/usr/bin/env bash
# Compile the go-ios Java SDK (committed generated client + hand-written facade)
# and run the JUnit test suite WITHOUT Maven, using javac and the JUnit Platform
# Console Standalone launcher.
#
# This mirrors what `mvn -q package` does, for environments where only a JDK 17+
# is available. Dependency jars and the JUnit launcher are downloaded once into
# .tools/lib/ (gitignored).
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LIB="${HERE}/.tools/lib"
M="https://repo1.maven.org/maven2"

JACKSON="2.18.2"
HTTPCORE="4.4.16"
HTTPCLIENT="4.5.14"
JUNIT_LAUNCHER="1.11.4"

deps=(
  "com/fasterxml/jackson/core/jackson-databind/${JACKSON}/jackson-databind-${JACKSON}.jar"
  "com/fasterxml/jackson/core/jackson-core/${JACKSON}/jackson-core-${JACKSON}.jar"
  "com/fasterxml/jackson/core/jackson-annotations/${JACKSON}/jackson-annotations-${JACKSON}.jar"
  "com/fasterxml/jackson/datatype/jackson-datatype-jsr310/${JACKSON}/jackson-datatype-jsr310-${JACKSON}.jar"
  "org/apache/httpcomponents/httpmime/${HTTPCLIENT}/httpmime-${HTTPCLIENT}.jar"
  "org/apache/httpcomponents/httpclient/${HTTPCLIENT}/httpclient-${HTTPCLIENT}.jar"
  "org/apache/httpcomponents/httpcore/${HTTPCORE}/httpcore-${HTTPCORE}.jar"
  "jakarta/annotation/jakarta.annotation-api/3.0.0/jakarta.annotation-api-3.0.0.jar"
  "org/junit/platform/junit-platform-console-standalone/${JUNIT_LAUNCHER}/junit-platform-console-standalone-${JUNIT_LAUNCHER}.jar"
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
find "${HERE}/generated/src/main/java" "${HERE}/src/main/java" -name '*.java' > "${HERE}/.tools/sources.txt"
javac --release 17 -cp "${CP}" -d "${HERE}/target/classes" @"${HERE}/.tools/sources.txt"

echo "Compiling tests..."
rm -rf "${HERE}/target/test-classes"
mkdir -p "${HERE}/target/test-classes"
find "${HERE}/src/test/java" -name '*.java' > "${HERE}/.tools/test-sources.txt"
javac --release 17 -cp "${CP}${HERE}/target/classes" -d "${HERE}/target/test-classes" \
  @"${HERE}/.tools/test-sources.txt"

echo "Running JUnit console..."
java -jar "${LIB}/junit-platform-console-standalone-${JUNIT_LAUNCHER}.jar" execute \
  -cp "${HERE}/target/classes:${HERE}/target/test-classes:${CP}" \
  --scan-classpath
