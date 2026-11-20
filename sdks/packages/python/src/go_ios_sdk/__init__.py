"""go-ios SDK — ergonomic Python client for the go-ios REST API.

Import name is ``go_ios_sdk`` (PyPI package ``go-ios-sdk``)::

    from go_ios_sdk import IosClient, AsyncIosClient

    with IosClient(base_url="http://localhost:60105", api_key="...") as client:
        for dev in client.devices.list()["deviceList"]:
            print(dev)
        png = client.device(udid).screenshot()

See :class:`IosClient` / :class:`AsyncIosClient` for the full facade, and
:mod:`go_ios_sdk.events` for the typed streaming event payloads.
"""

from __future__ import annotations

from .async_client import (
    AsyncApps,
    AsyncCrashes,
    AsyncDevice,
    AsyncDevices,
    AsyncFiles,
    AsyncFsync,
    AsyncIosClient,
    AsyncJobs,
    AsyncPrepare,
    AsyncSign,
    AsyncTunnels,
    AsyncUi,
    AsyncWda,
    AsyncWebInspector,
)
from .client import (
    Apps,
    Crashes,
    Device,
    Devices,
    Files,
    Fsync,
    IosClient,
    Jobs,
    Prepare,
    Sign,
    Tunnels,
    Ui,
    Wda,
    WebInspector,
)
from .discovery import DiscoveryError, discover_base_url, go_ios_home
from .errors import ApiError, GoIosError
from .events import (
    AppStateNotification,
    AttachDetachEvent,
    CpuUsageSample,
    Heartbeat,
    JobLogLine,
    OsTraceEntry,
    SyslogMessage,
    UnknownEvent,
)

__version__ = "0.1.0"

__all__ = [
    # clients
    "IosClient",
    "AsyncIosClient",
    # handles (sync)
    "Devices",
    "Device",
    "Apps",
    "Wda",
    "Files",
    "Crashes",
    "Jobs",
    "Tunnels",
    "Fsync",
    "WebInspector",
    "Ui",
    "Sign",
    "Prepare",
    # handles (async)
    "AsyncDevices",
    "AsyncDevice",
    "AsyncApps",
    "AsyncWda",
    "AsyncFiles",
    "AsyncCrashes",
    "AsyncJobs",
    "AsyncTunnels",
    "AsyncFsync",
    "AsyncWebInspector",
    "AsyncUi",
    "AsyncSign",
    "AsyncPrepare",
    # errors
    "GoIosError",
    "ApiError",
    "DiscoveryError",
    # discovery
    "discover_base_url",
    "go_ios_home",
    # event payloads
    "AppStateNotification",
    "SyslogMessage",
    "OsTraceEntry",
    "AttachDetachEvent",
    "CpuUsageSample",
    "JobLogLine",
    "Heartbeat",
    "UnknownEvent",
    "__version__",
]
