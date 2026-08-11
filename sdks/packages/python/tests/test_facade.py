"""Facade tests against a mocked httpx transport (no real device/server)."""

from __future__ import annotations

from typing import List, Tuple

import httpx
import pytest

from go_ios_sdk import (
    AsyncIosClient,
    CpuUsageSample,
    IosClient,
    JobLogLine,
    SyslogMessage,
)
from go_ios_sdk.errors import ApiError

BASE = "http://ios.test"

DEVICE_LIST = {
    "deviceList": [
        {"deviceID": 1, "properties": {"serialNumber": "UDID1"}},
        {"deviceID": 2, "properties": {"serialNumber": "UDID2"}},
    ]
}
SYSLOG_STREAM = (
    b'event: syslog\ndata: {"message":"line1","timestamp":1}\n\n'
    b"event: heartbeat\ndata: {}\n\n"
    b'event: syslog\ndata: {"message":"line2","timestamp":2}\n\n'
)
SYSMONTAP_STREAM = (
    b'event: sample\ndata: {"CPU_TotalLoad":12.5,"SystemLoad":4.0,"UserLoad":8.5}\n\n'
    b"event: heartbeat\ndata: {}\n\n"
    b'event: sample\ndata: {"CPU_TotalLoad":20.0,"SystemLoad":5.0,"UserLoad":15.0}\n\n'
)
JOB_LOGS_STREAM = (
    b'event: log\ndata: {"line":"starting"}\n\n'
    b"event: heartbeat\ndata: {}\n\n"
    b'event: log\ndata: {"line":"done"}\n\n'
)


def _handler(record: List[httpx.Request]):
    def handle(request: httpx.Request) -> httpx.Response:
        record.append(request)
        path = request.url.path
        if path == "/api/v1/list":
            return httpx.Response(200, json=DEVICE_LIST)
        if path == "/api/v1/tunnels":
            return httpx.Response(200, json=[{"udid": "UDID1"}])
        if path.endswith("/info"):
            return httpx.Response(200, json={"ProductType": "iPhone14,2"})
        if path.endswith("/screenshot") or path.endswith("/wallpaper"):
            return httpx.Response(200, content=b"\x89PNG\r\n", headers={"content-type": "image/png"})
        if path.endswith("/files/pull"):
            return httpx.Response(200, content=b"FILEDATA",
                                  headers={"content-type": "application/octet-stream"})
        if path.endswith("/setlocation"):
            return httpx.Response(200, json={"message": "ok"})
        if path.endswith("/apps/launch"):
            return httpx.Response(200, json={"message": "launched"})
        if path.endswith("/enable-condition"):
            return httpx.Response(200, json={"message": "enabled"})
        if path.endswith("/battery"):
            return httpx.Response(200, json={"CurrentCapacity": 80, "IsCharging": True})
        if path.endswith("/wda/session"):
            return httpx.Response(200, json={"sessionId": "s1", "udid": "UDID1", "config": {}})
        if path.endswith("/syslog"):
            return httpx.Response(200, content=SYSLOG_STREAM,
                                  headers={"content-type": "text/event-stream"})
        if path.endswith("/sysmontap"):
            return httpx.Response(200, content=SYSMONTAP_STREAM,
                                  headers={"content-type": "text/event-stream"})
        if path.endswith("/logs"):
            return httpx.Response(200, content=JOB_LOGS_STREAM,
                                  headers={"content-type": "text/event-stream"})
        if path.endswith("/missing"):
            return httpx.Response(404, json={"message": "device not found"})
        return httpx.Response(200, json={"message": "ok"})

    return handle


def _sync_client() -> Tuple[IosClient, List[httpx.Request]]:
    record: List[httpx.Request] = []
    transport = httpx.MockTransport(_handler(record))
    http = httpx.Client(base_url=BASE, transport=transport,
                        headers={"Authorization": "Bearer tok"})
    return IosClient(base_url=BASE, api_key="tok", http_client=http), record


def test_devices_list() -> None:
    client, record = _sync_client()
    result = client.devices.list()
    assert result["deviceList"][0]["properties"]["serialNumber"] == "UDID1"
    assert record[0].url.path == "/api/v1/list"
    assert record[0].headers["Authorization"] == "Bearer tok"


def test_device_info_and_screenshot() -> None:
    client, _ = _sync_client()
    dev = client.device("UDID1")
    assert dev.info()["ProductType"] == "iPhone14,2"
    png = dev.screenshot()
    assert isinstance(png, bytes) and png.startswith(b"\x89PNG")


