from http import HTTPStatus
from typing import Any, Optional, Union

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.generic_response import GenericResponse
from ...models.mobile_gestalt import MobileGestalt
from ...types import UNSET, Response


def _get_kwargs(
    udid: str,
    *,
    key: list[str],
) -> dict[str, Any]:
    params: dict[str, Any] = {}

    json_key = key

    params["key"] = json_key

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/api/v1/device/{udid}/mobilegestalt".format(
            udid=udid,
        ),
        "params": params,
    }

    return _kwargs


def _parse_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Optional[Union[GenericResponse, MobileGestalt]]:
    if response.status_code == 200:
        response_200 = MobileGestalt.from_dict(response.json())

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
) -> Response[Union[GenericResponse, MobileGestalt]]:
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
    key: list[str],
) -> Response[Union[GenericResponse, MobileGestalt]]:
    """Query MobileGestalt

     Query one or more MobileGestalt keys (CLI: `ios mobilegestalt <key>...`).
    Pass repeated `key` query params.

    Args:
        udid (str):
        key (list[str]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[GenericResponse, MobileGestalt]]
    """

    kwargs = _get_kwargs(
        udid=udid,
        key=key,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    key: list[str],
) -> Optional[Union[GenericResponse, MobileGestalt]]:
    """Query MobileGestalt

     Query one or more MobileGestalt keys (CLI: `ios mobilegestalt <key>...`).
    Pass repeated `key` query params.

    Args:
        udid (str):
        key (list[str]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[GenericResponse, MobileGestalt]
    """

    return sync_detailed(
        udid=udid,
        client=client,
        key=key,
    ).parsed


async def asyncio_detailed(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    key: list[str],
) -> Response[Union[GenericResponse, MobileGestalt]]:
    """Query MobileGestalt

     Query one or more MobileGestalt keys (CLI: `ios mobilegestalt <key>...`).
    Pass repeated `key` query params.

    Args:
        udid (str):
        key (list[str]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[GenericResponse, MobileGestalt]]
    """

    kwargs = _get_kwargs(
        udid=udid,
        key=key,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    key: list[str],
) -> Optional[Union[GenericResponse, MobileGestalt]]:
    """Query MobileGestalt

     Query one or more MobileGestalt keys (CLI: `ios mobilegestalt <key>...`).
    Pass repeated `key` query params.

    Args:
        udid (str):
        key (list[str]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[GenericResponse, MobileGestalt]
    """

    return (
        await asyncio_detailed(
            udid=udid,
            client=client,
            key=key,
        )
    ).parsed
