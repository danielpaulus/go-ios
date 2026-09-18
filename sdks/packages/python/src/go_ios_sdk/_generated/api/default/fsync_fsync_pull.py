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
    bundle_id: Union[Unset, str] = UNSET,
    path: str,
) -> dict[str, Any]:
    params: dict[str, Any] = {}

    params["bundleID"] = bundle_id

    params["path"] = path

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/api/v1/device/{udid}/fsync/pull".format(
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
    bundle_id: Union[Unset, str] = UNSET,
    path: str,
) -> Response[Union[Any, GenericResponse]]:
    """Download a file over AFC

     Download a file from the device over AFC (CLI: `ios fsync pull`). Returns
    the raw file bytes. `path` is required.

    Args:
        udid (str):
        bundle_id (Union[Unset, str]):
        path (str):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[Any, GenericResponse]]
    """

    kwargs = _get_kwargs(
        udid=udid,
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
    bundle_id: Union[Unset, str] = UNSET,
    path: str,
) -> Optional[Union[Any, GenericResponse]]:
    """Download a file over AFC

     Download a file from the device over AFC (CLI: `ios fsync pull`). Returns
    the raw file bytes. `path` is required.

    Args:
        udid (str):
        bundle_id (Union[Unset, str]):
        path (str):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[Any, GenericResponse]
    """

    return sync_detailed(
        udid=udid,
        client=client,
        bundle_id=bundle_id,
        path=path,
    ).parsed


async def asyncio_detailed(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    bundle_id: Union[Unset, str] = UNSET,
    path: str,
) -> Response[Union[Any, GenericResponse]]:
    """Download a file over AFC

     Download a file from the device over AFC (CLI: `ios fsync pull`). Returns
    the raw file bytes. `path` is required.

    Args:
        udid (str):
        bundle_id (Union[Unset, str]):
        path (str):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[Any, GenericResponse]]
    """

    kwargs = _get_kwargs(
        udid=udid,
        bundle_id=bundle_id,
        path=path,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    bundle_id: Union[Unset, str] = UNSET,
    path: str,
) -> Optional[Union[Any, GenericResponse]]:
    """Download a file over AFC

     Download a file from the device over AFC (CLI: `ios fsync pull`). Returns
    the raw file bytes. `path` is required.

    Args:
        udid (str):
        bundle_id (Union[Unset, str]):
        path (str):

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
            bundle_id=bundle_id,
            path=path,
        )
    ).parsed
