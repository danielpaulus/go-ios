"""A minimal, dependency-free Server-Sent Events (SSE) parser.

The go-ios streaming endpoints emit ``text/event-stream`` frames of the form::

    event: <event-name>\\n
    data: <compact-json>\\n
    \\n

(see ``docs/DESIGN.md``). This module turns a byte stream into typed events:

* :func:`iter_sse_frames` / :func:`aiter_sse_frames` -- split a raw byte iterator
  into :class:`SseFrame` objects, tolerant of chunk boundaries falling anywhere.
* :func:`iter_events` / :func:`aiter_events` -- decode frames into the typed
  dataclasses from :mod:`go_ios_sdk.events`, JSON-loading the ``data`` payload,
  optionally skipping ``heartbeat`` frames.

The parser follows the SSE spec's line handling (``\\n``, ``\\r\\n`` and ``\\r``
line terminators, ``field: value`` with an optional single leading space on the
value, multiple ``data:`` lines joined with ``\\n``) which is a strict superset
of what the go-ios server emits.
"""

from __future__ import annotations

import json
from dataclasses import dataclass, field
from typing import Any, AsyncIterator, Iterator, List, Optional

from .events import decode_event


@dataclass
class SseFrame:
    """One dispatched SSE event (a block of lines terminated by a blank line)."""

    event: str = "message"
    data: str = ""
    id: Optional[str] = None
    retry: Optional[int] = None
    _data_lines: List[str] = field(default_factory=list, repr=False)


class _FrameAccumulator:
    """Incremental SSE line/frame state machine shared by sync and async paths.

    Feed it decoded text via :meth:`feed`; it yields completed :class:`SseFrame`
    objects. Byte-level buffering (and UTF-8 decoding across chunk boundaries) is
    handled by the callers.
    """

    def __init__(self) -> None:
        self._buf = ""
        self._cur = SseFrame(_data_lines=[])
        self._has_fields = False

    def feed(self, text: str) -> Iterator[SseFrame]:
        self._buf += text
        # Normalise CRLF / lone CR into LF so splitting is uniform.
        self._buf = self._buf.replace("\r\n", "\n").replace("\r", "\n")
        while "\n" in self._buf:
            line, self._buf = self._buf.split("\n", 1)
            frame = self._consume_line(line)
            if frame is not None:
                yield frame

    def _consume_line(self, line: str) -> Optional[SseFrame]:
        if line == "":
            # Blank line -> dispatch the accumulated frame (if any fields seen).
            if not self._has_fields:
                return None
            self._cur.data = "\n".join(self._cur._data_lines)
            frame = self._cur
            self._cur = SseFrame(_data_lines=[])
            self._has_fields = False
            return frame

        if line.startswith(":"):
            # Comment line; ignore.
            return None

        if ":" in line:
            field_name, value = line.split(":", 1)
            if value.startswith(" "):
                value = value[1:]
        else:
            field_name, value = line, ""

        self._has_fields = True
        if field_name == "event":
            self._cur.event = value
        elif field_name == "data":
            self._cur._data_lines.append(value)
        elif field_name == "id":
            self._cur.id = value
        elif field_name == "retry":
            try:
                self._cur.retry = int(value)
            except ValueError:
                pass
        # Unknown fields are ignored per the SSE spec.
        return None


def iter_sse_frames(chunks: Iterator[bytes]) -> Iterator[SseFrame]:
    """Split a synchronous iterator of raw bytes into :class:`SseFrame` objects."""
    acc = _FrameAccumulator()
    decoder = _incremental_utf8()
    next(decoder)  # prime the generator so it is ready for .send()
    for chunk in chunks:
        text = decoder.send(chunk)
        if text:
            yield from acc.feed(text)


async def aiter_sse_frames(chunks: AsyncIterator[bytes]) -> AsyncIterator[SseFrame]:
    """Split an async iterator of raw bytes into :class:`SseFrame` objects."""
    acc = _FrameAccumulator()
    decoder = _incremental_utf8()
    next(decoder)  # prime the generator so it is ready for .send()
    async for chunk in chunks:
        text = decoder.send(chunk)
        if text:
            for frame in acc.feed(text):
                yield frame


def _incremental_utf8():
    """A tiny generator that decodes UTF-8 bytes, buffering partial multibyte
    sequences that straddle chunk boundaries. ``.send(b)`` -> decoded ``str``."""
    import codecs

    dec = codecs.getincrementaldecoder("utf-8")()
    out = ""
    while True:
        chunk = yield out
        out = dec.decode(chunk)


def _decode_frame(frame: SseFrame) -> Any:
    """JSON-load a frame's data and dispatch to a typed event."""
    raw = frame.data
    try:
        payload: Any = json.loads(raw) if raw else {}
    except json.JSONDecodeError:
        payload = raw
    return decode_event(frame.event, payload)


def iter_events(
    chunks: Iterator[bytes],
    *,
    include_heartbeats: bool = False,
) -> Iterator[Any]:
    """Decode a synchronous byte iterator into typed go-ios events.

    Heartbeat frames are skipped unless ``include_heartbeats`` is True.
    """
    for frame in iter_sse_frames(chunks):
        if frame.event == "heartbeat" and not include_heartbeats:
            continue
        yield _decode_frame(frame)


async def aiter_events(
    chunks: AsyncIterator[bytes],
    *,
    include_heartbeats: bool = False,
) -> AsyncIterator[Any]:
    """Decode an async byte iterator into typed go-ios events.

    Heartbeat frames are skipped unless ``include_heartbeats`` is True.
    """
    async for frame in aiter_sse_frames(chunks):
        if frame.event == "heartbeat" and not include_heartbeats:
            continue
        yield _decode_frame(frame)
