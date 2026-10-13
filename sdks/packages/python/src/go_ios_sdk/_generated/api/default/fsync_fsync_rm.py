from http import HTTPStatus
from typing import Any, Optional, Union

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.fsync_message import FsyncMessage
from ...models.generic_response import GenericResponse
from ...types import UNSET, Response, Unset


def _get_kwargs(
    udid: str,
    *,
    bundle_id: Union[Unset, str] = UNSET,
    path: str,
    recursive: Union[Unset, bool] = UNSET,
) -> dict[str, Any]:
    params: dict[str, Any] = {}

    params["bundleID"] = bundle_id

    params["path"] = path

    params["recursive"] = recursive

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "delete",
        "url": "/api/v1/device/{udid}/fsync/rm".format(
            udid=udid,
        ),
        "params": params,
    }

    return _kwargs


def _parse_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Optional[Union[FsyncMessage, GenericResponse]]:
    if response.status_code == 200:
        response_200 = FsyncMessage.from_dict(response.json())

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

    if client.raise_on_unexpected_status:
        raise errors.UnexpectedStatus(response.status_code, response.content)
    else:
        return None


def _build_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Response[Union[FsyncMessage, GenericResponse]]:
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
    bundle_id: Union[Unset, str] = UNSET,
    path: str,
    recursive: Union[Unset, bool] = UNSET,
) -> Response[Union[FsyncMessage, GenericResponse]]:
    """Remove a file or directory over AFC

     Remove a file or directory over AFC (CLI: `ios fsync rm`). Pass
    `recursive=true` to delete a non-empty directory.

    Args:
        udid (str):
        bundle_id (Union[Unset, str]):
        path (str):
        recursive (Union[Unset, bool]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[FsyncMessage, GenericResponse]]
    """

    kwargs = _get_kwargs(
        udid=udid,
        bundle_id=bundle_id,
        path=path,
        recursive=recursive,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    bundle_id: Union[Unset, str] = UNSET,
    path: str,
    recursive: Union[Unset, bool] = UNSET,
) -> Optional[Union[FsyncMessage, GenericResponse]]:
    """Remove a file or directory over AFC

     Remove a file or directory over AFC (CLI: `ios fsync rm`). Pass
    `recursive=true` to delete a non-empty directory.

    Args:
        udid (str):
        bundle_id (Union[Unset, str]):
        path (str):
        recursive (Union[Unset, bool]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[FsyncMessage, GenericResponse]
    """

    return sync_detailed(
        udid=udid,
        client=client,
        bundle_id=bundle_id,
        path=path,
        recursive=recursive,
    ).parsed


async def asyncio_detailed(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    bundle_id: Union[Unset, str] = UNSET,
    path: str,
    recursive: Union[Unset, bool] = UNSET,
) -> Response[Union[FsyncMessage, GenericResponse]]:
    """Remove a file or directory over AFC

     Remove a file or directory over AFC (CLI: `ios fsync rm`). Pass
    `recursive=true` to delete a non-empty directory.

    Args:
        udid (str):
        bundle_id (Union[Unset, str]):
        path (str):
        recursive (Union[Unset, bool]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[FsyncMessage, GenericResponse]]
    """

    kwargs = _get_kwargs(
        udid=udid,
        bundle_id=bundle_id,
        path=path,
        recursive=recursive,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    bundle_id: Union[Unset, str] = UNSET,
    path: str,
    recursive: Union[Unset, bool] = UNSET,
) -> Optional[Union[FsyncMessage, GenericResponse]]:
    """Remove a file or directory over AFC

     Remove a file or directory over AFC (CLI: `ios fsync rm`). Pass
    `recursive=true` to delete a non-empty directory.

    Args:
        udid (str):
        bundle_id (Union[Unset, str]):
        path (str):
        recursive (Union[Unset, bool]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[FsyncMessage, GenericResponse]
    """

    return (
        await asyncio_detailed(
            udid=udid,
            client=client,
            bundle_id=bundle_id,
            path=path,
            recursive=recursive,
        )
    ).parsed
