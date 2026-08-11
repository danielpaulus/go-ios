"""Ergonomic asynchronous facade for the go-ios REST API.

Mirrors :mod:`go_ios_sdk.client` (same method names/shape) but every operation is
a coroutine and streaming endpoints are async generators::

    async with AsyncIosClient(base_url=..., api_key=...) as client:
        await client.devices.list()
        dev = client.device(udid)
        await dev.info()
        async for event in dev.syslog():
            ...
        async for sample in dev.sysmontap():
            ...
        async for line in dev.jobs.logs(job_id):
            ...

Streaming methods return an async iterator that is also an async context manager,
so cancelling the ``async for`` (or leaving an ``async with`` block) closes the
underlying HTTP response promptly.
"""

from __future__ import annotations

from typing import Any, AsyncIterator, Dict, List, Optional, Sequence, Union

import httpx

from ._http import API_PREFIX, build_headers, json_or_none, raise_for_status
from .client import (
    _STREAM_TIMEOUT,
    DEFAULT_BASE_URL,
    BytesLike,
    _bool_param,
    _close_multipart,
    _extract_udids,
    _file_tuple,
    _ostrace_params,
    _resolve_body,
)
from .sse import aiter_events


class _AsyncStream:
    """A closeable async iterator over an SSE stream.

    Usable as ``async for ev in stream`` or ``async with stream as s``. The HTTP
    response is released when iteration completes, the ``async with`` exits, or
    the consuming task is cancelled.
    """

    def __init__(self, request_cm: Any, include_heartbeats: bool) -> None:
        self._request_cm = request_cm
        self._response: Optional[httpx.Response] = None
        self._include_heartbeats = include_heartbeats

    async def _open(self) -> httpx.Response:
        if self._response is None:
            self._response = await self._request_cm.__aenter__()
            raise_for_status(self._response)
        return self._response

    async def __aiter__(self) -> AsyncIterator[Any]:
        resp = await self._open()
        try:
            async for event in aiter_events(
                resp.aiter_bytes(), include_heartbeats=self._include_heartbeats
            ):
                yield event
        finally:
            await self.aclose()

    async def __aenter__(self) -> "_AsyncStream":
        await self._open()
        return self

    async def __aexit__(self, *exc: Any) -> None:
        await self.aclose()

    async def aclose(self) -> None:
        if self._response is not None:
            try:
                await self._request_cm.__aexit__(None, None, None)
            finally:
                self._response = None


class AsyncApps:
    def __init__(self, device: "AsyncDevice") -> None:
        self._d = device

    async def list(self) -> List[Dict[str, Any]]:
        return (await self._d._get_json("/apps/")) or []

    async def launch(self, bundle_id: str) -> Dict[str, Any]:
        return await self._d._post_json("/apps/launch", params={"bundleID": bundle_id})

    async def kill(self, bundle_id: str) -> Dict[str, Any]:
        return await self._d._post_json("/apps/kill", params={"bundleID": bundle_id})

    async def install(self, ipa: BytesLike) -> Dict[str, Any]:
        return await self._d._multipart("post", "/apps/install", {"file": (ipa, "app.ipa")})

    async def uninstall(self, bundle_id: str) -> Dict[str, Any]:
        return await self._d._post_json("/apps/uninstall", params={"bundleID": bundle_id})


class AsyncWda:
    def __init__(self, device: "AsyncDevice") -> None:
        self._d = device

    async def create_session(self, config: Dict[str, Any]) -> Dict[str, Any]:
        return await self._d._post_json("/wda/session", json=config)

    async def read_session(self, session_id: str) -> Dict[str, Any]:
        return await self._d._get_json(f"/wda/session/{session_id}")

    async def delete_session(self, session_id: str) -> Dict[str, Any]:
        return await self._d._request_json("delete", f"/wda/session/{session_id}")


class AsyncFiles:
    def __init__(self, device: "AsyncDevice") -> None:
        self._d = device

    async def ls(
        self,
        path: Optional[str] = None,
        *,
        domain: str = "app",
        identifier: Optional[str] = None,
    ) -> Dict[str, Any]:
        params: Dict[str, Any] = {"domain": domain}
        if identifier is not None:
            params["identifier"] = identifier
        if path is not None:
            params["path"] = path
        return (await self._d._get_json("/files", params=params)) or {}

    async def pull(
        self,
        path: str,
        *,
        domain: str = "app",
        identifier: Optional[str] = None,
    ) -> bytes:
        params: Dict[str, Any] = {"domain": domain, "remote": path}
        if identifier is not None:
            params["identifier"] = identifier
        return await self._d._get_bytes("/files/pull", params=params)

    async def push(
        self,
        path: str,
        data: BytesLike,
        *,
        domain: str = "app",
        identifier: Optional[str] = None,
    ) -> Dict[str, Any]:
        params: Dict[str, Any] = {"domain": domain, "remote": path}
        if identifier is not None:
            params["identifier"] = identifier
        return await self._d._request_json(
            "post",
            "/files/push",
            params=params,
            content=_resolve_body(data),
            headers={"Content-Type": "application/octet-stream"},
        )


