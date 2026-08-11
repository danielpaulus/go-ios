"""Ergonomic synchronous facade for the go-ios REST API.

Public shape (mirrored by the TypeScript/Java/C# SDKs)::

    client = IosClient(base_url="http://localhost:60105", api_key="...")
    client.devices.list()
    client.tunnels.list(); client.tunnels.refresh(udid); client.tunnels.shutdown_agent()
    dev = client.device(udid)

    # device info
    dev.info(); dev.device_name(); dev.date()
    dev.battery(); dev.diagnostics(); dev.mobilegestalt(keys=[...])
    dev.processes(); dev.lockdown()

    # management
    dev.reboot(); dev.shutdown(); dev.erase(); dev.devmode(); dev.set_devmode(...)
    dev.lang(); dev.set_lang(...); dev.memlimitoff(...)

    # files / crashes
    dev.files.ls(path); dev.files.pull(path); dev.files.push(path, data)
    dev.crashes.list(); dev.crashes.remove(pattern)

    # media
    dev.get_wallpaper(); dev.set_wallpaper(...); dev.get_icon_layout()
    dev.set_icon_layout(...); dev.get_pasteboard(); dev.set_pasteboard(text)

    # profiles / images
    dev.profiles(); dev.add_profile(...); dev.remove_profile(name)
    dev.images(); dev.install_image(...); dev.unmount_image()

    # settings
    dev.assistive_touch(); dev.set_assistive_touch(True)
    dev.time_format(); dev.set_time_format(True)
    dev.set_wifi(...); dev.remove_wifi(ssid)

    # mdm / proxy
    dev.security_info(...); dev.fetch_unlock_token(...); dev.clear_passcode(...)
    dev.clear_screen_time_password(...)
    dev.set_http_proxy(...); dev.remove_http_proxy()

    # apps / wda / conditions / location
    dev.apps.list(); dev.apps.launch(bundle_id); ...
    dev.wda.create_session(config); ...
    dev.conditions(); dev.enable_condition(pt, p); dev.disable_condition()
    dev.set_location(lat, lon); dev.reset_location(); dev.reset_accessibility()

    # jobs (async server-side operations)
    dev.jobs.runtest(cfg); dev.jobs.runwda(cfg); dev.jobs.forward(cfg)
    dev.jobs.list(); dev.jobs.get(id); dev.jobs.delete(id)

    # streaming SSE
    for event in dev.syslog(): ...
    for event in dev.notifications(): ...
    for event in dev.ostrace(pid=123): ...
    for event in dev.listen(): ...
    for sample in dev.sysmontap(): ...
    for line in dev.jobs.logs(id): ...

Streaming methods return generators/context managers that close the underlying
HTTP response when iteration ends (or the ``with`` block exits).
"""

from __future__ import annotations

import io
import os
from typing import Any, Dict, Iterator, List, Optional, Sequence, Union

import httpx

from . import events as _events
from ._http import API_PREFIX, build_headers, json_or_none, raise_for_status
from .sse import iter_events

DEFAULT_BASE_URL = "http://localhost:60105"
# Streaming endpoints must not time out on read (they are long-lived, idle-tolerant
# thanks to heartbeats). Connect timeout is still applied.
_STREAM_TIMEOUT = httpx.Timeout(connect=10.0, read=None, write=10.0, pool=10.0)

BytesLike = Union[bytes, bytearray, io.IOBase, "os.PathLike[str]", str]


def _resolve_body(source: BytesLike) -> bytes:
    """Read a bytes/path/file-like source into bytes for a request body."""
    if isinstance(source, (bytes, bytearray)):
        return bytes(source)
    if isinstance(source, io.IOBase):
        data = source.read()
        return data if isinstance(data, bytes) else str(data).encode()
    # str / PathLike -> treat as a filesystem path
    with open(source, "rb") as fh:  # type: ignore[arg-type]
        return fh.read()


def _file_tuple(source: BytesLike, default_name: str) -> Any:
    """Build a value for httpx multipart ``files=`` from a path/bytes/file-like."""
    if isinstance(source, (bytes, bytearray)):
        return (default_name, bytes(source), "application/octet-stream")
    if isinstance(source, io.IOBase):
        name = getattr(source, "name", default_name)
        return (os.path.basename(str(name)), source, "application/octet-stream")
    name = os.path.basename(str(source))
    return (name, open(source, "rb"), "application/octet-stream")  # noqa: SIM115


def _close_multipart(files: Dict[str, Any], sources: Dict[str, Any]) -> None:
    """Close only the file handles the facade opened itself (path sources)."""
    for field, tup in files.items():
        fh = tup[1]
        src = sources.get(field)
        if (
            hasattr(fh, "close")
            and not isinstance(src, (bytes, bytearray))
            and not isinstance(src, io.IOBase)
        ):
            fh.close()


def _bool_param(value: bool) -> str:
    return "true" if value else "false"


class _SyncStream:
    """A closeable sync generator wrapper for an SSE stream.

    Usable directly as an iterator (``for ev in stream``) or as a context manager
    (``with dev.syslog() as s: ...``); either way the HTTP response is released
    when iteration finishes or the block exits.
    """

    def __init__(self, request_cm: Any, include_heartbeats: bool) -> None:
        self._request_cm = request_cm
        self._response: Optional[httpx.Response] = None
        self._include_heartbeats = include_heartbeats

    def _open(self) -> httpx.Response:
        if self._response is None:
            self._response = self._request_cm.__enter__()
            raise_for_status(self._response)
        return self._response

    def __iter__(self) -> Iterator[Any]:
        resp = self._open()
        try:
            yield from iter_events(
                resp.iter_bytes(), include_heartbeats=self._include_heartbeats
            )
        finally:
            self.close()

    def __enter__(self) -> "_SyncStream":
        self._open()
        return self

    def __exit__(self, *exc: Any) -> None:
        self.close()

    def close(self) -> None:
        if self._response is not None:
            try:
                self._request_cm.__exit__(None, None, None)
            finally:
                self._response = None


class _SyncBytesStream:
    """A closeable sync generator of raw byte chunks from a long-lived response.

    Backs the v3 **binary** streams (``ui/stream``, ``screenshot/stream``,
    ``pcap``) which are *not* SSE — they emit an unframed chunked byte stream. Use
    it as an iterator (``for chunk in stream``) or a context manager
    (``with dev.pcap() as s: ...``); either way the HTTP response is released when
    iteration finishes, the block exits, or ``close()`` is called (cancelable).
    """

    def __init__(self, request_cm: Any) -> None:
        self._request_cm = request_cm
        self._response: Optional[httpx.Response] = None

    def _open(self) -> httpx.Response:
        if self._response is None:
            self._response = self._request_cm.__enter__()
            raise_for_status(self._response)
        return self._response

    def __iter__(self) -> Iterator[bytes]:
        resp = self._open()
        try:
            yield from resp.iter_bytes()
        finally:
            self.close()

    def __enter__(self) -> "_SyncBytesStream":
        self._open()
        return self

    def __exit__(self, *exc: Any) -> None:
        self.close()

    def close(self) -> None:
        if self._response is not None:
            try:
                self._request_cm.__exit__(None, None, None)
            finally:
                self._response = None


