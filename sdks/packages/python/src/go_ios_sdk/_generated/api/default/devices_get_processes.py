from http import HTTPStatus
from typing import Any, Optional, Union

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.generic_response import GenericResponse
from ...models.process_info import ProcessInfo
from ...types import UNSET, Response, Unset


def _get_kwargs(
    udid: str,
    *,
    apps: Union[Unset, bool] = UNSET,
) -> dict[str, Any]:
    params: dict[str, Any] = {}

    params["apps"] = apps

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/api/v1/device/{udid}/processes".format(
            udid=udid,
        ),
        "params": params,
    }

    return _kwargs


def _parse_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Optional[Union[GenericResponse, list["ProcessInfo"]]]:
    if response.status_code == 200:
        response_200 = []
        _response_200 = response.json()
        for response_200_item_data in _response_200:
            response_200_item = ProcessInfo.from_dict(response_200_item_data)

            response_200.append(response_200_item)

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
) -> Response[Union[GenericResponse, list["ProcessInfo"]]]:
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
    apps: Union[Unset, bool] = UNSET,
) -> Response[Union[GenericResponse, list["ProcessInfo"]]]:
    """List processes

     List running processes (CLI: `ios ps [--apps]`).

    Args:
        udid (str):
        apps (Union[Unset, bool]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[GenericResponse, list['ProcessInfo']]]
    """

    kwargs = _get_kwargs(
        udid=udid,
        apps=apps,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    apps: Union[Unset, bool] = UNSET,
) -> Optional[Union[GenericResponse, list["ProcessInfo"]]]:
    """List processes

     List running processes (CLI: `ios ps [--apps]`).

    Args:
        udid (str):
        apps (Union[Unset, bool]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[GenericResponse, list['ProcessInfo']]
    """

    return sync_detailed(
        udid=udid,
        client=client,
        apps=apps,
    ).parsed


async def asyncio_detailed(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    apps: Union[Unset, bool] = UNSET,
) -> Response[Union[GenericResponse, list["ProcessInfo"]]]:
    """List processes

     List running processes (CLI: `ios ps [--apps]`).

    Args:
        udid (str):
        apps (Union[Unset, bool]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[GenericResponse, list['ProcessInfo']]]
    """

    kwargs = _get_kwargs(
        udid=udid,
        apps=apps,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    apps: Union[Unset, bool] = UNSET,
) -> Optional[Union[GenericResponse, list["ProcessInfo"]]]:
    """List processes

     List running processes (CLI: `ios ps [--apps]`).

    Args:
        udid (str):
        apps (Union[Unset, bool]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[GenericResponse, list['ProcessInfo']]
    """

    return (
        await asyncio_detailed(
            udid=udid,
            client=client,
            apps=apps,
        )
    ).parsed