class AsyncCrashes:
    def __init__(self, device: "AsyncDevice") -> None:
        self._d = device

    async def list(self, pattern: str = "*") -> List[str]:
        return (await self._d._get_json("/crashes", params={"pattern": pattern})) or []

    async def remove(self, pattern: str, *, cwd: str = ".") -> Dict[str, Any]:
        return await self._d._request_json(
            "delete", "/crashes", params={"cwd": cwd, "pattern": pattern}
        )


class AsyncJobs:
    def __init__(self, device: "AsyncDevice") -> None:
        self._d = device

    async def runtest(self, config: Dict[str, Any]) -> Dict[str, Any]:
        return await self._d._post_json("/jobs/runtest", json=config)

    async def runwda(self, config: Dict[str, Any]) -> Dict[str, Any]:
        return await self._d._post_json("/jobs/runwda", json=config)

    async def forward(self, config: Dict[str, Any]) -> Dict[str, Any]:
        return await self._d._post_json("/jobs/forward", json=config)

    async def list(self) -> List[Dict[str, Any]]:
        return (await self._d._get_json("/jobs")) or []

    async def get(self, job_id: str) -> Dict[str, Any]:
        return (await self._d._get_json(f"/jobs/{job_id}")) or {}

    async def delete(self, job_id: str) -> Dict[str, Any]:
        return await self._d._request_json("delete", f"/jobs/{job_id}")

    def logs(self, job_id: str, *, include_heartbeats: bool = False) -> _AsyncStream:
        return self._d._stream(f"/jobs/{job_id}/logs", include_heartbeats=include_heartbeats)


class AsyncDevices:
    def __init__(self, client: "AsyncIosClient") -> None:
        self._c = client

    async def list(self) -> Dict[str, Any]:
        resp = await self._c._http.get(f"{API_PREFIX}/list")
        raise_for_status(resp)
        return json_or_none(resp) or {}

    async def udids(self) -> List[str]:
        """Convenience: the udids (``properties.serialNumber``) of attached devices."""
        return _extract_udids(await self.list())


class AsyncTunnels:
    def __init__(self, client: "AsyncIosClient") -> None:
        self._c = client

    async def list(self) -> List[Dict[str, Any]]:
        return (await self._c._get_json("/tunnels")) or []

    async def delete(self, udid: str) -> Dict[str, Any]:
        return await self._c._request_json("delete", f"/tunnels/{udid}")

    async def refresh(self, udid: str) -> Dict[str, Any]:
        return await self._c._request_json("post", f"/tunnels/{udid}/refresh")

    async def shutdown_agent(self) -> Dict[str, Any]:
        return await self._c._request_json("post", "/tunnel-agent/shutdown")


