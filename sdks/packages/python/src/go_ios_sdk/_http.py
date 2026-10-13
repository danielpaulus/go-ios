"""Shared HTTP helpers for the facade (auth header, error mapping)."""

from __future__ import annotations

from typing import Any, Dict, Optional

import httpx

from .errors import ApiError

API_PREFIX = "/api/v1"


def build_headers(api_key: Optional[str], extra: Optional[Dict[str, str]] = None) -> Dict[str, str]:
    headers: Dict[str, str] = {}
    if api_key:
        headers["Authorization"] = f"Bearer {api_key}"
    if extra:
        headers.update(extra)
    return headers


def raise_for_status(response: httpx.Response) -> None:
    """Map a non-2xx response to :class:`ApiError`, parsing the GenericResponse body."""
    if response.is_success:
        return
    message: Optional[str] = None
    error: Optional[str] = None
    body: Optional[str] = None
    try:
        data = response.json()
        if isinstance(data, dict):
            message = data.get("message")
            error = data.get("error")
    except Exception:  # noqa: BLE001 - body may be non-JSON (e.g. binary/plain)
        try:
            body = response.text
        except Exception:  # noqa: BLE001
            body = None
    raise ApiError(response.status_code, message=message, error=error, body=body)


def json_or_none(response: httpx.Response) -> Any:
    if not response.content:
        return None
    try:
        return response.json()
    except Exception:  # noqa: BLE001
        return None