def test_set_location_uses_longitude_spelling() -> None:
    client, record = _sync_client()
    client.device("UDID1").set_location(52.5, 13.4)
    req = record[-1]
    assert req.method == "PUT"
    assert dict(req.url.params) == {"latitude": "52.5", "longitude": "13.4"}


def test_apps_launch_query_param() -> None:
    client, record = _sync_client()
    client.device("UDID1").apps.launch("com.foo.bar")
    req = record[-1]
    assert req.url.params["bundleID"] == "com.foo.bar"
    assert req.url.path.endswith("/apps/launch")


def test_enable_condition_params() -> None:
    client, record = _sync_client()
    client.device("UDID1").enable_condition("pt1", "p1")
    req = record[-1]
    assert dict(req.url.params) == {"profileTypeID": "pt1", "profileID": "p1"}


def test_wda_create_session_posts_json() -> None:
    client, record = _sync_client()
    cfg = {"bundleId": "b", "testBundleId": "t", "xcTestConfig": "c"}
    out = client.device("UDID1").wda.create_session(cfg)
    assert out["sessionId"] == "s1"


def test_error_maps_to_api_error() -> None:
    client, _ = _sync_client()
    with pytest.raises(ApiError) as ei:
        client.device("UDID1")._get_json("/missing")
    assert ei.value.status_code == 404
    assert ei.value.message == "device not found"


def test_sync_syslog_stream_typed_and_filters_heartbeat() -> None:
    client, _ = _sync_client()
    events = list(client.device("UDID1").syslog())
    assert len(events) == 2
    assert all(isinstance(e, SyslogMessage) for e in events)
    assert [e.message for e in events] == ["line1", "line2"]


def test_devices_udids_convenience() -> None:
    client, _ = _sync_client()
    assert client.devices.udids() == ["UDID1", "UDID2"]


def test_battery_get() -> None:
    client, record = _sync_client()
    out = client.device("UDID1").battery()
    assert out["CurrentCapacity"] == 80
    assert record[-1].url.path.endswith("/battery")


def test_mobilegestalt_repeats_key_query() -> None:
    client, record = _sync_client()
    client.device("UDID1").mobilegestalt(keys=["ProductName", "BuildVersion"])
    req = record[-1]
    assert req.url.params.get_list("key") == ["ProductName", "BuildVersion"]


def test_erase_requires_confirm_param() -> None:
    client, record = _sync_client()
    client.device("UDID1").erase(confirm=True)
    req = record[-1]
    assert req.method == "POST"
    assert dict(req.url.params) == {"confirm": "true"}


def test_set_devmode_body() -> None:
    client, record = _sync_client()
    client.device("UDID1").set_devmode("reveal", enable_post_restart=True)
    body = record[-1].read()
    assert b'"action":"reveal"' in body
    assert b'"enablePostRestart":true' in body


def test_set_assistive_touch_body() -> None:
    client, record = _sync_client()
    client.device("UDID1").set_assistive_touch(True)
    assert record[-1].method == "PUT"
    assert record[-1].read() == b'{"enabled":true}'


def test_set_pasteboard_text_plain() -> None:
    client, record = _sync_client()
    client.device("UDID1").set_pasteboard("hello")
    req = record[-1]
    assert req.method == "PUT"
    assert req.headers["Content-Type"] == "text/plain"
    assert req.read() == b"hello"


def test_files_pull_returns_bytes() -> None:
    client, record = _sync_client()
    data = client.device("UDID1").files.pull("/Documents/a.txt", domain="app")
    assert data == b"FILEDATA"
    req = record[-1]
    assert dict(req.url.params) == {"domain": "app", "remote": "/Documents/a.txt"}


def test_files_push_octet_stream() -> None:
    client, record = _sync_client()
    client.device("UDID1").files.push("/Documents/a.txt", b"payload")
    req = record[-1]
    assert req.method == "POST"
    assert req.headers["Content-Type"] == "application/octet-stream"
    assert req.read() == b"payload"


def test_crashes_list_and_remove() -> None:
    client, record = _sync_client()
    client.device("UDID1").crashes.remove("Foo-*", cwd="/tmp")
    req = record[-1]
    assert req.method == "DELETE"
    assert dict(req.url.params) == {"cwd": "/tmp", "pattern": "Foo-*"}


def test_jobs_forward_posts_json() -> None:
    client, record = _sync_client()
    client.device("UDID1").jobs.forward({"hostPort":8100,"targetPort":8100})
    req = record[-1]
    assert req.url.path.endswith("/jobs/forward")
    assert req.read() == b'{"hostPort":8100,"targetPort":8100}'


