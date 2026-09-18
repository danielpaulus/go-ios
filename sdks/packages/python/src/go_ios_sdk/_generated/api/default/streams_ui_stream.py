from http import HTTPStatus
from typing import Any, Optional, Union, cast

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.generic_response import GenericResponse
from ...types import UNSET, Response, Unset


def _get_kwargs(
    udid: str,
    *,
    backend: Union[Unset, str] = UNSET,
    wda_url: Union[Unset, str] = UNSET,
    timeout: Union[Unset, int] = UNSET,
    codec: Union[Unset, str] = UNSET,
    fps: Union[Unset, str] = UNSET,
    quality: Union[Unset, str] = UNSET,
    scale: Union[Unset, str] = UNSET,
    bitrate: Union[Unset, str] = UNSET,
) -> dict[str, Any]:
    params: dict[str, Any] = {}

    params["backend"] = backend

    params["wdaUrl"] = wda_url

    params["timeout"] = timeout

    params["codec"] = codec

    params["fps"] = fps

    params["quality"] = quality

    params["scale"] = scale

    params["bitrate"] = bitrate

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/api/v1/device/{udid}/ui/stream".format(
            udid=udid,
        ),
        "params": params,
    }

    return _kwargs


def _parse_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Optional[Union[Any, GenericResponse]]:
    if response.status_code == 200:
        response_200 = cast(Any, response.content)
        return response_200

    if response.status_code == 400:
        response_400 = GenericResponse.from_dict(response.json())

        return response_400

    if response.status_code == 401:
        response_401 = GenericResponse.from_dict(response.json())

        return response_401

    if response.status_code == 404:
        response_404 = GenericResponse.from_dict(response.json())

        return response_404

    if response.status_code == 422:
        response_422 = GenericResponse.from_dict(response.json())

        return response_422

    if response.status_code == 500:
        response_500 = GenericResponse.from_dict(response.json())

        return response_500

    if response.status_code == 501:
        response_501 = GenericResponse.from_dict(response.json())

        return response_501

    if response.status_code == 502:
        response_502 = GenericResponse.from_dict(response.json())

        return response_502

    if client.raise_on_unexpected_status:
        raise errors.UnexpectedStatus(response.status_code, response.content)
    else:
        return None


def _build_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Response[Union[Any, GenericResponse]]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    backend: Union[Unset, str] = UNSET,
    wda_url: Union[Unset, str] = UNSET,
    timeout: Union[Unset, int] = UNSET,
    codec: Union[Unset, str] = UNSET,
    fps: Union[Unset, str] = UNSET,
    quality: Union[Unset, str] = UNSET,
    scale: Union[Unset, str] = UNSET,
    bitrate: Union[Unset, str] = UNSET,
) -> Response[Union[Any, GenericResponse]]:
    """Stream UI video (binary)

     Open a live UI video stream against a forwarded WDA/DeviceKit backend and
    pipe it straight through to the client. Default codec is MJPEG
    (multipart/x-mixed-replace); `codec=h264` returns an H.264 elementary
    stream (requires the devicekit backend). Streams until the client
    disconnects or the backend ends.

    Requires a running, forwarded WDA/DeviceKit backend (see the UI routes).

    Args:
        udid (str):
        backend (Union[Unset, str]):
        wda_url (Union[Unset, str]):
        timeout (Union[Unset, int]):
        codec (Union[Unset, str]):
        fps (Union[Unset, str]):
        quality (Union[Unset, str]):
        scale (Union[Unset, str]):
        bitrate (Union[Unset, str]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[Any, GenericResponse]]
    """

    kwargs = _get_kwargs(
        udid=udid,
        backend=backend,
        wda_url=wda_url,
        timeout=timeout,
        codec=codec,
        fps=fps,
        quality=quality,
        scale=scale,
        bitrate=bitrate,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    backend: Union[Unset, str] = UNSET,
    wda_url: Union[Unset, str] = UNSET,
    timeout: Union[Unset, int] = UNSET,
    codec: Union[Unset, str] = UNSET,
    fps: Union[Unset, str] = UNSET,
    quality: Union[Unset, str] = UNSET,
    scale: Union[Unset, str] = UNSET,
    bitrate: Union[Unset, str] = UNSET,
) -> Optional[Union[Any, GenericResponse]]:
    """Stream UI video (binary)

     Open a live UI video stream against a forwarded WDA/DeviceKit backend and
    pipe it straight through to the client. Default codec is MJPEG
    (multipart/x-mixed-replace); `codec=h264` returns an H.264 elementary
    stream (requires the devicekit backend). Streams until the client
    disconnects or the backend ends.

    Requires a running, forwarded WDA/DeviceKit backend (see the UI routes).

    Args:
        udid (str):
        backend (Union[Unset, str]):
        wda_url (Union[Unset, str]):
        timeout (Union[Unset, int]):
        codec (Union[Unset, str]):
        fps (Union[Unset, str]):
        quality (Union[Unset, str]):
        scale (Union[Unset, str]):
        bitrate (Union[Unset, str]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[Any, GenericResponse]
    """

    return sync_detailed(
        udid=udid,
        client=client,
        backend=backend,
        wda_url=wda_url,
        timeout=timeout,
        codec=codec,
        fps=fps,
        quality=quality,
        scale=scale,
        bitrate=bitrate,
    ).parsed


