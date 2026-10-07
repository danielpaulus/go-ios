"""Shared helpers for the go-ios-sdk examples.

Every example is a standalone script, but they all read the same three
environment variables and resolve a target device the same way — so that logic
lives here to keep each example short and focused on the one API it demonstrates.

Environment variables
----------------------
``GO_IOS_BASE_URL``
    Base URL of the running go-ios daemon. Optional — when unset the examples
    auto-discover the local daemon via ``~/.go-ios/rest-api.json``. Set it only to
    target a pinned or remote daemon.
``GO_IOS_API_KEY``
    Bearer token the daemon was started with. **Required** — every example
    exits with a helpful message and a non-zero status if it is unset. (A daemon
    started with ``--disable-auth`` accepts any value, so set it to anything,
    e.g. ``GO_IOS_API_KEY=none``, in that case.)
``GO_IOS_UDID``
    Optional. The udid (serial number) of the device to target. When unset the
    examples pick the **first** attached device.

The examples deliberately do not swallow API errors from the daemon: a raised
``go_ios_sdk.ApiError`` means the request reached the server and came back non-2xx,
which is useful signal for both a human reading the output and the ``run_all.py``
pre-release smoke test.
"""

from __future__ import annotations

import os
import sys
from typing import List, Optional, Tuple

# Make ``from go_ios_sdk import ...`` work when the examples are run straight from
# a checkout (``uv run python examples/01_list_devices.py``) without the package
# being installed: add the sibling ``src/`` directory to the import path. When the
# package *is* installed this simply has no effect (the installed copy wins).
_SRC = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "src")
if os.path.isdir(_SRC) and _SRC not in sys.path:
    sys.path.insert(0, _SRC)

from go_ios_sdk import Device, IosClient  # noqa: E402  (after the sys.path shim above)

#: Exit code used for a configuration/usage problem (e.g. no API key). Distinct
#: from ``1`` (which we let uncaught exceptions produce) purely for readability.
EXIT_CONFIG = 2


class SkipExample(Exception):
    """Raised by an example to signal it cannot run in this environment.

    The runner (``run_all.py``) treats this as a *SKIP*, not a failure — e.g. no
    device attached, or a UI backend that is not reachable. Running an example
    directly turns it into a printed ``SKIP:`` line and a zero exit code.
    """


def base_url() -> Optional[str]:
    """Return the configured daemon base URL, or ``None`` to auto-discover.

    Unset ``GO_IOS_BASE_URL`` -> ``None``, so the SDK falls through to local
    daemon discovery (``~/.go-ios/rest-api.json``).
    """
    return os.environ.get("GO_IOS_BASE_URL") or None


def require_api_key() -> str:
    """Return ``GO_IOS_API_KEY`` or print help and exit non-zero if it is unset.

    Every example calls this first so the failure mode for a missing key is a
    clear, actionable message rather than an opaque 401 from the server.
    """
    key = os.environ.get("GO_IOS_API_KEY")
    if not key:
        print(
            "ERROR: GO_IOS_API_KEY is not set.\n"
            "\n"
            "The go-ios daemon uses bearer auth. Export the key it was started\n"
            "with (or any value if it was started with --disable-auth):\n"
            "\n"
            "    export GO_IOS_API_KEY=your-key\n"
            "    # GO_IOS_BASE_URL is optional; unset, the local daemon is auto-discovered\n"
            "    export GO_IOS_UDID=00008110-000...           # optional; first device otherwise\n",
            file=sys.stderr,
        )
        raise SystemExit(EXIT_CONFIG)
    return key


def target_udid() -> Optional[str]:
    """Return the explicitly configured ``GO_IOS_UDID`` (or ``None`` to auto-pick)."""
    return os.environ.get("GO_IOS_UDID") or None


def list_udids(client: IosClient) -> List[str]:
    """Return the udids (serial numbers) of all attached devices."""
    return list(client.devices.udids())


def resolve_device(client: IosClient) -> Tuple[str, Device]:
    """Resolve a target device against a **sync** ``IosClient``.

    Returns ``(udid, device_handle)``. Uses ``GO_IOS_UDID`` when set, otherwise
    the first attached device. Raises :class:`SkipExample` when no device is
    attached (so the runner records a SKIP instead of failing the suite).
    """
    udid = target_udid()
    if udid is None:
        udids = list_udids(client)
        if not udids:
            raise SkipExample("no devices attached (set GO_IOS_UDID or attach a device)")
        udid = udids[0]
        print(f"(no GO_IOS_UDID set; using first device: {udid})")
    return udid, client.device(udid)


def print_header(title: str) -> None:
    """Print a small banner so multi-example runs are easy to read."""
    print(f"\n=== {title} ===")
    print(f"    daemon: {base_url() or '(auto-discovered)'}")
