#!/usr/bin/env python3
"""01 — List attached devices.

The "hello world" of the SDK: construct an :class:`~go_ios_sdk.IosClient`, ask the
daemon which devices are attached, and print each one's udid (serial number) and
a couple of identifying properties.

``client.devices.list()`` returns the raw ``DeviceList`` envelope from the daemon
(a dict with a ``deviceList`` array). ``client.devices.udids()`` is a convenience
that pulls just the serial numbers out of that envelope.

Run it::

    export GO_IOS_API_KEY=your-key
    uv run python examples/01_list_devices.py
"""

from __future__ import annotations

from _common import base_url, print_header, require_api_key

from go_ios_sdk import IosClient


def main() -> None:
    print_header("01 list_devices")
    api_key = require_api_key()

    # The client is a context manager; leaving the ``with`` closes the underlying
    # httpx connection pool. base_url + api_key come from the environment.
    with IosClient(base_url=base_url(), api_key=api_key) as client:
        envelope = client.devices.list()
        device_list = envelope.get("deviceList", []) or []

        print(f"attached devices: {len(device_list)}")
        for entry in device_list:
            props = entry.get("properties", {}) or {}
            udid = props.get("serialNumber", "<unknown>")
            name = props.get("DeviceName", props.get("ProductType", ""))
            version = props.get("ProductVersion", "")
            print(f"  - {udid}  {name} {version}".rstrip())

        # udids() is the shortcut you'll reach for most often.
        print(f"udids(): {client.devices.udids()}")


if __name__ == "__main__":
    main()
