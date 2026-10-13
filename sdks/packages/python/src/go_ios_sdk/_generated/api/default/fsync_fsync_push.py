from http import HTTPStatus
from typing import Any, Optional, Union

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.fsync_push_result import FsyncPushResult
from ...models.generic_response import GenericResponse
from ...types import UNSET, Response, Unset


def _get_kwargs(
    udid: str,
    *,
    body: Any,
    bundle_id: Union[Unset, str] = UNSET,
    path: str,
) -> dict[str, Any]:
    headers: dict[str, Any] = {}

    params: dict[str, Any] = {}

    params["bundleID"] = bundle_id

    params["path"] = path

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "post",
        "url": "/api/v1/device/{udid}/fsync/push".format(
            udid=udid,
        ),
        "params": params,
    }

    _kwargs["content"] = body.payload

    headers["Content-Type"] = "application/octet-stream"

    _kwargs["headers"] = headers
    return _kwargs


def _parse_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Optional[Union[FsyncPushResult, GenericResponse]]:
    if response.status_code == 200:
        response_200 = FsyncPushResult.from_dict(response.json())

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

    if response.status_code == 413:
        response_413 = GenericResponse.from_dict(response.json())

        return response_413

    if response.status_code == 422:
        response_422 = GenericResponse.from_dict(response.json())

        return response_422

    if response.status_code == 500:
        response_500 = GenericResponse.from_dict(response.json())

        return response_500

    if client.raise_on_unexpected_status:
        raise errors.UnexpectedStatus(response.status_code, response.content)
    else:
        return None


def _build_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Response[Union[FsyncPushResult, GenericResponse]]:
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
    body: Any,
    bundle_id: Union[Unset, str] = UNSET,
    path: str,
) -> Response[Union[FsyncPushResult, GenericResponse]]:
    """Upload a file over AFC

     Upload a file to the device over AFC (CLI: `ios fsync push`). Accepts either
    raw bytes (application/octet-stream) or a multipart form with a `file`
    field. `path` is required. Bounded server-side; oversized uploads get `413`.

    Args:
        udid (str):
        bundle_id (Union[Unset, str]):
        path (str):
        body (Any):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[FsyncPushResult, GenericResponse]]
    """

    kwargs = _get_kwargs(
        udid=udid,
        body=body,
        bundle_id=bundle_id,
        path=path,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    body: Any,
    bundle_id: Union[Unset, str] = UNSET,
    path: str,
) -> Optional[Union[FsyncPushResult, GenericResponse]]:
    """Upload a file over AFC

     Upload a file to the device over AFC (CLI: `ios fsync push`). Accepts either
    raw bytes (application/octet-stream) or a multipart form with a `file`
    field. `path` is required. Bounded server-side; oversized uploads get `413`.

    Args:
        udid (str):
        bundle_id (Union[Unset, str]):
        path (str):
        body (Any):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[FsyncPushResult, GenericResponse]
    """

    return sync_detailed(
        udid=udid,
        client=client,
        body=body,
        bundle_id=bundle_id,
        path=path,
    ).parsed


async def asyncio_detailed(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    body: Any,
    bundle_id: Union[Unset, str] = UNSET,
    path: str,
) -> Response[Union[FsyncPushResult, GenericResponse]]:
    """Upload a file over AFC

     Upload a file to the device over AFC (CLI: `ios fsync push`). Accepts either
    raw bytes (application/octet-stream) or a multipart form with a `file`
    field. `path` is required. Bounded server-side; oversized uploads get `413`.

    Args:
        udid (str):
        bundle_id (Union[Unset, str]):
        path (str):
        body (Any):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[FsyncPushResult, GenericResponse]]
    """

    kwargs = _get_kwargs(
        udid=udid,
        body=body,
        bundle_id=bundle_id,
        path=path,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    body: Any,
    bundle_id: Union[Unset, str] = UNSET,
    path: str,
) -> Optional[Union[FsyncPushResult, GenericResponse]]:
    """Upload a file over AFC

     Upload a file to the device over AFC (CLI: `ios fsync push`). Accepts either
    raw bytes (application/octet-stream) or a multipart form with a `file`
    field. `path` is required. Bounded server-side; oversized uploads get `413`.

    Args:
        udid (str):
        bundle_id (Union[Unset, str]):
        path (str):
        body (Any):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[FsyncPushResult, GenericResponse]
    """

    return (
        await asyncio_detailed(
            udid=udid,
            client=client,
            body=body,
            bundle_id=bundle_id,
            path=path,
        )
    ).parsed
