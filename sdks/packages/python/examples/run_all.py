#!/usr/bin/env python3
"""Run every example in sequence — the pre-release smoke test.

Runs examples 01–06 against a live go-ios daemon, in order. Example 07 (UI
automation) only runs when ``RUN_UI=1`` because it needs a WebDriverAgent backend.

Exit behavior
-------------
* Any example raising an unexpected exception -> the runner prints ``FAIL`` and
  exits non-zero (so CI / a release gate catches it).
* An example that cannot run for environmental reasons (no device attached, UI
  backend unreachable) raises ``SkipExample`` -> printed as ``SKIP``; it does
  **not** fail the suite.
* Missing ``GO_IOS_API_KEY`` is reported once up front with a non-zero exit.

Run it::

    export GO_IOS_API_KEY=your-key
    # export GO_IOS_UDID=00008110-000...   # optional; first device otherwise
    uv run python examples/run_all.py
    RUN_UI=1 uv run python examples/run_all.py   # also run 07
"""

from __future__ import annotations

import importlib
import os
import sys
import traceback
from typing import Callable, List, Tuple

# Ensure this directory is importable whether run as ``python examples/run_all.py``
# or via the installed ``go-ios-examples`` console script (whose cwd is arbitrary).
_HERE = os.path.dirname(os.path.abspath(__file__))
if _HERE not in sys.path:
    sys.path.insert(0, _HERE)

from _common import SkipExample, base_url, require_api_key  # noqa: E402

# (module name, human label). Modules are imported lazily so an import error in
# one example is reported like any other failure rather than aborting the run.
_CORE: List[Tuple[str, str]] = [
    ("01_list_devices", "list devices"),
    ("02_device_info", "device info"),
    ("03_list_apps", "list apps"),
    ("04_screenshot", "screenshot"),
    ("05_stream_syslog", "stream syslog"),
    ("06_async_stream", "async stream"),
]
_UI: Tuple[str, str] = ("07_ui_automation", "ui automation")


def _load_main(module_name: str) -> Callable[[], None]:
    """Import an example module and return its ``main`` callable."""
    module = importlib.import_module(module_name)
    return module.main  # type: ignore[no-any-return]


def _run_one(module_name: str, label: str) -> str:
    """Run a single example. Returns one of ``PASS`` / ``SKIP`` / ``FAIL``."""
    print(f"\n########## {module_name} — {label} ##########")
    try:
        _load_main(module_name)()
    except SkipExample as skip:
        print(f"SKIP: {module_name}: {skip}")
        return "SKIP"
    except Exception:  # noqa: BLE001 — the runner is a top-level harness
        print(f"FAIL: {module_name}:")
        traceback.print_exc()
        return "FAIL"
    print(f"PASS: {module_name}")
    return "PASS"


def main() -> int:
    # Fail fast and clearly if the daemon key is missing (one message, not six).
    require_api_key()
    print(f"go-ios-sdk examples smoke test against {base_url() or '(auto-discovered daemon)'}")

    plan = list(_CORE)
    if os.environ.get("RUN_UI") == "1":
        plan.append(_UI)
    else:
        print("(RUN_UI!=1; skipping 07_ui_automation — set RUN_UI=1 to include it)")

    results = [(name, _run_one(name, label)) for name, label in plan]

    print("\n========== summary ==========")
    for name, status in results:
        print(f"  {status:4s}  {name}")

    failed = [name for name, status in results if status == "FAIL"]
    if failed:
        print(f"\n{len(failed)} example(s) FAILED: {', '.join(failed)}")
        return 1
    print("\nall examples passed (or skipped gracefully)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