def test_get_wallpaper_returns_bytes() -> None:
    client, _ = _sync_client()
    png = client.device("UDID1").get_wallpaper()
    assert isinstance(png, bytes) and png.startswith(b"\x89PNG")


def test_tunnels_list_and_refresh() -> None:
    client, record = _sync_client()
    assert client.tunnels.list() == [{"udid": "UDID1"}]
    client.tunnels.refresh("UDID1")
    req = record[-1]
    assert req.method == "POST"
    assert req.url.path == "/api/v1/tunnels/UDID1/refresh"


def test_tunnel_agent_shutdown() -> None:
    client, record = _sync_client()
    client.tunnels.shutdown_agent()
    assert record[-1].url.path == "/api/v1/tunnel-agent/shutdown"


def test_remove_profile_path() -> None:
    client, record = _sync_client()
    client.device("UDID1").remove_profile("com.example.profile")
    req = record[-1]
    assert req.method == "DELETE"
    assert req.url.path.endswith("/profiles/com.example.profile")


def test_add_profile_multipart() -> None:
    client, record = _sync_client()
    client.device("UDID1").add_profile(b"<plist/>", p12=b"cert", password="pw")
    req = record[-1]
    assert req.method == "POST"
    assert req.headers["Content-Type"].startswith("multipart/form-data")
    body = req.read()
    assert b"<plist/>" in body and b"cert" in body and b"pw" in body


def test_sync_sysmontap_stream_typed() -> None:
    client, _ = _sync_client()
    events = list(client.device("UDID1").sysmontap())
    assert len(events) == 2
    assert all(isinstance(e, CpuUsageSample) for e in events)
    assert events[0].total_load == 12.5
    assert events[1].user_load == 15.0


def test_sync_job_logs_stream_typed() -> None:
    client, _ = _sync_client()
    lines = list(client.device("UDID1").jobs.logs("job-1"))
    assert [ln.line for ln in lines] == ["starting", "done"]
    assert all(isinstance(ln, JobLogLine) for ln in lines)


# --------------------------------------------------------------------------
# Async
# --------------------------------------------------------------------------


def _async_client() -> Tuple[AsyncIosClient, List[httpx.Request]]:
    record: List[httpx.Request] = []
    transport = httpx.MockTransport(_handler(record))
    http = httpx.AsyncClient(base_url=BASE, transport=transport,
                             headers={"Authorization": "Bearer tok"})
    return AsyncIosClient(base_url=BASE, api_key="tok", http_client=http), record


@pytest.mark.asyncio
async def test_async_info_and_list() -> None:
    client, _ = _async_client()
    async with client:
        result = await client.devices.list()
        assert result["deviceList"][0]["deviceID"] == 1
        info = await client.device("UDID1").info()
        assert info["ProductType"] == "iPhone14,2"


@pytest.mark.asyncio
async def test_async_syslog_stream() -> None:
    client, _ = _async_client()
    async with client:
        events = [e async for e in client.device("UDID1").syslog()]
    assert [e.message for e in events] == ["line1", "line2"]


@pytest.mark.asyncio
async def test_async_stream_context_manager() -> None:
    client, _ = _async_client()
    async with client:
        collected = []
        async with client.device("UDID1").syslog() as stream:
            async for ev in stream:
                collected.append(ev)
        assert len(collected) == 2


@pytest.mark.asyncio
async def test_async_sysmontap_stream() -> None:
    client, _ = _async_client()
    async with client:
        events = [e async for e in client.device("UDID1").sysmontap()]
    assert [e.total_load for e in events] == [12.5, 20.0]
    assert all(isinstance(e, CpuUsageSample) for e in events)


@pytest.mark.asyncio
async def test_async_job_logs_stream() -> None:
    client, _ = _async_client()
    async with client:
        lines = [ln async for ln in client.device("UDID1").jobs.logs("job-1")]
    assert [ln.line for ln in lines] == ["starting", "done"]
    assert all(isinstance(ln, JobLogLine) for ln in lines)


@pytest.mark.asyncio
async def test_async_udids_and_set_time_format() -> None:
    client, record = _async_client()
    async with client:
        assert await client.devices.udids() == ["UDID1", "UDID2"]
        await client.device("UDID1").set_time_format(True)
    req = record[-1]
    assert req.method == "PUT"
    assert req.url.path.endswith("/timeformat")
    assert req.read() == b'{"uses24Hour":true}'
