#!/usr/bin/env python3
"""04 — Capture a screenshot.

``device.screenshot()`` returns the raw PNG bytes of a full-screen capture. Here
we write them to ``./screenshot.png`` next to the examples and print the file
size, so you have proof the round-trip worked end to end.

Run it::

    export GO_IOS_API_KEY=your-key
    uv run python examples/04_screenshot.py
"""

from __future__ import annotations

import os

from _common import base_url, print_header, require_api_key, resolve_device

from go_ios_sdk import IosClient

# Write the PNG next to this script regardless of the current working directory.
_OUT_PATH = os.path.join(os.path.dirname(os.path.abspath(__file__)), "screenshot.png")


def main() -> None:
    print_header("04 screenshot")
    api_key = require_api_key()

    with IosClient(base_url=base_url(), api_key=api_key) as client:
        udid, device = resolve_device(client)
        print(f"target udid: {udid}")

        png: bytes = device.screenshot()

        with open(_OUT_PATH, "wb") as f:
            f.write(png)

        # A well-formed PNG starts with this 8-byte signature — a cheap sanity check.
        looks_like_png = png[:8] == b"\x89PNG\r\n\x1a\n"
        print(f"wrote {len(png)} bytes to {_OUT_PATH}")
        print(f"looks like a PNG: {looks_like_png}")


if __name__ == "__main__":
    main()