class AsyncDevice:
    def __init__(self, client: "AsyncIosClient", udid: str) -> None:
        self._c = client
        self.udid = udid
        self.apps = AsyncApps(self)
        self.wda = AsyncWda(self)
        self.files = AsyncFiles(self)
        self.crashes = AsyncCrashes(self)
        self.jobs = AsyncJobs(self)

    def _url(self, suffix: str) -> str:
        return f"{API_PREFIX}/device/{self.udid}{suffix}"

    async def _request_json(self, method: str, suffix: str, **kwargs: Any) -> Dict[str, Any]:
        resp = await self._c._http.request(method, self._url(suffix), **kwargs)
        raise_for_status(resp)
        return json_or_none(resp) or {}

    async def _get_json(self, suffix: str, **kwargs: Any) -> Any:
        resp = await self._c._http.get(self._url(suffix), **kwargs)
        raise_for_status(resp)
        return json_or_none(resp)

    async def _get_bytes(self, suffix: str, **kwargs: Any) -> bytes:
        resp = await self._c._http.get(self._url(suffix), **kwargs)
        raise_for_status(resp)
        return resp.content

    async def _post_json(self, suffix: str, **kwargs: Any) -> Dict[str, Any]:
        return await self._request_json("post", suffix, **kwargs)

    async def _put_json(self, suffix: str, **kwargs: Any) -> Dict[str, Any]:
        return await self._request_json("put", suffix, **kwargs)

    async def _multipart(
        self,
        method: str,
        suffix: str,
        file_fields: Dict[str, Any],
        data_fields: Optional[Dict[str, str]] = None,
    ) -> Dict[str, Any]:
        sources = {name: src for name, (src, _) in file_fields.items()}
        files = {
            name: _file_tuple(src, default)
            for name, (src, default) in file_fields.items()
        }
        try:
            return await self._request_json(
                method, suffix, files=files, data=data_fields or None
            )
        finally:
            _close_multipart(files, sources)

    # -- unary operations ----------------------------------------------------
    async def info(self) -> Dict[str, Any]:
        return (await self._get_json("/info")) or {}

    async def screenshot(self) -> bytes:
        return await self._get_bytes("/screenshot")

    async def activate(self) -> Dict[str, Any]:
        return await self._post_json("/activate")

    async def pair(
        self,
        *,
        supervised: bool = False,
        p12file: Optional[BytesLike] = None,
        supervision_password: Optional[str] = None,
    ) -> Dict[str, Any]:
        params = {"supervised": _bool_param(supervised)}
        headers: Dict[str, str] = {}
        if supervision_password is not None:
            headers["Supervision-Password"] = supervision_password
        if p12file is None:
            return await self._post_json("/pair", params=params, headers=headers or None)
        files = {"p12file": _file_tuple(p12file, "identity.p12")}
        try:
            return await self._post_json(
                "/pair", params=params, headers=headers or None, files=files
            )
        finally:
            _close_multipart(files, {"p12file": p12file})

    # -- device info ---------------------------------------------------------
    async def device_name(self) -> Dict[str, Any]:
        return (await self._get_json("/devicename")) or {}

    async def date(self) -> Dict[str, Any]:
        return (await self._get_json("/date")) or {}

    async def battery(self) -> Dict[str, Any]:
        return (await self._get_json("/battery")) or {}

    async def diagnostics(self) -> Dict[str, Any]:
        return (await self._get_json("/diagnostics")) or {}

    async def mobilegestalt(self, keys: Optional[Sequence[str]] = None) -> Dict[str, Any]:
        params = {"key": list(keys)} if keys else None
        return (await self._get_json("/mobilegestalt", params=params)) or {}

    async def processes(self, *, apps: Optional[bool] = None) -> List[Dict[str, Any]]:
        params = {"apps": _bool_param(apps)} if apps is not None else None
        return (await self._get_json("/processes", params=params)) or []

    async def lockdown(self) -> Dict[str, Any]:
        return (await self._get_json("/lockdown")) or {}

    # -- management ----------------------------------------------------------
    async def reboot(self) -> Dict[str, Any]:
        return await self._post_json("/reboot")

    async def shutdown(self) -> Dict[str, Any]:
        return await self._post_json("/shutdown")

    async def erase(self, *, confirm: bool = False) -> Dict[str, Any]:
        return await self._post_json("/erase", params={"confirm": _bool_param(confirm)})

    async def devmode(self) -> Dict[str, Any]:
        return (await self._get_json("/devmode")) or {}

    async def set_devmode(
        self, action: str = "enable", *, enable_post_restart: Optional[bool] = None
    ) -> Dict[str, Any]:
        body: Dict[str, Any] = {"action": action}
        if enable_post_restart is not None:
            body["enablePostRestart"] = enable_post_restart
        return await self._post_json("/devmode", json=body)

    async def lang(self) -> Dict[str, Any]:
        return (await self._get_json("/lang")) or {}

    async def set_lang(
        self, *, language: Optional[str] = None, locale: Optional[str] = None
    ) -> Dict[str, Any]:
        body: Dict[str, Any] = {}
        if language is not None:
            body["language"] = language
        if locale is not None:
            body["locale"] = locale
        return await self._put_json("/lang", json=body)

    async def memlimitoff(self, process: Optional[str] = None) -> Dict[str, Any]:
        body = {"process": process} if process is not None else {}
        params = {"process": process} if process is not None else None
        return await self._post_json("/memlimitoff", json=body, params=params)

    # -- media ---------------------------------------------------------------
    async def get_wallpaper(self) -> bytes:
        return await self._get_bytes("/wallpaper")

    async def set_wallpaper(
        self,
        image: BytesLike,
        p12: BytesLike,
        *,
        password: Optional[str] = None,
        screen: Optional[str] = None,
    ) -> Dict[str, Any]:
        data: Dict[str, str] = {}
        if password is not None:
            data["password"] = password
        if screen is not None:
            data["screen"] = screen
        return await self._multipart(
            "put",
            "/wallpaper",
            {"image": (image, "wallpaper.png"), "p12": (p12, "identity.p12")},
            data or None,
        )

    async def get_icon_layout(self) -> Dict[str, Any]:
        return (await self._get_json("/icon-layout")) or {}

    async def set_icon_layout(self, layout: Dict[str, Any]) -> Dict[str, Any]:
        return await self._put_json("/icon-layout", json=layout)

    async def get_pasteboard(self) -> Dict[str, Any]:
        return (await self._get_json("/pasteboard")) or {}

    async def set_pasteboard(self, text: str) -> Dict[str, Any]:
        return await self._put_json(
            "/pasteboard",
            content=text.encode("utf-8"),
            headers={"Content-Type": "text/plain"},
        )

    # -- settings ------------------------------------------------------------
    async def assistive_touch(self) -> Dict[str, Any]:
        return (await self._get_json("/assistivetouch")) or {}

    async def set_assistive_touch(self, enabled: bool) -> Dict[str, Any]:
        return await self._put_json("/assistivetouch", json={"enabled": enabled})

    async def time_format(self) -> Dict[str, Any]:
        return (await self._get_json("/timeformat")) or {}

    async def set_time_format(self, uses_24_hour: bool) -> Dict[str, Any]:
        return await self._put_json("/timeformat", json={"uses24Hour": uses_24_hour})

    async def set_wifi(
        self, ssid: str, *, password: Optional[str] = None, enc_type: Optional[str] = None
    ) -> Dict[str, Any]:
        body: Dict[str, Any] = {"ssid": ssid}
        if password is not None:
            body["password"] = password
        if enc_type is not None:
            body["encType"] = enc_type
        return await self._put_json("/wifi", json=body)

    async def remove_wifi(self, ssid: str) -> Dict[str, Any]:
        return await self._request_json("delete", "/wifi", params={"ssid": ssid})

    # -- mdm -----------------------------------------------------------------
    async def security_info(
        self, p12: BytesLike, *, password: Optional[str] = None
    ) -> Dict[str, Any]:
        data = {"password": password} if password is not None else None
        return await self._multipart(
            "post", "/mdm/security-info", {"p12": (p12, "identity.p12")}, data
        )

    async def fetch_unlock_token(
        self, p12: BytesLike, *, password: Optional[str] = None
    ) -> Dict[str, Any]:
        data = {"password": password} if password is not None else None
        return await self._multipart(
            "post", "/mdm/fetch-unlock-token", {"p12": (p12, "identity.p12")}, data
        )

    async def clear_passcode(
        self, p12: BytesLike, token: str, *, password: Optional[str] = None
    ) -> Dict[str, Any]:
        data = {"token": token}
        if password is not None:
            data["password"] = password
        return await self._multipart(
            "post", "/mdm/clear-passcode", {"p12": (p12, "identity.p12")}, data
        )

    async def clear_screen_time_password(
        self, p12: BytesLike, *, password: Optional[str] = None
    ) -> Dict[str, Any]:
        data = {"password": password} if password is not None else None
        return await self._multipart(
            "post",
            "/mdm/clear-screen-time-password",
            {"p12": (p12, "identity.p12")},
            data,
        )

    # -- proxy ---------------------------------------------------------------
    async def set_http_proxy(
        self,
        host: str,
        port: Union[int, str],
        *,
        user: Optional[str] = None,
        password: Optional[str] = None,
        p12: Optional[BytesLike] = None,
        p12_password: Optional[str] = None,
    ) -> Dict[str, Any]:
        data: Dict[str, str] = {"host": host, "port": str(port)}
        if user is not None:
            data["user"] = user
        if password is not None:
            data["pass"] = password
        if p12_password is not None:
            data["password"] = p12_password
        file_fields = {"p12": (p12, "identity.p12")} if p12 is not None else {}
        return await self._multipart("put", "/httpproxy", file_fields, data)

    async def remove_http_proxy(self) -> Dict[str, Any]:
        return await self._request_json("delete", "/httpproxy")

    # -- conditions / images / profiles -------------------------------------
    async def conditions(self) -> List[Dict[str, Any]]:
        return (await self._get_json("/conditions")) or []

    async def enable_condition(self, profile_type_id: str, profile_id: str) -> Dict[str, Any]:
        return await self._request_json(
            "put",
            "/enable-condition",
            params={"profileTypeID": profile_type_id, "profileID": profile_id},
        )

    async def disable_condition(self) -> Dict[str, Any]:
        return await self._post_json("/disable-condition")

    async def images(self) -> List[str]:
        return (await self._get_json("/image")) or []

    async def mounted_images(self) -> Dict[str, Any]:
        return (await self._get_json("/image/list")) or {}

    async def install_image(
        self,
        image: Optional[BytesLike] = None,
        *,
        auto: bool = False,
        basedir: Optional[str] = None,
    ) -> Dict[str, Any]:
        params: Dict[str, Any] = {}
        if auto:
            params["auto"] = "true"
        if basedir is not None:
            params["basedir"] = basedir
        content = _resolve_body(image) if image is not None else None
        return await self._request_json(
            "put",
            "/image",
            params=params or None,
            content=content,
            headers={"Content-Type": "application/octet-stream"} if content else None,
        )

    async def unmount_image(self) -> Dict[str, Any]:
        return await self._request_json("delete", "/image")

    async def profiles(self) -> Dict[str, Any]:
        return (await self._get_json("/profiles")) or {}

    async def add_profile(
        self,
        profile: BytesLike,
        *,
        p12: Optional[BytesLike] = None,
        password: Optional[str] = None,
    ) -> Dict[str, Any]:
        data = {"password": password} if password is not None else None
        file_fields = {"profile": (profile, "profile.mobileconfig")}
        if p12 is not None:
            file_fields["p12"] = (p12, "identity.p12")
        return await self._multipart("post", "/profiles", file_fields, data)

    async def remove_profile(self, name: str) -> Dict[str, Any]:
        return await self._request_json("delete", f"/profiles/{name}")

    async def reset_accessibility(self) -> Dict[str, Any]:
        return await self._post_json("/resetaccessibility")

    async def reset_location(self) -> Dict[str, Any]:
        return await self._post_json("/resetlocation")

    async def set_location(self, latitude: float, longitude: float) -> Dict[str, Any]:
        return await self._request_json(
            "put",
            "/setlocation",
            params={"latitude": str(latitude), "longitude": str(longitude)},
        )

    # -- streaming operations ------------------------------------------------
    def _stream(self, suffix: str, *, params: Optional[Dict[str, Any]] = None,
                include_heartbeats: bool = False) -> _AsyncStream:
        cm = self._c._http.stream(
            "GET", self._url(suffix), params=params, timeout=_STREAM_TIMEOUT
        )
        return _AsyncStream(cm, include_heartbeats)

    def syslog(self, *, include_heartbeats: bool = False) -> _AsyncStream:
        return self._stream("/syslog", include_heartbeats=include_heartbeats)

    def notifications(self, *, include_heartbeats: bool = False) -> _AsyncStream:
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
    ) -> _AsyncStream:
        params = _ostrace_params(pid, level, subsystem, match, exclude)
        return self._stream("/ostrace", params=params, include_heartbeats=include_heartbeats)

    def listen(self, *, include_heartbeats: bool = False) -> _AsyncStream:
        return self._stream("/listen", include_heartbeats=include_heartbeats)

    def sysmontap(self, *, include_heartbeats: bool = False) -> _AsyncStream:
        return self._stream("/sysmontap", include_heartbeats=include_heartbeats)


