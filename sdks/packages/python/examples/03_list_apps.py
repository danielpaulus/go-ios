#!/usr/bin/env python3
"""03 — List installed apps.

``device.apps.list()`` returns the installed applications as a list of dicts
(bundle id, display name, version, and more). This example prints the first
handful so the output stays readable even on a device with hundreds of apps.

Run it::

    export GO_IOS_API_KEY=your-key
    uv run python examples/03_list_apps.py
"""

from __future__ import annotations

from _common import base_url, print_header, require_api_key, resolve_device

from go_ios_sdk import IosClient

# Cap how many apps we print — device app lists can be long.
_MAX_SHOWN = 15


def main() -> None:
    print_header("03 list_apps")
    api_key = require_api_key()

    with IosClient(base_url=base_url(), api_key=api_key) as client:
        udid, device = resolve_device(client)
        print(f"target udid: {udid}")

        apps = device.apps.list()
        print(f"installed apps: {len(apps)} (showing up to {_MAX_SHOWN})")

        for app in apps[:_MAX_SHOWN]:
            # Different iOS versions use slightly different key casing; fall back
            # gracefully so the example is robust across devices.
            bundle_id = app.get("CFBundleIdentifier") or app.get("bundleId") or "<no-id>"
            name = (
                app.get("CFBundleDisplayName")
                or app.get("CFBundleName")
                or app.get("name")
                or ""
            )
            version = app.get("CFBundleShortVersionString") or app.get("version") or ""
            print(f"  - {bundle_id}  {name} {version}".rstrip())


if __name__ == "__main__":
    main()
