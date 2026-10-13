from http import HTTPStatus
from typing import Any, Optional, Union

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.generic_response import GenericResponse
from ...models.tunnel import Tunnel
from ...types import Response


def _get_kwargs(
    udid: str,
) -> dict[str, Any]:
    _kwargs: dict[str, Any] = {
        "method": "post",
        "url": "/api/v1/tunnels/{udid}/refresh".format(
            udid=udid,
        ),
    }

    return _kwargs


def _parse_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Optional[Union[GenericResponse, Tunnel]]:
    if response.status_code == 200:
        response_200 = Tunnel.from_dict(response.json())

        return response_200

    if response.status_code == 401:
        response_401 = GenericResponse.from_dict(response.json())

        return response_401

    if response.status_code == 502:
        response_502 = GenericResponse.from_dict(response.json())

        return response_502

    if client.raise_on_unexpected_status:
        raise errors.UnexpectedStatus(response.status_code, response.content)
    else:
        return None


def _build_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Response[Union[GenericResponse, Tunnel]]:
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
) -> Response[Union[GenericResponse, Tunnel]]:
    """Refresh tunnel

     Restart the tunnel for a device and wait for it to come up
    (CLI: `ios tunnel refresh`).

    Args:
        udid (str):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[GenericResponse, Tunnel]]
    """

    kwargs = _get_kwargs(
        udid=udid,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
) -> Optional[Union[GenericResponse, Tunnel]]:
    """Refresh tunnel

     Restart the tunnel for a device and wait for it to come up
    (CLI: `ios tunnel refresh`).

    Args:
        udid (str):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[GenericResponse, Tunnel]
    """

    return sync_detailed(
        udid=udid,
        client=client,
    ).parsed


async def asyncio_detailed(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
) -> Response[Union[GenericResponse, Tunnel]]:
    """Refresh tunnel

     Restart the tunnel for a device and wait for it to come up
    (CLI: `ios tunnel refresh`).

    Args:
        udid (str):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[GenericResponse, Tunnel]]
    """

    kwargs = _get_kwargs(
        udid=udid,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
) -> Optional[Union[GenericResponse, Tunnel]]:
    """Refresh tunnel

     Restart the tunnel for a device and wait for it to come up
    (CLI: `ios tunnel refresh`).

    Args:
        udid (str):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[GenericResponse, Tunnel]
    """

    return (
        await asyncio_detailed(
            udid=udid,
            client=client,
        )
    ).parsed
