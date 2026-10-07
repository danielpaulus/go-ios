"""Exceptions raised by the ergonomic go-ios SDK facade."""

from __future__ import annotations

from typing import Optional


class GoIosError(Exception):
    """Base class for all go-ios SDK errors."""


class ApiError(GoIosError):
    """A non-2xx HTTP response from the go-ios REST server.

    The server returns errors as a ``GenericResponse`` envelope
    (``{ message?, error? }``); the parsed fields are exposed here when present.
    """

    def __init__(
        self,
        status_code: int,
        message: Optional[str] = None,
        error: Optional[str] = None,
        *,
        body: Optional[str] = None,
    ) -> None:
        self.status_code = status_code
        self.message = message
        self.error = error
        self.body = body
        detail = message or error or body or ""
        super().__init__(f"HTTP {status_code}: {detail}".rstrip(": ").rstrip())
