"""Typed Server-Sent Event payloads for the go-ios streaming endpoints.

The wire contract is documented in ``docs/DESIGN.md``. Each SSE frame carries an
``event:`` name and a compact-JSON ``data:`` payload. The ``event:`` name selects
which of the dataclasses below the payload is decoded into.

Unknown events (event names not in the spec) are surfaced as :class:`UnknownEvent`
rather than dropped, so forward-compatible servers do not silently lose data.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any, Dict, Optional


@dataclass(frozen=True)
class Heartbeat:
    """Idle keep-alive frame (``event: heartbeat``, payload ``{}``).

    Emitted on every stream on an idle interval. The high-level async/sync
    generators skip heartbeats by default; pass ``include_heartbeats=True`` to
    receive them.
    """

    #: The (typically empty) raw JSON payload of the heartbeat frame.
    raw: Dict[str, Any] = field(default_factory=dict)


@dataclass(frozen=True)
class AppStateNotification:
    """``event: appstate`` on ``GET /device/{udid}/notifications``.

    ``state`` is one of ``foreground``, ``background``, ``suspended``,
    ``terminated`` or ``unknown``.
    """

    bundle_id: str
    state: str
    timestamp: Optional[int] = None
    raw: Dict[str, Any] = field(default_factory=dict, repr=False)

    @classmethod
    def from_data(cls, data: Dict[str, Any]) -> "AppStateNotification":
        return cls(
            bundle_id=data.get("bundleId", ""),
            state=data.get("state", ""),
            timestamp=data.get("timestamp"),
            raw=data,
        )


@dataclass(frozen=True)
class SyslogMessage:
    """``event: syslog`` on ``GET /device/{udid}/syslog``."""

    message: str
    timestamp: Optional[int] = None
    raw: Dict[str, Any] = field(default_factory=dict, repr=False)

    @classmethod
    def from_data(cls, data: Dict[str, Any]) -> "SyslogMessage":
        return cls(
            message=data.get("message", ""),
            timestamp=data.get("timestamp"),
            raw=data,
        )


@dataclass(frozen=True)
class OsTraceEntry:
    """``event: ostrace`` on ``GET /device/{udid}/ostrace``.

    ``level`` is one of ``default``, ``info``, ``debug``, ``error`` or ``fault``.
    """

    message: str
    pid: Optional[int] = None
    process_name: Optional[str] = None
    level: Optional[str] = None
    subsystem: Optional[str] = None
    category: Optional[str] = None
    timestamp: Optional[int] = None
    raw: Dict[str, Any] = field(default_factory=dict, repr=False)

    @classmethod
    def from_data(cls, data: Dict[str, Any]) -> "OsTraceEntry":
        return cls(
            message=data.get("message", ""),
            pid=data.get("pid"),
            process_name=data.get("processName"),
            level=data.get("level"),
            subsystem=data.get("subsystem"),
            category=data.get("category"),
            timestamp=data.get("timestamp"),
            raw=data,
        )


@dataclass(frozen=True)
class AttachDetachEvent:
    """``event: attachdetach`` on ``GET /device/{udid}/listen``.

    ``event`` is one of ``attached``, ``detached`` or ``paired``. ``properties``
    is present on ``attached`` frames.
    """

    event: str
    device_id: Optional[int] = None
    udid: Optional[str] = None
    properties: Optional[Dict[str, Any]] = None
    raw: Dict[str, Any] = field(default_factory=dict, repr=False)

    @classmethod
    def from_data(cls, data: Dict[str, Any]) -> "AttachDetachEvent":
        return cls(
            event=data.get("event", ""),
            device_id=data.get("deviceID"),
            udid=data.get("udid"),
            properties=data.get("properties"),
            raw=data,
        )


@dataclass(frozen=True)
class CpuUsageSample:
    """``event: sample`` on ``GET /device/{udid}/sysmontap``.

    A periodic system-load sample. ``total_load`` is the aggregate CPU load and
    ``system_load`` / ``user_load`` its kernel/user split (all percentages).
    """

    total_load: Optional[float] = None
    system_load: Optional[float] = None
    user_load: Optional[float] = None
    raw: Dict[str, Any] = field(default_factory=dict, repr=False)

    @classmethod
    def from_data(cls, data: Dict[str, Any]) -> "CpuUsageSample":
        return cls(
            total_load=data.get("CPU_TotalLoad"),
            system_load=data.get("SystemLoad"),
            user_load=data.get("UserLoad"),
            raw=data,
        )


@dataclass(frozen=True)
class JobLogLine:
    """``event: log`` on ``GET /device/{udid}/jobs/{id}/logs``.

    One log line emitted by an asynchronous job (test run / WDA / forward).
    """

    line: str
    raw: Dict[str, Any] = field(default_factory=dict, repr=False)

    @classmethod
    def from_data(cls, data: Dict[str, Any]) -> "JobLogLine":
        return cls(line=data.get("line", ""), raw=data)


@dataclass(frozen=True)
class UnknownEvent:
    """A frame whose ``event:`` name is not part of the known spec.

    Surfaced rather than dropped so a newer server can add event types without
    breaking older SDK consumers.
    """

    event: str
    data: Any
    raw: Optional[str] = None


# ---------------------------------------------------------------------------
# Event-name -> decoder registry, keyed per DESIGN.md x-sse-events mapping.
# ---------------------------------------------------------------------------

_DECODERS = {
    "appstate": AppStateNotification.from_data,
    "syslog": SyslogMessage.from_data,
    "ostrace": OsTraceEntry.from_data,
    "attachdetach": AttachDetachEvent.from_data,
    "sample": CpuUsageSample.from_data,
    "log": JobLogLine.from_data,
    "heartbeat": lambda data: Heartbeat(raw=data if isinstance(data, dict) else {}),
}


def decode_event(event_name: str, data: Any) -> Any:
    """Decode a parsed SSE ``data`` payload into its typed dataclass.

    Returns an :class:`UnknownEvent` for event names not covered by the spec.
    ``data`` is the already-``json.loads``-ed payload (usually a ``dict``).
    """
    decoder = _DECODERS.get(event_name)
    if decoder is None:
        return UnknownEvent(event=event_name, data=data)
    if not isinstance(data, dict):
        return UnknownEvent(event=event_name, data=data)
    return decoder(data)