async def asyncio_detailed(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    backend: Union[Unset, str] = UNSET,
    wda_url: Union[Unset, str] = UNSET,
    timeout: Union[Unset, int] = UNSET,
    codec: Union[Unset, str] = UNSET,
    fps: Union[Unset, str] = UNSET,
    quality: Union[Unset, str] = UNSET,
    scale: Union[Unset, str] = UNSET,
    bitrate: Union[Unset, str] = UNSET,
) -> Response[Union[Any, GenericResponse]]:
    """Stream UI video (binary)

     Open a live UI video stream against a forwarded WDA/DeviceKit backend and
    pipe it straight through to the client. Default codec is MJPEG
    (multipart/x-mixed-replace); `codec=h264` returns an H.264 elementary
    stream (requires the devicekit backend). Streams until the client
    disconnects or the backend ends.

    Requires a running, forwarded WDA/DeviceKit backend (see the UI routes).

    Args:
        udid (str):
        backend (Union[Unset, str]):
        wda_url (Union[Unset, str]):
        timeout (Union[Unset, int]):
        codec (Union[Unset, str]):
        fps (Union[Unset, str]):
        quality (Union[Unset, str]):
        scale (Union[Unset, str]):
        bitrate (Union[Unset, str]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[Any, GenericResponse]]
    """

    kwargs = _get_kwargs(
        udid=udid,
        backend=backend,
        wda_url=wda_url,
        timeout=timeout,
        codec=codec,
        fps=fps,
        quality=quality,
        scale=scale,
        bitrate=bitrate,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    backend: Union[Unset, str] = UNSET,
    wda_url: Union[Unset, str] = UNSET,
    timeout: Union[Unset, int] = UNSET,
    codec: Union[Unset, str] = UNSET,
    fps: Union[Unset, str] = UNSET,
    quality: Union[Unset, str] = UNSET,
    scale: Union[Unset, str] = UNSET,
    bitrate: Union[Unset, str] = UNSET,
) -> Optional[Union[Any, GenericResponse]]:
    """Stream UI video (binary)

     Open a live UI video stream against a forwarded WDA/DeviceKit backend and
    pipe it straight through to the client. Default codec is MJPEG
    (multipart/x-mixed-replace); `codec=h264` returns an H.264 elementary
    stream (requires the devicekit backend). Streams until the client
    disconnects or the backend ends.

    Requires a running, forwarded WDA/DeviceKit backend (see the UI routes).

    Args:
        udid (str):
        backend (Union[Unset, str]):
        wda_url (Union[Unset, str]):
        timeout (Union[Unset, int]):
        codec (Union[Unset, str]):
        fps (Union[Unset, str]):
        quality (Union[Unset, str]):
        scale (Union[Unset, str]):
        bitrate (Union[Unset, str]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[Any, GenericResponse]
    """

    return (
        await asyncio_detailed(
            udid=udid,
            client=client,
            backend=backend,
            wda_url=wda_url,
            timeout=timeout,
            codec=codec,
            fps=fps,
            quality=quality,
            scale=scale,
            bitrate=bitrate,
        )
    ).parsed
