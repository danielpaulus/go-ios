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
    AsyncIosClient,
    AsyncJobs,
    AsyncTunnels,
    AsyncWda,
)
from .client import (
    Apps,
    Crashes,
    Device,
    Devices,
    Files,
    IosClient,
    Jobs,
    Tunnels,
    Wda,
)
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
    # handles (async)
    "AsyncDevices",
    "AsyncDevice",
    "AsyncApps",
    "AsyncWda",
    "AsyncFiles",
    "AsyncCrashes",
    "AsyncJobs",
    "AsyncTunnels",
    # errors
    "GoIosError",
    "ApiError",
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
