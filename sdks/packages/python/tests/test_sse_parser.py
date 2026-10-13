"""Unit tests for the SSE parser against canned event-stream bytes."""

from __future__ import annotations

import asyncio
from typing import AsyncIterator, Iterator, List

import pytest

from go_ios_sdk.events import (
    AppStateNotification,
    AttachDetachEvent,
    Heartbeat,
    OsTraceEntry,
    SyslogMessage,
    UnknownEvent,
)
from go_ios_sdk.sse import aiter_events, iter_events, iter_sse_frames

SYSLOG_FRAME = b'event: syslog\ndata: {"message":"hello","timestamp":1723200000000}\n\n'
APPSTATE_FRAME = (
    b"event: appstate\n"
    b'data: {"bundleId":"com.apple.Preferences","state":"foreground","timestamp":1}\n\n'
)
HEARTBEAT_FRAME = b"event: heartbeat\ndata: {}\n\n"
OSTRACE_FRAME = (
    b"event: ostrace\n"
    b'data: {"pid":123,"processName":"SpringBoard","level":"info","message":"m"}\n\n'
)
ATTACH_FRAME = (
    b"event: attachdetach\n"
    b'data: {"event":"attached","deviceID":5,"udid":"UDID","properties":{"serialNumber":"UDID"}}\n\n'
)
UNKNOWN_FRAME = b'event: somethingnew\ndata: {"x":1}\n\n'


def _sync_iter(chunks: List[bytes]) -> Iterator[bytes]:
    yield from chunks


def test_single_frame_typed_decode() -> None:
    events = list(iter_events(_sync_iter([SYSLOG_FRAME])))
    assert len(events) == 1
    ev = events[0]
    assert isinstance(ev, SyslogMessage)
    assert ev.message == "hello"
    assert ev.timestamp == 1723200000000


def test_multi_frame_stream_dispatches_each_type() -> None:
    stream = [APPSTATE_FRAME, OSTRACE_FRAME, ATTACH_FRAME]
    events = list(iter_events(_sync_iter(stream)))
    assert isinstance(events[0], AppStateNotification)
    assert events[0].bundle_id == "com.apple.Preferences"
    assert events[0].state == "foreground"
    assert isinstance(events[1], OsTraceEntry)
    assert events[1].pid == 123
    assert events[1].process_name == "SpringBoard"
    assert isinstance(events[2], AttachDetachEvent)
    assert events[2].event == "attached"
    assert events[2].device_id == 5
    assert events[2].properties == {"serialNumber": "UDID"}


def test_heartbeat_skipped_by_default_but_surfaced_when_requested() -> None:
    stream = [SYSLOG_FRAME, HEARTBEAT_FRAME, SYSLOG_FRAME]
    default = list(iter_events(_sync_iter(stream)))
    assert len(default) == 2
    assert all(isinstance(e, SyslogMessage) for e in default)

    with_hb = list(iter_events(_sync_iter(stream), include_heartbeats=True))
    assert len(with_hb) == 3
    assert isinstance(with_hb[1], Heartbeat)


def test_unknown_event_is_surfaced_not_dropped() -> None:
    events = list(iter_events(_sync_iter([UNKNOWN_FRAME])))
    assert len(events) == 1
    assert isinstance(events[0], UnknownEvent)
    assert events[0].event == "somethingnew"
    assert events[0].data == {"x": 1}


def test_split_chunks_across_arbitrary_boundaries() -> None:
    # Concatenate two frames, then split the bytes at every single-byte boundary
    # to prove the parser tolerates chunk splits anywhere (including mid-field,
    # mid-json, and on the blank-line terminator).
    blob = SYSLOG_FRAME + APPSTATE_FRAME
    for split in range(1, len(blob)):
        chunks = [blob[:split], blob[split:]]
        events = list(iter_events(_sync_iter(chunks)))
        assert len(events) == 2, f"split at {split} lost a frame"
        assert isinstance(events[0], SyslogMessage)
        assert isinstance(events[1], AppStateNotification)


def test_byte_at_a_time_streaming() -> None:
    blob = SYSLOG_FRAME + HEARTBEAT_FRAME + APPSTATE_FRAME
    chunks = [blob[i : i + 1] for i in range(len(blob))]
    events = list(iter_events(_sync_iter(chunks)))
    assert len(events) == 2  # heartbeat filtered
    assert isinstance(events[0], SyslogMessage)
    assert isinstance(events[1], AppStateNotification)


def test_crlf_line_endings_are_handled() -> None:
    frame = SYSLOG_FRAME.replace(b"\n", b"\r\n")
    events = list(iter_events(_sync_iter([frame])))
    assert len(events) == 1
    assert isinstance(events[0], SyslogMessage)


def test_multiple_data_lines_are_joined() -> None:
    frame = b"event: syslog\ndata: {\"message\":\ndata: \"multi\"}\n\n"
    frames = list(iter_sse_frames(_sync_iter([frame])))
    assert frames[0].data == '{"message":\n"multi"}'


def test_comment_and_blank_leading_lines_ignored() -> None:
    blob = b": keepalive comment\n\n" + SYSLOG_FRAME
    events = list(iter_events(_sync_iter([blob])))
    assert len(events) == 1
    assert isinstance(events[0], SyslogMessage)


# --------------------------------------------------------------------------
# Async path
# --------------------------------------------------------------------------


async def _async_iter(chunks: List[bytes]) -> AsyncIterator[bytes]:
    for c in chunks:
        await asyncio.sleep(0)
        yield c


@pytest.mark.asyncio
async def test_async_multi_frame() -> None:
    stream = [SYSLOG_FRAME, HEARTBEAT_FRAME, APPSTATE_FRAME]
    events = [e async for e in aiter_events(_async_iter(stream))]
    assert len(events) == 2
    assert isinstance(events[0], SyslogMessage)
    assert isinstance(events[1], AppStateNotification)


@pytest.mark.asyncio
async def test_async_split_chunks() -> None:
    blob = SYSLOG_FRAME + APPSTATE_FRAME
    chunks = [blob[:7], blob[7:20], blob[20:]]
    events = [e async for e in aiter_events(_async_iter(chunks))]
    assert len(events) == 2


@pytest.mark.asyncio
async def test_async_cancellation_stops_iteration() -> None:
    async def infinite() -> AsyncIterator[bytes]:
        while True:
            await asyncio.sleep(0.001)
            yield SYSLOG_FRAME

    seen = 0

    async def consume() -> None:
        nonlocal seen
        async for _ev in aiter_events(infinite()):
            seen += 1
            if seen >= 3:
                await asyncio.sleep(3600)  # block until cancelled

    task = asyncio.create_task(consume())
    # let it collect a few, then cancel
    while seen < 3:
        await asyncio.sleep(0.001)
    task.cancel()
    with pytest.raises(asyncio.CancelledError):
        await task
    assert seen == 3
