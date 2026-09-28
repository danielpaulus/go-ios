#!/usr/bin/env python3
"""07 — UI automation (OPTIONAL; needs a running WDA backend).

Drives the on-device UI through go-ios's ``/ui/*`` endpoints, which the daemon
proxies to a **WebDriverAgent / DeviceKit backend**. That backend must already be
running and reachable before this example can do anything:

    1. Start WDA as a server-side job:
         device.jobs.runwda({"bundleId":
             "com.facebook.WebDriverAgentRunner.xctrunner"})
    2. Forward its port so the daemon can reach it:
         device.jobs.forward({"hostPort": 8100, "targetPort": 8100})

Because that setup is device-, signing-, and provisioning-specific, this example
does **not** perform it for you — it assumes a backend is up and simply asks the
daemon for the screen size, taps, and types. If the UI backend is unreachable it
**skips gracefully** (prints ``SKIP`` and exits 0) instead of failing, so it is
safe to include in the smoke suite on machines without WDA configured.

Run it (only meaningful with a WDA backend up)::

    export GO_IOS_API_KEY=your-key
    RUN_UI=1 uv run python examples/07_ui_automation.py
"""

from __future__ import annotations

import httpx
from _common import SkipExample, base_url, print_header, require_api_key, resolve_device

from go_ios_sdk import ApiError, IosClient

# The UI group accepts a backend selector (``backend=``/``wda_url=``/``timeout=``).
# Adjust to match how your backend is exposed; "wda" is the common default.
_BACKEND = {"backend": "wda"}


def main() -> None:
    print_header("07 ui_automation")
    api_key = require_api_key()

    with IosClient(base_url=base_url(), api_key=api_key) as client:
        udid, device = resolve_device(client)
        print(f"target udid: {udid}")

        try:
            # First touch the backend cheaply; if WDA isn't up this is where it
            # will fail, and we convert that into a graceful SKIP.
            size = device.ui.size(**_BACKEND)
            width = int(size.get("width", 0)) or 200
            height = int(size.get("height", 0)) or 400
            print(f"screen size: {width}x{height}")

            # Tap roughly in the center, then type some text into whatever field
            # is focused. Both go through the forwarded WDA backend.
            cx, cy = width // 2, height // 2
            print(f"tapping at ({cx}, {cy})")
            device.ui.tap(cx, cy, **_BACKEND)

            print("typing 'hello from go-ios-sdk'")
            device.ui.type("hello from go-ios-sdk", **_BACKEND)

            print("UI automation calls completed")
        except (ApiError, httpx.HTTPError) as exc:
            # ApiError -> the daemon reached the backend and it returned non-2xx
            # (e.g. no session / backend not running). httpx.HTTPError -> we could
            # not even connect. Either way, there is no usable UI backend here.
            raise SkipExample(f"UI backend not reachable: {exc}") from exc


if __name__ == "__main__":
    main()
