from http import HTTPStatus
from typing import Any, Optional, Union

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.generic_response import GenericResponse
from ...models.lockdown_values import LockdownValues
from ...types import UNSET, Response, Unset


def _get_kwargs(
    udid: str,
    *,
    domain: Union[Unset, str] = UNSET,
) -> dict[str, Any]:
    params: dict[str, Any] = {}

    params["domain"] = domain

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/api/v1/device/{udid}/lockdown".format(
            udid=udid,
        ),
        "params": params,
    }

    return _kwargs


def _parse_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Optional[Union[GenericResponse, LockdownValues]]:
    if response.status_code == 200:
        response_200 = LockdownValues.from_dict(response.json())

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
) -> Response[Union[GenericResponse, LockdownValues]]:
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
    domain: Union[Unset, str] = UNSET,
) -> Response[Union[GenericResponse, LockdownValues]]:
    """Get lockdown values

     Get lockdown values (CLI: `ios lockdown get`). Without `domain` the full set
    is returned; with `domain` the values are scoped to that lockdown domain.

    Args:
        udid (str):
        domain (Union[Unset, str]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[GenericResponse, LockdownValues]]
    """

    kwargs = _get_kwargs(
        udid=udid,
        domain=domain,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    domain: Union[Unset, str] = UNSET,
) -> Optional[Union[GenericResponse, LockdownValues]]:
    """Get lockdown values

     Get lockdown values (CLI: `ios lockdown get`). Without `domain` the full set
    is returned; with `domain` the values are scoped to that lockdown domain.

    Args:
        udid (str):
        domain (Union[Unset, str]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[GenericResponse, LockdownValues]
    """

    return sync_detailed(
        udid=udid,
        client=client,
        domain=domain,
    ).parsed


async def asyncio_detailed(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    domain: Union[Unset, str] = UNSET,
) -> Response[Union[GenericResponse, LockdownValues]]:
    """Get lockdown values

     Get lockdown values (CLI: `ios lockdown get`). Without `domain` the full set
    is returned; with `domain` the values are scoped to that lockdown domain.

    Args:
        udid (str):
        domain (Union[Unset, str]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[GenericResponse, LockdownValues]]
    """

    kwargs = _get_kwargs(
        udid=udid,
        domain=domain,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    domain: Union[Unset, str] = UNSET,
) -> Optional[Union[GenericResponse, LockdownValues]]:
    """Get lockdown values

     Get lockdown values (CLI: `ios lockdown get`). Without `domain` the full set
    is returned; with `domain` the values are scoped to that lockdown domain.

    Args:
        udid (str):
        domain (Union[Unset, str]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[GenericResponse, LockdownValues]
    """

    return (
        await asyncio_detailed(
            udid=udid,
            client=client,
            domain=domain,
        )
    ).parsed
