#!/usr/bin/env python3
"""02 — Device info.

Resolve a single device (``GO_IOS_UDID`` or the first attached one) and print its
lockdown + instruments info via ``device.info()``. This is the go-to call for
"what is this device?" — product type, iOS version, name, etc.

Run it::

    export GO_IOS_API_KEY=your-key
    # export GO_IOS_UDID=00008110-000...   # optional
    uv run python examples/02_device_info.py
"""

from __future__ import annotations

from _common import base_url, print_header, require_api_key, resolve_device

from go_ios_sdk import IosClient

# A short, stable set of keys worth surfacing from the (large) info dict. The full
# dict is printed below too, so nothing is hidden — this is just a readable summary.
_INTERESTING = [
    "DeviceName",
    "ProductType",
    "ProductVersion",
    "BuildVersion",
    "CPUArchitecture",
    "SerialNumber",
]


def main() -> None:
    print_header("02 device_info")
    api_key = require_api_key()

    with IosClient(base_url=base_url(), api_key=api_key) as client:
        udid, device = resolve_device(client)
        print(f"target udid: {udid}")

        info = device.info()

        print("summary:")
        for key in _INTERESTING:
            if key in info:
                print(f"  {key}: {info[key]}")

        print(f"\n(info() returned {len(info)} keys in total)")


if __name__ == "__main__":
    main()
