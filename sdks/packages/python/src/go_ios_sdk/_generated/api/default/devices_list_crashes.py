from http import HTTPStatus
from typing import Any, Optional, Union

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.crash_listing import CrashListing
from ...models.generic_response import GenericResponse
from ...types import UNSET, Response, Unset


def _get_kwargs(
    udid: str,
    *,
    pattern: Union[Unset, str] = UNSET,
) -> dict[str, Any]:
    params: dict[str, Any] = {}

    params["pattern"] = pattern

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/api/v1/device/{udid}/crashes".format(
            udid=udid,
        ),
        "params": params,
    }

    return _kwargs


def _parse_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Optional[Union[CrashListing, GenericResponse]]:
    if response.status_code == 200:
        response_200 = CrashListing.from_dict(response.json())

        return response_200

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
) -> Response[Union[CrashListing, GenericResponse]]:
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
    pattern: Union[Unset, str] = UNSET,
) -> Response[Union[CrashListing, GenericResponse]]:
    """List crash reports

     List crash reports (CLI: `ios crash ls`).

    Args:
        udid (str):
        pattern (Union[Unset, str]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[CrashListing, GenericResponse]]
    """

    kwargs = _get_kwargs(
        udid=udid,
        pattern=pattern,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    pattern: Union[Unset, str] = UNSET,
) -> Optional[Union[CrashListing, GenericResponse]]:
    """List crash reports

     List crash reports (CLI: `ios crash ls`).

    Args:
        udid (str):
        pattern (Union[Unset, str]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[CrashListing, GenericResponse]
    """

    return sync_detailed(
        udid=udid,
        client=client,
        pattern=pattern,
    ).parsed


async def asyncio_detailed(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    pattern: Union[Unset, str] = UNSET,
) -> Response[Union[CrashListing, GenericResponse]]:
    """List crash reports

     List crash reports (CLI: `ios crash ls`).

    Args:
        udid (str):
        pattern (Union[Unset, str]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[CrashListing, GenericResponse]]
    """

    kwargs = _get_kwargs(
        udid=udid,
        pattern=pattern,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    pattern: Union[Unset, str] = UNSET,
) -> Optional[Union[CrashListing, GenericResponse]]:
    """List crash reports

     List crash reports (CLI: `ios crash ls`).

    Args:
        udid (str):
        pattern (Union[Unset, str]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[CrashListing, GenericResponse]
    """

    return (
        await asyncio_detailed(
            udid=udid,
            client=client,
            pattern=pattern,
        )
    ).parsed