class Apps:
    """App-management operations for a single device."""

    def __init__(self, device: "Device") -> None:
        self._d = device

    def list(self) -> List[Dict[str, Any]]:
        """List installed apps (``GET /apps/``)."""
        return self._d._get_json("/apps/") or []

    def launch(self, bundle_id: str) -> Dict[str, Any]:
        """Launch an app by bundle id (``POST /apps/launch``)."""
        return self._d._post_json("/apps/launch", params={"bundleID": bundle_id})

    def kill(self, bundle_id: str) -> Dict[str, Any]:
        """Kill a running app by bundle id (``POST /apps/kill``)."""
        return self._d._post_json("/apps/kill", params={"bundleID": bundle_id})

    def install(self, ipa: BytesLike) -> Dict[str, Any]:
        """Install an ``.ipa``/``.app`` archive (``POST /apps/install``).

        ``ipa`` may be a path, bytes, or a file-like object.
        """
        return self._d._post_multipart("/apps/install", {"file": (ipa, "app.ipa")})

    def uninstall(self, bundle_id: str) -> Dict[str, Any]:
        """Uninstall an app by bundle id (``POST /apps/uninstall``)."""
        return self._d._post_json("/apps/uninstall", params={"bundleID": bundle_id})


class Wda:
    """WebDriverAgent (XCUITest) session operations for a single device."""

    def __init__(self, device: "Device") -> None:
        self._d = device

    def create_session(self, config: Dict[str, Any]) -> Dict[str, Any]:
        """Create a WDA session (``POST /wda/session``). ``config`` is a WdaConfig dict."""
        return self._d._post_json("/wda/session", json=config)

    def read_session(self, session_id: str) -> Dict[str, Any]:
        """Read a WDA session (``GET /wda/session/{id}``)."""
        return self._d._get_json(f"/wda/session/{session_id}")

    def delete_session(self, session_id: str) -> Dict[str, Any]:
        """Delete a WDA session (``DELETE /wda/session/{id}``)."""
        return self._d._request_json("delete", f"/wda/session/{session_id}")


class Files:
    """On-device file-service operations for a single device."""

    def __init__(self, device: "Device") -> None:
        self._d = device

    def ls(
        self,
        path: Optional[str] = None,
        *,
        domain: str = "app",
        identifier: Optional[str] = None,
    ) -> Dict[str, Any]:
        """List files in a house-arrest domain (``GET /files``)."""
        params: Dict[str, Any] = {"domain": domain}
        if identifier is not None:
            params["identifier"] = identifier
        if path is not None:
            params["path"] = path
        return self._d._get_json("/files", params=params) or {}

    def pull(
        self,
        path: str,
        *,
        domain: str = "app",
        identifier: Optional[str] = None,
    ) -> bytes:
        """Pull a file's raw bytes off the device (``GET /files/pull``)."""
        params: Dict[str, Any] = {"domain": domain, "remote": path}
        if identifier is not None:
            params["identifier"] = identifier
        return self._d._get_bytes("/files/pull", params=params)

    def push(
        self,
        path: str,
        data: BytesLike,
        *,
        domain: str = "app",
        identifier: Optional[str] = None,
    ) -> Dict[str, Any]:
        """Push raw bytes to a file on the device (``POST /files/push``)."""
        params: Dict[str, Any] = {"domain": domain, "remote": path}
        if identifier is not None:
            params["identifier"] = identifier
        return self._d._request_json(
            "post",
            "/files/push",
            params=params,
            content=_resolve_body(data),
            headers={"Content-Type": "application/octet-stream"},
        )


class Crashes:
    """Crash-report operations for a single device."""

    def __init__(self, device: "Device") -> None:
        self._d = device

    def list(self, pattern: str = "*") -> List[str]:
        """List crash reports matching ``pattern`` (``GET /crashes``)."""
        return self._d._get_json("/crashes", params={"pattern": pattern}) or []

    def remove(self, pattern: str, *, cwd: str = ".") -> Dict[str, Any]:
        """Remove crash reports matching ``pattern`` under ``cwd`` (``DELETE /crashes``)."""
        return self._d._request_json(
            "delete", "/crashes", params={"cwd": cwd, "pattern": pattern}
        )


class Jobs:
    """Asynchronous server-side jobs (test runs, WDA, port forwards)."""

    def __init__(self, device: "Device") -> None:
        self._d = device

    def runtest(self, config: Dict[str, Any]) -> Dict[str, Any]:
        """Start an XCUITest run job (``POST /jobs/runtest``)."""
        return self._d._post_json("/jobs/runtest", json=config)

    def runwda(self, config: Dict[str, Any]) -> Dict[str, Any]:
        """Start a WebDriverAgent run job (``POST /jobs/runwda``)."""
        return self._d._post_json("/jobs/runwda", json=config)

    def forward(self, config: Dict[str, Any]) -> Dict[str, Any]:
        """Start a TCP port-forward job (``POST /jobs/forward``).

        ``config`` is a ``{"hostPort": int, "targetPort": int}`` dict.
        """
        return self._d._post_json("/jobs/forward", json=config)

    def list(self) -> List[Dict[str, Any]]:
        """List active jobs (``GET /jobs``)."""
        return self._d._get_json("/jobs") or []

    def get(self, job_id: str) -> Dict[str, Any]:
        """Get one job's status (``GET /jobs/{id}``)."""
        return self._d._get_json(f"/jobs/{job_id}") or {}

    def delete(self, job_id: str) -> Dict[str, Any]:
        """Stop/delete a job (``DELETE /jobs/{id}``)."""
        return self._d._request_json("delete", f"/jobs/{job_id}")

    def logs(self, job_id: str, *, include_heartbeats: bool = False) -> _SyncStream:
        """Stream a job's log lines (``GET /jobs/{id}/logs``) as typed events."""
        return self._d._stream(
            f"/jobs/{job_id}/logs", include_heartbeats=include_heartbeats
        )


