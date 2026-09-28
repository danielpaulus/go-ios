"""Discovery / ephemeral-port resolution tests.

Resolution order (contract): explicit base_url > GO_IOS_BASE_URL env >
discovery file (<home>/rest-api.json) > clear error.
"""

from __future__ import annotations

import json
import os
from pathlib import Path

import pytest

from go_ios_sdk import AsyncIosClient, DiscoveryError, IosClient
from go_ios_sdk.discovery import (
    discover_base_url,
    discovery_file_path,
    go_ios_home,
    resolve_base_url,
)

DISCOVERED_URL = "http://127.0.0.1:54321"


@pytest.fixture(autouse=True)
def _clean_env(monkeypatch: pytest.MonkeyPatch) -> None:
    """Ensure discovery-related env vars never leak in from the host."""
    monkeypatch.delenv("GO_IOS_HOME", raising=False)
    monkeypatch.delenv("GO_IOS_BASE_URL", raising=False)
    monkeypatch.delenv("GO_IOS_API_KEY", raising=False)


def _write_discovery(home: Path, base_url: str = DISCOVERED_URL, **extra: object) -> Path:
    home.mkdir(parents=True, exist_ok=True)
    payload: dict[str, object] = {
        "baseUrl": base_url,
        "host": "127.0.0.1",
        "port": 54321,
        "pid": os.getpid(),
        "startedAt": "2026-08-11T15:00:00Z",
        "tls": False,
    }
    payload.update(extra)
    path = home / "rest-api.json"
    path.write_text(json.dumps(payload), encoding="utf-8")
    return path


# -- home resolution --------------------------------------------------------