class AsyncIosClient:
    """Asynchronous go-ios REST client. See :class:`go_ios_sdk.client.IosClient`."""

    def __init__(
        self,
        base_url: str = DEFAULT_BASE_URL,
        *,
        api_key: Optional[str] = None,
        timeout: float = 30.0,
        verify: Any = True,
        headers: Optional[Dict[str, str]] = None,
        http_client: Optional[httpx.AsyncClient] = None,
    ) -> None:
        self.base_url = base_url.rstrip("/")
        self.api_key = api_key
        if http_client is not None:
            self._http = http_client
            self._owns_http = False
        else:
            self._http = httpx.AsyncClient(
                base_url=self.base_url,
                headers=build_headers(api_key, headers),
                timeout=timeout,
                verify=verify,
            )
            self._owns_http = True
        self.devices = AsyncDevices(self)
        self.tunnels = AsyncTunnels(self)

    async def _get_json(self, suffix: str, **kwargs: Any) -> Any:
        resp = await self._http.get(f"{API_PREFIX}{suffix}", **kwargs)
        raise_for_status(resp)
        return json_or_none(resp)

    async def _request_json(self, method: str, suffix: str, **kwargs: Any) -> Dict[str, Any]:
        resp = await self._http.request(method, f"{API_PREFIX}{suffix}", **kwargs)
        raise_for_status(resp)
        return json_or_none(resp) or {}

    def device(self, udid: str) -> AsyncDevice:
        return AsyncDevice(self, udid)

    async def aclose(self) -> None:
        if self._owns_http:
            await self._http.aclose()

    async def __aenter__(self) -> "AsyncIosClient":
        return self

    async def __aexit__(self, *exc: Any) -> None:
        await self.aclose()