class Fsync:
    """AFC file-sync operations for a single device (``/fsync/*``)."""

    def __init__(self, device: "Device") -> None:
        self._d = device

    def ls(self, path: Optional[str] = None, *, bundle_id: Optional[str] = None) -> Dict[str, Any]:
        """List a device directory over AFC (``GET /fsync/ls``)."""
        return self._d._get_json("/fsync/ls", params=_fsync_params(path, bundle_id)) or {}

    def tree(self, path: Optional[str] = None, *, bundle_id: Optional[str] = None) -> Dict[str, Any]:
        """Recursively list a device directory over AFC (``GET /fsync/tree``)."""
        return self._d._get_json("/fsync/tree", params=_fsync_params(path, bundle_id)) or {}

    def pull(self, path: str, *, bundle_id: Optional[str] = None) -> bytes:
        """Download a file's raw bytes over AFC (``GET /fsync/pull``)."""
        return self._d._get_bytes("/fsync/pull", params=_fsync_params(path, bundle_id))

    def push(
        self, path: str, data: BytesLike, *, bundle_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """Upload raw bytes to a device path over AFC (``POST /fsync/push``)."""
        return self._d._request_json(
            "post",
            "/fsync/push",
            params=_fsync_params(path, bundle_id),
            content=_resolve_body(data),
            headers={"Content-Type": "application/octet-stream"},
        )

    def rm(
        self, path: str, *, recursive: bool = False, bundle_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """Remove a file/directory over AFC (``DELETE /fsync/rm``)."""
        params = _fsync_params(path, bundle_id) or {}
        if recursive:
            params["recursive"] = "true"
        return self._d._request_json("delete", "/fsync/rm", params=params)

    def mkdir(self, path: str, *, bundle_id: Optional[str] = None) -> Dict[str, Any]:
        """Create a directory over AFC (``POST /fsync/mkdir``)."""
        return self._d._post_json("/fsync/mkdir", params=_fsync_params(path, bundle_id))


class WebInspector:
    """Safari Web Inspector operations for a single device (``/webinspector/*``)."""

    def __init__(self, device: "Device") -> None:
        self._d = device

    def pages(self) -> List[Dict[str, Any]]:
        """List inspectable pages (``GET /webinspector/pages``)."""
        return self._d._get_json("/webinspector/pages") or []

    def launch(self, url: str, *, bundle_id: Optional[str] = None) -> Dict[str, Any]:
        """Open a URL in a new inspectable page (``POST /webinspector/launch``)."""
        body: Dict[str, Any] = {"url": url}
        if bundle_id is not None:
            body["bundleId"] = bundle_id
        return self._d._post_json("/webinspector/launch", json=body)

    def eval(
        self, script: str, *, page: Optional[str] = None, bundle_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """Evaluate JavaScript against a page (``POST /webinspector/eval``)."""
        body: Dict[str, Any] = {"script": script}
        if page is not None:
            body["page"] = page
        if bundle_id is not None:
            body["bundleId"] = bundle_id
        return self._d._post_json("/webinspector/eval", json=body)


class UiApp:
    """App lifecycle over the UI (WDA/DeviceKit) backend (``/ui/app/*``)."""

    def __init__(self, ui: "Ui") -> None:
        self._ui = ui

    def launch(self, bundle_id: str, **backend: Any) -> Dict[str, Any]:
        """Launch an app via the UI backend (``POST /ui/app/launch``)."""
        return self._ui._post("/ui/app/launch", backend, json={"bundleId": bundle_id})

    def terminate(self, bundle_id: str, **backend: Any) -> Dict[str, Any]:
        """Terminate an app via the UI backend (``POST /ui/app/terminate``)."""
        return self._ui._post("/ui/app/terminate", backend, json={"bundleId": bundle_id})

    def foreground(self, **backend: Any) -> Dict[str, Any]:
        """Bring the active app to the foreground (``POST /ui/app/foreground``)."""
        return self._ui._post("/ui/app/foreground", backend)


class Ui:
    """UI automation proxied to a forwarded WDA/DeviceKit backend (``/ui/*``).

    Every method accepts ``backend=``, ``wda_url=`` and ``timeout=`` keyword
    arguments selecting/locating the backend (mapped to ``?backend/?wdaUrl/?timeout``).
    The backend must already be running (bring up WDA via ``jobs.runwda`` +
    ``jobs.forward`` first).
    """

    def __init__(self, device: "Device") -> None:
        self._d = device
        self.app = UiApp(self)

    def _params(self, backend: Dict[str, Any], extra: Optional[Dict[str, Any]] = None) -> Any:
        return _ui_backend_params(
            backend.get("backend"), backend.get("wda_url"), backend.get("timeout"), extra
        )

    def _post(
        self, suffix: str, backend: Dict[str, Any], *, json: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        return self._d._post_json(suffix, params=self._params(backend), json=json)

    def _get(self, suffix: str, backend: Dict[str, Any]) -> Any:
        return self._d._get_json(suffix, params=self._params(backend))

    # -- gestures -----------------------------------------------------------
    def tap(self, x: int, y: int, **backend: Any) -> Dict[str, Any]:
        """Tap at ``(x, y)`` (``POST /ui/tap``)."""
        return self._post("/ui/tap", backend, json={"x": x, "y": y})

    def swipe(
        self, x1: int, y1: int, x2: int, y2: int, *, duration: Optional[float] = None, **backend: Any
    ) -> Dict[str, Any]:
        """Swipe from ``(x1, y1)`` to ``(x2, y2)`` (``POST /ui/swipe``)."""
        body: Dict[str, Any] = {"x1": x1, "y1": y1, "x2": x2, "y2": y2}
        if duration is not None:
            body["duration"] = duration
        return self._post("/ui/swipe", backend, json=body)

    def long_press(
        self, x: int, y: int, *, duration: Optional[float] = None, **backend: Any
    ) -> Dict[str, Any]:
        """Long-press at ``(x, y)`` (``POST /ui/longpress``)."""
        body: Dict[str, Any] = {"x": x, "y": y}
        if duration is not None:
            body["duration"] = duration
        return self._post("/ui/longpress", backend, json=body)

    def type(self, text: str, **backend: Any) -> Dict[str, Any]:
        """Type ``text`` into the focused field (``POST /ui/type``)."""
        return self._post("/ui/type", backend, json={"text": text})

    def button(self, name: str, **backend: Any) -> Dict[str, Any]:
        """Press a hardware/software button by name (``POST /ui/button``)."""
        return self._post("/ui/button", backend, json={"name": name})

    def api(self, request: Dict[str, Any], **backend: Any) -> Dict[str, Any]:
        """Send a raw request to the backend automation API (``POST /ui/api``)."""
        return self._post("/ui/api", backend, json=request)

    # -- queries ------------------------------------------------------------
    def screenshot(self, **backend: Any) -> bytes:
        """Capture a PNG screenshot via the UI backend (``GET /ui/screenshot``)."""
        return self._d._get_bytes("/ui/screenshot", params=self._params(backend))

    def source(self, **backend: Any) -> str:
        """Get the UI source tree as XML text (``GET /ui/source``)."""
        return self._d._get_text("/ui/source", params=self._params(backend))

    def size(self, **backend: Any) -> Dict[str, Any]:
        """Get the screen size (``GET /ui/size``)."""
        return self._get("/ui/size", backend) or {}

    def orientation(self, **backend: Any) -> Dict[str, Any]:
        """Get the current orientation (``GET /ui/orientation``)."""
        return self._get("/ui/orientation", backend) or {}

    def set_orientation(self, orientation: str, **backend: Any) -> Dict[str, Any]:
        """Set the orientation (``PUT /ui/orientation``)."""
        return self._d._request_json(
            "put", "/ui/orientation", params=self._params(backend),
            json={"orientation": orientation},
        )

    def status(self, **backend: Any) -> Dict[str, Any]:
        """Get the backend status (``GET /ui/status``)."""
        return self._get("/ui/status", backend) or {}

    def stream(
        self,
        *,
        codec: Optional[str] = None,
        backend: Optional[str] = None,
        wda_url: Optional[str] = None,
        timeout: Optional[Union[int, float]] = None,
        fps: Optional[int] = None,
        quality: Optional[int] = None,
        scale: Optional[float] = None,
        bitrate: Optional[int] = None,
    ) -> _SyncBytesStream:
        """Open the live UI video byte stream (``GET /ui/stream``).

        Yields raw bytes chunks (MJPEG, or H.264 with ``codec="h264"``). Binary,
        not SSE. Cancelable via ``close()`` / leaving the ``with`` block.
        """
        extra: Dict[str, Any] = {}
        for key, val in (
            ("codec", codec), ("fps", fps), ("quality", quality),
            ("scale", scale), ("bitrate", bitrate),
        ):
            if val is not None:
                extra[key] = val
        params = _ui_backend_params(backend, wda_url, timeout, extra)
        return self._d._bytes_stream("/ui/stream", params=params)


class Sign:
    """Host-scoped codesigning operations (``/sign/*``, device-free)."""

    def __init__(self, client: "IosClient") -> None:
        self._c = client

    def certificate(
        self,
        asc_private_key: BytesLike,
        asc_key_id: str,
        asc_issuer_id: str,
        *,
        revoke_existing: Optional[bool] = None,
        p12_password: Optional[str] = None,
    ) -> bytes:
        """Create an Apple dev certificate as a P12 (``POST /sign/certificate``) -> bytes."""
        data: Dict[str, str] = {"asc-key-id": asc_key_id, "asc-issuer-id": asc_issuer_id}
        if revoke_existing is not None:
            data["revoke-existing"] = _bool_param(revoke_existing)
        if p12_password is not None:
            data["p12password"] = p12_password
        return self._c._multipart_bytes(
            "/sign/certificate", {"asc-private-key": (asc_private_key, "AuthKey.p8")}, data
        )

    def provision(
        self,
        asc_private_key: BytesLike,
        asc_key_id: str,
        asc_issuer_id: str,
        bundle_id: str,
        udid: str,
        *,
        bundle_name: Optional[str] = None,
        profile_name: Optional[str] = None,
        device_name: Optional[str] = None,
        certificate_id: Optional[str] = None,
        revoke_existing: Optional[bool] = None,
        p12_password: Optional[str] = None,
    ) -> Dict[str, Any]:
        """Create a provisioning profile (``POST /sign/provision``) -> JSON envelope."""
        data: Dict[str, str] = {
            "asc-key-id": asc_key_id, "asc-issuer-id": asc_issuer_id,
            "bundleid": bundle_id, "udid": udid,
        }
        for key, val in (
            ("bundlename", bundle_name), ("profilename", profile_name),
            ("devicename", device_name), ("certificate-id", certificate_id),
            ("p12password", p12_password),
        ):
            if val is not None:
                data[key] = val
        if revoke_existing is not None:
            data["revoke-existing"] = _bool_param(revoke_existing)
        return self._c._multipart(
            "post", "/sign/provision",
            {"asc-private-key": (asc_private_key, "AuthKey.p8")}, data,
        )

    def app(
        self,
        ipa: BytesLike,
        p12file: BytesLike,
        profile: BytesLike,
        *,
        p12_password: Optional[str] = None,
        bundle_id: Optional[str] = None,
    ) -> bytes:
        """Resign an app/IPA and return the signed IPA (``POST /sign/app``) -> bytes."""
        data: Dict[str, str] = {}
        if p12_password is not None:
            data["p12password"] = p12_password
        if bundle_id is not None:
            data["bundleid"] = bundle_id
        return self._c._multipart_bytes(
            "/sign/app",
            {
                "ipa": (ipa, "app.ipa"),
                "p12file": (p12file, "identity.p12"),
                "profile": (profile, "profile.mobileprovision"),
            },
            data,
        )


class Prepare:
    """Host-scoped provisioning helpers (``/prepare/*``, device-free)."""

    def __init__(self, client: "IosClient") -> None:
        self._c = client

    def create_cert(self) -> Dict[str, Any]:
        """Generate a self-signed supervision identity (``POST /prepare/create-cert``)."""
        return self._c._request_json("post", "/prepare/create-cert")

    def skip_options(self) -> Dict[str, Any]:
        """List setup-pane skip options (``GET /prepare/skip-options``)."""
        return self._c._get_json("/prepare/skip-options") or {}


def _fsync_params(path: Optional[str], bundle_id: Optional[str]) -> Optional[Dict[str, Any]]:
    params: Dict[str, Any] = {}
    if path is not None:
        params["path"] = path
    if bundle_id is not None:
        params["bundleID"] = bundle_id
    return params or None


class Devices:
    """Fleet-level device operations."""

    def __init__(self, client: "IosClient") -> None:
        self._c = client

    def list(self) -> Dict[str, Any]:
        """List attached devices (``GET /list``). Returns the DeviceList envelope."""
        resp = self._c._http.get(f"{API_PREFIX}/list")
        raise_for_status(resp)
        return json_or_none(resp) or {}

    def udids(self) -> List[str]:
        """Convenience: the udids (``properties.serialNumber``) of attached devices."""
        return _extract_udids(self.list())


def _extract_udids(device_list: Dict[str, Any]) -> List[str]:
    out: List[str] = []
    for entry in device_list.get("deviceList", []) or []:
        serial = (entry.get("properties") or {}).get("serialNumber")
        if serial:
            out.append(serial)
    return out


class Tunnels:
    """userspace-tunnel (RemoteXPC) management (iOS 17+)."""

    def __init__(self, client: "IosClient") -> None:
        self._c = client

    def list(self) -> List[Dict[str, Any]]:
        """List active tunnels (``GET /tunnels``)."""
        return self._c._get_json("/tunnels") or []

    def delete(self, udid: str) -> Dict[str, Any]:
        """Stop the tunnel for ``udid`` (``DELETE /tunnels/{udid}``)."""
        return self._c._request_json("delete", f"/tunnels/{udid}")

    def refresh(self, udid: str) -> Dict[str, Any]:
        """Refresh the tunnel for ``udid`` (``POST /tunnels/{udid}/refresh``)."""
        return self._c._request_json("post", f"/tunnels/{udid}/refresh")

    def shutdown_agent(self) -> Dict[str, Any]:
        """Shut down the whole tunnel agent (``POST /tunnel-agent/shutdown``)."""
        return self._c._request_json("post", "/tunnel-agent/shutdown")


class Device:
    """Operations scoped to a single device udid."""

    def __init__(self, client: "IosClient", udid: str) -> None:
        self._c = client
        self.udid = udid
        self.apps = Apps(self)
        self.wda = Wda(self)
        self.files = Files(self)
        self.crashes = Crashes(self)
        self.jobs = Jobs(self)
        self.fsync = Fsync(self)
        self.webinspector = WebInspector(self)
        self.ui = Ui(self)

    # -- internal request helpers -------------------------------------------
    def _url(self, suffix: str) -> str:
        return f"{API_PREFIX}/device/{self.udid}{suffix}"

    def _request_json(self, method: str, suffix: str, **kwargs: Any) -> Dict[str, Any]:
        resp = self._c._http.request(method, self._url(suffix), **kwargs)
        raise_for_status(resp)
        return json_or_none(resp) or {}

    def _get_json(self, suffix: str, **kwargs: Any) -> Any:
        resp = self._c._http.get(self._url(suffix), **kwargs)
        raise_for_status(resp)
        return json_or_none(resp)

    def _get_bytes(self, suffix: str, **kwargs: Any) -> bytes:
        resp = self._c._http.get(self._url(suffix), **kwargs)
        raise_for_status(resp)
        return resp.content

    def _get_text(self, suffix: str, **kwargs: Any) -> str:
        resp = self._c._http.get(self._url(suffix), **kwargs)
        raise_for_status(resp)
        return resp.text

    def _bytes_stream(
        self, suffix: str, *, params: Optional[Dict[str, Any]] = None
    ) -> _SyncBytesStream:
        cm = self._c._http.stream(
            "GET", self._url(suffix), params=params, timeout=_STREAM_TIMEOUT
        )
        return _SyncBytesStream(cm)

    def _post_json(self, suffix: str, **kwargs: Any) -> Dict[str, Any]:
        return self._request_json("post", suffix, **kwargs)

    def _put_json(self, suffix: str, **kwargs: Any) -> Dict[str, Any]:
        return self._request_json("put", suffix, **kwargs)

    def _multipart(
        self,
        method: str,
        suffix: str,
        file_fields: Dict[str, Any],
        data_fields: Optional[Dict[str, Any]] = None,
    ) -> Dict[str, Any]:
        """Send a multipart request.

        ``file_fields`` maps field -> ``(source, default_name)``; ``data_fields``
        maps field -> a string (or list of strings for repeated fields).
        """
        sources = {name: src for name, (src, _) in file_fields.items()}
        files = {
            name: _file_tuple(src, default)
            for name, (src, default) in file_fields.items()
        }
        try:
            return self._request_json(
                method, suffix, files=files, data=data_fields or None
            )
        finally:
            _close_multipart(files, sources)

    def _multipart_bytes(
        self,
        suffix: str,
        file_fields: Dict[str, Any],
        data_fields: Optional[Dict[str, Any]] = None,
    ) -> bytes:
        """Like :meth:`_multipart` but returns the raw response bytes (POST)."""
        sources = {name: src for name, (src, _) in file_fields.items()}
        files = {
            name: _file_tuple(src, default)
            for name, (src, default) in file_fields.items()
        }
        try:
            resp = self._c._http.post(
                self._url(suffix), files=files, data=data_fields or None
            )
            raise_for_status(resp)
            return resp.content
        finally:
            _close_multipart(files, sources)

    def _post_multipart(self, suffix: str, file_fields: Dict[str, Any]) -> Dict[str, Any]:
        return self._multipart("post", suffix, file_fields)

    # -- unary operations ----------------------------------------------------
    def info(self) -> Dict[str, Any]:
        """Get device info (lockdown + instruments values) (``GET /info``)."""
        return self._get_json("/info") or {}

    def screenshot(self) -> bytes:
        """Capture a PNG screenshot as raw bytes (``GET /screenshot``)."""
        return self._get_bytes("/screenshot")

    def activate(self) -> Dict[str, Any]:
        """Activate the device (``POST /activate``)."""
        return self._post_json("/activate")

    def pair(
        self,
        *,
        supervised: bool = False,
        p12file: Optional[BytesLike] = None,
        supervision_password: Optional[str] = None,
    ) -> Dict[str, Any]:
        """Pair the device (``POST /pair``).

        For supervised pairing pass ``supervised=True``, the ``p12file``
        supervision identity, and optionally ``supervision_password``.
        """
        params = {"supervised": _bool_param(supervised)}
        headers: Dict[str, str] = {}
        if supervision_password is not None:
            headers["Supervision-Password"] = supervision_password
        if p12file is None:
            return self._post_json("/pair", params=params, headers=headers or None)
        files = {"p12file": _file_tuple(p12file, "identity.p12")}
        try:
            return self._post_json(
                "/pair", params=params, headers=headers or None, files=files
            )
        finally:
            _close_multipart(files, {"p12file": p12file})

    # -- device info ---------------------------------------------------------
    def device_name(self) -> Dict[str, Any]:
        """Get the device name (``GET /devicename``)."""
        return self._get_json("/devicename") or {}

    def date(self) -> Dict[str, Any]:
        """Get the device date/time (``GET /date``)."""
        return self._get_json("/date") or {}

    def battery(self) -> Dict[str, Any]:
        """Get battery info (``GET /battery``)."""
        return self._get_json("/battery") or {}

    def diagnostics(self) -> Dict[str, Any]:
        """Get IORegistry diagnostics (``GET /diagnostics``)."""
        return self._get_json("/diagnostics") or {}

    def mobilegestalt(self, keys: Optional[Sequence[str]] = None) -> Dict[str, Any]:
        """Query MobileGestalt values by key (``GET /mobilegestalt``)."""
        params = {"key": list(keys)} if keys else None
        return self._get_json("/mobilegestalt", params=params) or {}

    def processes(self, *, apps: Optional[bool] = None) -> List[Dict[str, Any]]:
        """List running processes (``GET /processes``)."""
        params = {"apps": _bool_param(apps)} if apps is not None else None
        return self._get_json("/processes", params=params) or []

    def lockdown(self, *, domain: Optional[str] = None) -> Dict[str, Any]:
        """Read lockdown values, optionally scoped to a ``domain`` (``GET /lockdown``)."""
        params = {"domain": domain} if domain is not None else None
        return self._get_json("/lockdown", params=params) or {}

    # -- diagnostics / network (v3) -----------------------------------------
    def disk_space(self) -> Dict[str, Any]:
        """Get filesystem disk-space info (``GET /diskspace``)."""
        return self._get_json("/diskspace") or {}

    def ip(self) -> Dict[str, Any]:
        """Get the device's IP address info (``GET /ip``)."""
        return self._get_json("/ip") or {}

    def rsd(self) -> Dict[str, Any]:
        """Get RemoteServiceDiscovery info (``GET /rsd``; requires a tunnel)."""
        return self._get_json("/rsd") or {}

    def battery_registry(self) -> Dict[str, Any]:
        """Get the raw battery IORegistry dump (``GET /battery/registry``)."""
        return self._get_json("/battery/registry") or {}

    # -- accessibility (v3) --------------------------------------------------
    def voice_over(self) -> Dict[str, Any]:
        """Get VoiceOver enabled state (``GET /voiceover``)."""
        return self._get_json("/voiceover") or {}

    def set_voice_over(self, enabled: bool) -> Dict[str, Any]:
        """Enable/disable VoiceOver (``PUT /voiceover``)."""
        return self._put_json("/voiceover", json={"enabled": enabled})

    def zoom(self) -> Dict[str, Any]:
        """Get Zoom enabled state (``GET /zoom``)."""
        return self._get_json("/zoom") or {}

    def set_zoom(self, enabled: bool) -> Dict[str, Any]:
        """Enable/disable Zoom (``PUT /zoom``)."""
        return self._put_json("/zoom", json={"enabled": enabled})

    def ax(self) -> Dict[str, Any]:
        """Get an accessibility snapshot of the screen (``GET /ax``)."""
        return self._get_json("/ax") or {}

    def ax_audit(self, *, timeout: Optional[Union[int, float]] = None) -> Dict[str, Any]:
        """Run an accessibility audit (``POST /ax/audit``)."""
        params = {"timeout": timeout} if timeout is not None else None
        return self._post_json("/ax/audit", params=params)

    def set_location_gpx(self, gpx: BytesLike) -> Dict[str, Any]:
        """Replay a GPX track as the simulated location (``PUT /setlocation/gpx``)."""
        return self._multipart("put", "/setlocation/gpx", {"gpx": (gpx, "track.gpx")})

    def cloud_config(self) -> Dict[str, Any]:
        """Get the device cloud configuration (``GET /cloudconfig``)."""
        return self._get_json("/cloudconfig") or {}

    def prepare(
        self,
        cert: BytesLike,
        *,
        p12_password: Optional[str] = None,
        skip: Optional[Sequence[str]] = None,
        org_name: Optional[str] = None,
        locale: Optional[str] = None,
        lang: Optional[str] = None,
    ) -> Dict[str, Any]:
        """Prepare/supervise the device with a certificate (``POST /prepare``, multipart)."""
        data: Dict[str, Any] = {}
        if p12_password is not None:
            data["p12password"] = p12_password
        if skip:
            data["skip"] = list(skip)
        if org_name is not None:
            data["orgname"] = org_name
        if locale is not None:
            data["locale"] = locale
        if lang is not None:
            data["lang"] = lang
        return self._multipart("post", "/prepare", {"cert": (cert, "supervision.p12")}, data or None)

    # -- management ----------------------------------------------------------
    def reboot(self) -> Dict[str, Any]:
        """Reboot the device (``POST /reboot``)."""
        return self._post_json("/reboot")

    def shutdown(self) -> Dict[str, Any]:
        """Shut down the device (``POST /shutdown``)."""
        return self._post_json("/shutdown")

    def erase(self, *, confirm: bool = False) -> Dict[str, Any]:
        """Erase the device (``POST /erase``). Requires ``confirm=True``."""
        return self._post_json("/erase", params={"confirm": _bool_param(confirm)})

    def devmode(self) -> Dict[str, Any]:
        """Get developer-mode state (``GET /devmode``)."""
        return self._get_json("/devmode") or {}

    def set_devmode(
        self, action: str = "enable", *, enable_post_restart: Optional[bool] = None
    ) -> Dict[str, Any]:
        """Set developer mode (``POST /devmode``). ``action`` is ``enable`` or ``reveal``."""
        body: Dict[str, Any] = {"action": action}
        if enable_post_restart is not None:
            body["enablePostRestart"] = enable_post_restart
        return self._post_json("/devmode", json=body)

    def lang(self) -> Dict[str, Any]:
        """Get language/locale configuration (``GET /lang``)."""
        return self._get_json("/lang") or {}

    def set_lang(
        self, *, language: Optional[str] = None, locale: Optional[str] = None
    ) -> Dict[str, Any]:
        """Set language and/or locale (``PUT /lang``)."""
        body: Dict[str, Any] = {}
        if language is not None:
            body["language"] = language
        if locale is not None:
            body["locale"] = locale
        return self._put_json("/lang", json=body)

    def memlimitoff(self, process: Optional[str] = None) -> Dict[str, Any]:
        """Waive the memory limit for a process (``POST /memlimitoff``)."""
        body = {"process": process} if process is not None else {}
        params = {"process": process} if process is not None else None
        return self._post_json("/memlimitoff", json=body, params=params)

    # -- media ---------------------------------------------------------------
    def get_wallpaper(self) -> bytes:
        """Get the current wallpaper PNG bytes (``GET /wallpaper``)."""
        return self._get_bytes("/wallpaper")

    def set_wallpaper(
        self,
        image: BytesLike,
        p12: BytesLike,
        *,
        password: Optional[str] = None,
        screen: Optional[str] = None,
    ) -> Dict[str, Any]:
        """Set the wallpaper (supervised) (``PUT /wallpaper``)."""
        data: Dict[str, str] = {}
        if password is not None:
            data["password"] = password
        if screen is not None:
            data["screen"] = screen
        return self._multipart(
            "put",
            "/wallpaper",
            {"image": (image, "wallpaper.png"), "p12": (p12, "identity.p12")},
            data or None,
        )

    def get_icon_layout(self) -> Dict[str, Any]:
        """Get the SpringBoard icon layout (``GET /icon-layout``)."""
        return self._get_json("/icon-layout") or {}

    def set_icon_layout(self, layout: Dict[str, Any]) -> Dict[str, Any]:
        """Set the SpringBoard icon layout (``PUT /icon-layout``)."""
        return self._put_json("/icon-layout", json=layout)

    def get_pasteboard(self) -> Dict[str, Any]:
        """Read the device pasteboard (``GET /pasteboard``)."""
        return self._get_json("/pasteboard") or {}

    def set_pasteboard(self, text: str) -> Dict[str, Any]:
        """Write text to the device pasteboard (``PUT /pasteboard``)."""
        return self._put_json(
            "/pasteboard",
            content=text.encode("utf-8"),
            headers={"Content-Type": "text/plain"},
        )

    # -- settings ------------------------------------------------------------
    def assistive_touch(self) -> Dict[str, Any]:
        """Get AssistiveTouch state (``GET /assistivetouch``)."""
        return self._get_json("/assistivetouch") or {}

    def set_assistive_touch(self, enabled: bool) -> Dict[str, Any]:
        """Enable/disable AssistiveTouch (``PUT /assistivetouch``)."""
        return self._put_json("/assistivetouch", json={"enabled": enabled})

    def time_format(self) -> Dict[str, Any]:
        """Get the 24-hour clock setting (``GET /timeformat``)."""
        return self._get_json("/timeformat") or {}

    def set_time_format(self, uses_24_hour: bool) -> Dict[str, Any]:
        """Set the 24-hour clock setting (``PUT /timeformat``)."""
        return self._put_json("/timeformat", json={"uses24Hour": uses_24_hour})

    def set_wifi(
        self, ssid: str, *, password: Optional[str] = None, enc_type: Optional[str] = None
    ) -> Dict[str, Any]:
        """Configure a Wi-Fi network (``PUT /wifi``)."""
        body: Dict[str, Any] = {"ssid": ssid}
        if password is not None:
            body["password"] = password
        if enc_type is not None:
            body["encType"] = enc_type
        return self._put_json("/wifi", json=body)

    def remove_wifi(self, ssid: str) -> Dict[str, Any]:
        """Forget a Wi-Fi network (``DELETE /wifi``)."""
        return self._request_json("delete", "/wifi", params={"ssid": ssid})

    # -- mdm -----------------------------------------------------------------
    def security_info(
        self, p12: BytesLike, *, password: Optional[str] = None
    ) -> Dict[str, Any]:
        """Query MDM security info (``POST /mdm/security-info``)."""
        data = {"password": password} if password is not None else None
        return self._multipart(
            "post", "/mdm/security-info", {"p12": (p12, "identity.p12")}, data
        )

    def fetch_unlock_token(
        self, p12: BytesLike, *, password: Optional[str] = None
    ) -> Dict[str, Any]:
        """Fetch the escrow unlock token (``POST /mdm/fetch-unlock-token``)."""
        data = {"password": password} if password is not None else None
        return self._multipart(
            "post", "/mdm/fetch-unlock-token", {"p12": (p12, "identity.p12")}, data
        )

    def clear_passcode(
        self, p12: BytesLike, token: str, *, password: Optional[str] = None
    ) -> Dict[str, Any]:
        """Clear the device passcode via MDM (``POST /mdm/clear-passcode``)."""
        data = {"token": token}
        if password is not None:
            data["password"] = password
        return self._multipart(
            "post", "/mdm/clear-passcode", {"p12": (p12, "identity.p12")}, data
        )

    def clear_screen_time_password(
        self, p12: BytesLike, *, password: Optional[str] = None
    ) -> Dict[str, Any]:
        """Clear the Screen Time password via MDM (``POST /mdm/clear-screen-time-password``)."""
        data = {"password": password} if password is not None else None
        return self._multipart(
            "post",
            "/mdm/clear-screen-time-password",
            {"p12": (p12, "identity.p12")},
            data,
        )

    # -- proxy ---------------------------------------------------------------
    def set_http_proxy(
        self,
        host: str,
        port: Union[int, str],
        *,
        user: Optional[str] = None,
        password: Optional[str] = None,
        p12: Optional[BytesLike] = None,
        p12_password: Optional[str] = None,
    ) -> Dict[str, Any]:
        """Set a global HTTP proxy (supervised) (``PUT /httpproxy``)."""
        data: Dict[str, str] = {"host": host, "port": str(port)}
        if user is not None:
            data["user"] = user
        if password is not None:
            data["pass"] = password
        if p12_password is not None:
            data["password"] = p12_password
        file_fields = {"p12": (p12, "identity.p12")} if p12 is not None else {}
        return self._multipart("put", "/httpproxy", file_fields, data)

    def remove_http_proxy(self) -> Dict[str, Any]:
        """Remove the global HTTP proxy (``DELETE /httpproxy``)."""
        return self._request_json("delete", "/httpproxy")

    # -- conditions / images / profiles -------------------------------------
    def conditions(self) -> List[Dict[str, Any]]:
        """List available condition profile types (``GET /conditions``)."""
        return self._get_json("/conditions") or []

    def enable_condition(self, profile_type_id: str, profile_id: str) -> Dict[str, Any]:
        """Enable a device condition (``PUT /enable-condition``)."""
        return self._request_json(
            "put",
            "/enable-condition",
            params={"profileTypeID": profile_type_id, "profileID": profile_id},
        )

    def disable_condition(self) -> Dict[str, Any]:
        """Disable the active device condition (``POST /disable-condition``)."""
        return self._post_json("/disable-condition")

    def images(self) -> List[str]:
        """List available developer disk images on the server (``GET /image``)."""
        return self._get_json("/image") or []

    def mounted_images(self) -> Dict[str, Any]:
        """List mounted developer disk images (``GET /image/list``)."""
        return self._get_json("/image/list") or {}

    def install_image(
        self,
        image: Optional[BytesLike] = None,
        *,
        auto: bool = False,
        basedir: Optional[str] = None,
    ) -> Dict[str, Any]:
        """Mount a developer disk image (``PUT /image``).

        Either pass raw ``image`` bytes/path, or set ``auto=True`` (optionally
        with ``basedir``) to have the server auto-resolve the image.
        """
        params: Dict[str, Any] = {}
        if auto:
            params["auto"] = "true"
        if basedir is not None:
            params["basedir"] = basedir
        content = _resolve_body(image) if image is not None else None
        return self._request_json(
            "put",
            "/image",
            params=params or None,
            content=content,
            headers={"Content-Type": "application/octet-stream"} if content else None,
        )

    def unmount_image(self) -> Dict[str, Any]:
        """Unmount the developer disk image (``DELETE /image``)."""
        return self._request_json("delete", "/image")

    def profiles(self) -> Dict[str, Any]:
        """List installed configuration profiles (``GET /profiles``)."""
        return self._get_json("/profiles") or {}

    def add_profile(
        self,
        profile: BytesLike,
        *,
        p12: Optional[BytesLike] = None,
        password: Optional[str] = None,
    ) -> Dict[str, Any]:
        """Install a ``.mobileconfig`` profile (``POST /profiles``)."""
        data = {"password": password} if password is not None else None
        file_fields = {"profile": (profile, "profile.mobileconfig")}
        if p12 is not None:
            file_fields["p12"] = (p12, "identity.p12")
        return self._multipart("post", "/profiles", file_fields, data)

    def remove_profile(self, name: str) -> Dict[str, Any]:
        """Remove an installed profile by identifier (``DELETE /profiles/{name}``)."""
        return self._request_json("delete", f"/profiles/{name}")

    def reset_accessibility(self) -> Dict[str, Any]:
        """Reset accessibility settings (``POST /resetaccessibility``)."""
        return self._post_json("/resetaccessibility")

    def reset_location(self) -> Dict[str, Any]:
        """Reset the simulated location (``POST /resetlocation``)."""
        return self._post_json("/resetlocation")

    def set_location(self, latitude: float, longitude: float) -> Dict[str, Any]:
        """Set a simulated GPS location (``PUT /setlocation``)."""
        return self._request_json(
            "put",
            "/setlocation",
            params={"latitude": str(latitude), "longitude": str(longitude)},
        )

    # -- streaming operations ------------------------------------------------
    def _stream(self, suffix: str, *, params: Optional[Dict[str, Any]] = None,
                include_heartbeats: bool = False) -> _SyncStream:
        cm = self._c._http.stream(
            "GET", self._url(suffix), params=params, timeout=_STREAM_TIMEOUT
        )
        return _SyncStream(cm, include_heartbeats)

    def syslog(self, *, include_heartbeats: bool = False) -> _SyncStream:
        """Stream syslog messages (``GET /syslog``) as typed events."""
        return self._stream("/syslog", include_heartbeats=include_heartbeats)

    def notifications(self, *, include_heartbeats: bool = False) -> _SyncStream:
        """Stream app-state notifications (``GET /notifications``)."""
        return self._stream("/notifications", include_heartbeats=include_heartbeats)

    def ostrace(
        self,
        *,
        pid: Optional[int] = None,
        level: Optional[str] = None,
        subsystem: Optional[str] = None,
        match: Optional[str] = None,
        exclude: Optional[str] = None,
        include_heartbeats: bool = False,
    ) -> _SyncStream:
        """Stream os_trace entries (``GET /ostrace``) with optional AND filters."""
        params = _ostrace_params(pid, level, subsystem, match, exclude)
        return self._stream("/ostrace", params=params, include_heartbeats=include_heartbeats)

    def listen(self, *, include_heartbeats: bool = False) -> _SyncStream:
        """Stream device attach/detach/pair events (``GET /listen``)."""
        return self._stream("/listen", include_heartbeats=include_heartbeats)

    def sysmontap(self, *, include_heartbeats: bool = False) -> _SyncStream:
        """Stream CPU-usage samples (``GET /sysmontap``) as typed events."""
        return self._stream("/sysmontap", include_heartbeats=include_heartbeats)

    # -- binary streams (v3; raw bytes, NOT SSE) -----------------------------
    def screenshot_stream(self, *, quality: Optional[int] = None) -> _SyncBytesStream:
        """Stream MJPEG screenshots as raw byte chunks (``GET /screenshot/stream``)."""
        params = {"quality": quality} if quality is not None else None
        return self._bytes_stream("/screenshot/stream", params=params)

    def pcap(self, *, timeout: Optional[Union[int, float]] = None) -> _SyncBytesStream:
        """Stream a live libpcap packet capture as raw byte chunks (``GET /pcap``)."""
        params = {"timeout": timeout} if timeout is not None else None
        return self._bytes_stream("/pcap", params=params)


def _ui_backend_params(
    backend: Optional[str],
    wda_url: Optional[str],
    timeout: Optional[Union[int, float]],
    extra: Optional[Dict[str, Any]] = None,
) -> Optional[Dict[str, Any]]:
    """Build the shared ``?backend=/?wdaUrl=/?timeout=`` query for the UI group."""
    params: Dict[str, Any] = dict(extra or {})
    if backend is not None:
        params["backend"] = backend
    if wda_url is not None:
        params["wdaUrl"] = wda_url
    if timeout is not None:
        params["timeout"] = timeout
    return params or None


def _ostrace_params(
    pid: Optional[int],
    level: Optional[str],
    subsystem: Optional[str],
    match: Optional[str],
    exclude: Optional[str],
) -> Optional[Dict[str, Any]]:
    params: Dict[str, Any] = {}
    if pid is not None:
        params["pid"] = pid
    if level is not None:
        params["level"] = level
    if subsystem is not None:
        params["subsystem"] = subsystem
    if match is not None:
        params["match"] = match
    if exclude is not None:
        params["exclude"] = exclude
    return params or None


class IosClient:
    """Synchronous go-ios REST client.

    Args:
        base_url: Server base URL (default ``http://localhost:60105``).
        api_key: Bearer token; sent as ``Authorization: Bearer <key>`` when set.
            Optional (a server started with ``--disable-auth`` needs none), but
            strongly encouraged.
        timeout: Per-request timeout in seconds for non-streaming calls.
        verify: TLS verification (bool / CA path / SSLContext), forwarded to httpx.
        http_client: Bring your own configured ``httpx.Client`` (overrides the
            other transport options).
    """

    def __init__(
        self,
        base_url: str = DEFAULT_BASE_URL,
        *,
        api_key: Optional[str] = None,
        timeout: float = 30.0,
        verify: Any = True,
        headers: Optional[Dict[str, str]] = None,
        http_client: Optional[httpx.Client] = None,
    ) -> None:
        self.base_url = base_url.rstrip("/")
        self.api_key = api_key
        if http_client is not None:
            self._http = http_client
            self._owns_http = False
        else:
            self._http = httpx.Client(
                base_url=self.base_url,
                headers=build_headers(api_key, headers),
                timeout=timeout,
                verify=verify,
            )
            self._owns_http = True
        self.devices = Devices(self)
        self.tunnels = Tunnels(self)
        self.sign = Sign(self)
        self.prepare = Prepare(self)

    # -- fleet-level request helpers ----------------------------------------
    def _get_json(self, suffix: str, **kwargs: Any) -> Any:
        resp = self._http.get(f"{API_PREFIX}{suffix}", **kwargs)
        raise_for_status(resp)
        return json_or_none(resp)

    def _request_json(self, method: str, suffix: str, **kwargs: Any) -> Dict[str, Any]:
        resp = self._http.request(method, f"{API_PREFIX}{suffix}", **kwargs)
        raise_for_status(resp)
        return json_or_none(resp) or {}

    def _multipart(
        self,
        method: str,
        suffix: str,
        file_fields: Dict[str, Any],
        data_fields: Optional[Dict[str, Any]] = None,
    ) -> Dict[str, Any]:
        """Host-scoped multipart request. See :meth:`Device._multipart`."""
        sources = {name: src for name, (src, _) in file_fields.items()}
        files = {name: _file_tuple(src, d) for name, (src, d) in file_fields.items()}
        try:
            return self._request_json(method, suffix, files=files, data=data_fields or None)
        finally:
            _close_multipart(files, sources)

    def _multipart_bytes(
        self,
        suffix: str,
        file_fields: Dict[str, Any],
        data_fields: Optional[Dict[str, Any]] = None,
    ) -> bytes:
        """Host-scoped multipart POST returning raw response bytes."""
        sources = {name: src for name, (src, _) in file_fields.items()}
        files = {name: _file_tuple(src, d) for name, (src, d) in file_fields.items()}
        try:
            resp = self._http.post(
                f"{API_PREFIX}{suffix}", files=files, data=data_fields or None
            )
            raise_for_status(resp)
            return resp.content
        finally:
            _close_multipart(files, sources)

    def device(self, udid: str) -> Device:
        """Return a :class:`Device` handle for ``udid``."""
        return Device(self, udid)

    def close(self) -> None:
        if self._owns_http:
            self._http.close()

    def __enter__(self) -> "IosClient":
        return self

    def __exit__(self, *exc: Any) -> None:
        self.close()


# Re-export the typed event classes for convenience.
AppStateNotification = _events.AppStateNotification
SyslogMessage = _events.SyslogMessage
OsTraceEntry = _events.OsTraceEntry
AttachDetachEvent = _events.AttachDetachEvent
CpuUsageSample = _events.CpuUsageSample
JobLogLine = _events.JobLogLine
Heartbeat = _events.Heartbeat
UnknownEvent = _events.UnknownEvent
