"""Local go-ios REST daemon discovery.

The daemon (``restapi``) binds an ephemeral loopback port by default and writes a
discovery file at ``<home>/rest-api.json`` after a successful bind. SDKs read that
file's ``baseUrl`` to auto-connect without hardcoding a port.

Home resolution (identical across all go-ios SDKs):
    ``GO_IOS_HOME`` env (if set and non-empty), else ``~/.go-ios``.

See ``DISCOVERY-CONTRACT.md`` for the authoritative shape of the file.
"""

from __future__ import annotations

import json
import os
from pathlib import Path
from typing import Optional

from .errors import GoIosError

DISCOVERY_FILE = "rest-api.json"


class DiscoveryError(GoIosError):
    """Raised when the local go-ios REST daemon cannot be discovered."""


def go_ios_home() -> Path:
    """Return the go-ios home directory.

    ``GO_IOS_HOME`` env if set and non-empty; else ``~/.go-ios``.
    """
    env = os.environ.get("GO_IOS_HOME")
    if env:
        return Path(env)
    return Path.home() / ".go-ios"


def discovery_file_path() -> Path:
    """Return the path to the discovery file ``<home>/rest-api.json``."""
    return go_ios_home() / DISCOVERY_FILE


def discover_base_url() -> str:
    """Read the local daemon's ``baseUrl`` from the discovery file.

    Raises :class:`DiscoveryError` with a clear message if the file is missing,
    unreadable, malformed, or does not contain a usable ``baseUrl``.
    """
    path = discovery_file_path()
    try:
        raw = path.read_text(encoding="utf-8")
    except FileNotFoundError as exc:
        raise DiscoveryError(_not_found_message(path)) from exc
    except OSError as exc:
        raise DiscoveryError(
            f"{_not_found_message(path)} (could not read discovery file: {exc})"
        ) from exc

    try:
        data = json.loads(raw)
    except ValueError as exc:
        raise DiscoveryError(
            f"{_not_found_message(path)} (discovery file is not valid JSON)"
        ) from exc

    if not isinstance(data, dict):
        raise DiscoveryError(
            f"{_not_found_message(path)} (discovery file has an unexpected shape)"
        )

    base_url = data.get("baseUrl")
    if not isinstance(base_url, str) or not base_url:
        raise DiscoveryError(
            f"{_not_found_message(path)} (discovery file has no 'baseUrl')"
        )

    hint = _stale_pid_hint(data.get("pid"))
    if hint is not None:
        raise DiscoveryError(f"{_not_found_message(path)} ({hint})")
    return base_url


def _stale_pid_hint(pid: object) -> Optional[str]:
    """Return a hint if the recorded pid is present and clearly not running."""
    if not isinstance(pid, int) or pid <= 0:
        return None
    try:
        os.kill(pid, 0)
    except ProcessLookupError:
        return (
            f"discovery file records pid {pid}, which is not running; "
            "the daemon may have exited without cleaning up"
        )
    except (PermissionError, OSError):
        # Can't determine liveness (e.g. different user / unsupported) — assume ok.
        return None
    return None


def _not_found_message(path: Path) -> str:
    return (
        f"no local go-ios REST daemon found at {path}; "
        "start it (run the go-ios REST API) or pass an explicit base_url"
    )


def resolve_base_url(explicit: Optional[str]) -> str:
    """Resolve the daemon base URL.

    Order: explicit argument > ``GO_IOS_BASE_URL`` env > discovery file > error.
    """
    if explicit is not None and explicit != "":
        return explicit
    env = os.environ.get("GO_IOS_BASE_URL")
    if env:
        return env
    return discover_base_url()