def test_home_defaults_to_dot_go_ios(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setattr(Path, "home", classmethod(lambda cls: Path("/fake/home")))
    assert go_ios_home() == Path("/fake/home/.go-ios")


def test_home_honors_go_ios_home_env(
    monkeypatch: pytest.MonkeyPatch, tmp_path: Path
) -> None:
    monkeypatch.setenv("GO_IOS_HOME", str(tmp_path))
    assert go_ios_home() == tmp_path
    assert discovery_file_path() == tmp_path / "rest-api.json"


def test_empty_go_ios_home_falls_back_to_default(
    monkeypatch: pytest.MonkeyPatch
) -> None:
    monkeypatch.setenv("GO_IOS_HOME", "")
    monkeypatch.setattr(Path, "home", classmethod(lambda cls: Path("/fake/home")))
    assert go_ios_home() == Path("/fake/home/.go-ios")


# -- discovery file reading -------------------------------------------------


def test_discover_reads_base_url(
    monkeypatch: pytest.MonkeyPatch, tmp_path: Path
) -> None:
    monkeypatch.setenv("GO_IOS_HOME", str(tmp_path))
    _write_discovery(tmp_path)
    assert discover_base_url() == DISCOVERED_URL


def test_missing_file_raises_clear_error(
    monkeypatch: pytest.MonkeyPatch, tmp_path: Path
) -> None:
    monkeypatch.setenv("GO_IOS_HOME", str(tmp_path))
    with pytest.raises(DiscoveryError) as ei:
        discover_base_url()
    msg = str(ei.value)
    assert "no local go-ios REST daemon found" in msg
    assert str(tmp_path / "rest-api.json") in msg
    assert "base_url" in msg


def test_malformed_json_raises_clear_error(
    monkeypatch: pytest.MonkeyPatch, tmp_path: Path
) -> None:
    monkeypatch.setenv("GO_IOS_HOME", str(tmp_path))
    tmp_path.mkdir(parents=True, exist_ok=True)
    (tmp_path / "rest-api.json").write_text("{ not json", encoding="utf-8")
    with pytest.raises(DiscoveryError) as ei:
        discover_base_url()
    assert "no local go-ios REST daemon found" in str(ei.value)


def test_missing_base_url_key_raises(
    monkeypatch: pytest.MonkeyPatch, tmp_path: Path
) -> None:
    monkeypatch.setenv("GO_IOS_HOME", str(tmp_path))
    tmp_path.mkdir(parents=True, exist_ok=True)
    (tmp_path / "rest-api.json").write_text(
        json.dumps({"host": "127.0.0.1", "port": 1}), encoding="utf-8"
    )
    with pytest.raises(DiscoveryError):
        discover_base_url()


def test_dead_pid_hint_in_error(
    monkeypatch: pytest.MonkeyPatch, tmp_path: Path
) -> None:
    monkeypatch.setenv("GO_IOS_HOME", str(tmp_path))
    # A pid extremely unlikely to be alive.
    _write_discovery(tmp_path, pid=2_000_000_000)
    with pytest.raises(DiscoveryError) as ei:
        discover_base_url()
    assert "not running" in str(ei.value)


# -- resolution order -------------------------------------------------------


def test_resolve_prefers_explicit(
    monkeypatch: pytest.MonkeyPatch, tmp_path: Path
) -> None:
    monkeypatch.setenv("GO_IOS_HOME", str(tmp_path))
    monkeypatch.setenv("GO_IOS_BASE_URL", "http://env:1")
    _write_discovery(tmp_path)
    assert resolve_base_url("http://explicit:2") == "http://explicit:2"


def test_resolve_env_over_discovery(
    monkeypatch: pytest.MonkeyPatch, tmp_path: Path
) -> None:
    monkeypatch.setenv("GO_IOS_HOME", str(tmp_path))
    monkeypatch.setenv("GO_IOS_BASE_URL", "http://env:1")
    _write_discovery(tmp_path)
    assert resolve_base_url(None) == "http://env:1"


def test_resolve_discovery_when_no_arg_no_env(
    monkeypatch: pytest.MonkeyPatch, tmp_path: Path
) -> None:
    monkeypatch.setenv("GO_IOS_HOME", str(tmp_path))
    _write_discovery(tmp_path)
    assert resolve_base_url(None) == DISCOVERED_URL


# -- client integration -----------------------------------------------------


def test_ios_client_uses_discovered_base_url(
    monkeypatch: pytest.MonkeyPatch, tmp_path: Path
) -> None:
    monkeypatch.setenv("GO_IOS_HOME", str(tmp_path))
    _write_discovery(tmp_path)
    with IosClient() as client:
        assert client.base_url == DISCOVERED_URL


def test_async_client_uses_discovered_base_url(
    monkeypatch: pytest.MonkeyPatch, tmp_path: Path
) -> None:
    monkeypatch.setenv("GO_IOS_HOME", str(tmp_path))
    _write_discovery(tmp_path)
    client = AsyncIosClient()
    assert client.base_url == DISCOVERED_URL


def test_explicit_base_url_overrides_discovery(
    monkeypatch: pytest.MonkeyPatch, tmp_path: Path
) -> None:
    monkeypatch.setenv("GO_IOS_HOME", str(tmp_path))
    _write_discovery(tmp_path)
    with IosClient(base_url="http://explicit:9000/") as client:
        assert client.base_url == "http://explicit:9000"


def test_env_base_url_used_when_no_arg(
    monkeypatch: pytest.MonkeyPatch, tmp_path: Path
) -> None:
    monkeypatch.setenv("GO_IOS_HOME", str(tmp_path))
    monkeypatch.setenv("GO_IOS_BASE_URL", "http://from-env:8080")
    # No discovery file on disk: env must still be used.
    with IosClient() as client:
        assert client.base_url == "http://from-env:8080"


def test_ios_client_missing_daemon_raises(
    monkeypatch: pytest.MonkeyPatch, tmp_path: Path
) -> None:
    monkeypatch.setenv("GO_IOS_HOME", str(tmp_path))
    with pytest.raises(DiscoveryError) as ei:
        IosClient()
    assert "no local go-ios REST daemon found" in str(ei.value)


def test_api_key_from_env(
    monkeypatch: pytest.MonkeyPatch, tmp_path: Path
) -> None:
    monkeypatch.setenv("GO_IOS_HOME", str(tmp_path))
    monkeypatch.setenv("GO_IOS_API_KEY", "env-tok")
    _write_discovery(tmp_path)
    with IosClient() as client:
        assert client.api_key == "env-tok"
