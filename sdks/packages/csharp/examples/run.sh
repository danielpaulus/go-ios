#!/usr/bin/env bash
#
# Runnable pre-release smoke test for the go-ios C#/.NET SDK.
#
# Runs examples 1-5 against a live go-ios daemon and exits non-zero if any of
# them throws. Steps that need a device (or a forwarded WDA) print SKIP without
# failing, so this is safe to run even with no device attached — though for a
# real smoke test you want a device connected.
#
# Configure via environment (see examples/README.md):
#   GO_IOS_BASE_URL   default http://localhost:8080
#   GO_IOS_API_KEY    required (unless the daemon runs with --disable-auth)
#   GO_IOS_UDID       optional; first device is used when unset
#   RUN_UI=1          also run the mutating ui-automation example
#
# `dotnet` must be on PATH.
set -euo pipefail

# Resolve this script's directory so it works from anywhere.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

exec dotnet run --project "${SCRIPT_DIR}/GoIos.Examples" -- run-all
