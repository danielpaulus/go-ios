from http import HTTPStatus
from typing import Any, Optional, Union

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
) -> dict[str, Any]:
    params: dict[str, Any] = {}

    params["backend"] = backend

    params["wdaUrl"] = wda_url

    params["timeout"] = timeout

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/api/v1/device/{udid}/ui/source".format(
            udid=udid,
        ),
        "params": params,
    }

    return _kwargs


def _parse_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Optional[GenericResponse]:
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
) -> Response[GenericResponse]:
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
) -> Response[GenericResponse]:
    """UI source hierarchy

     Return the current view hierarchy (XML for WDA; backend Content-Type preserved).

    Args:
        udid (str):
        backend (Union[Unset, str]):
        wda_url (Union[Unset, str]):
        timeout (Union[Unset, int]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[GenericResponse]
    """

    kwargs = _get_kwargs(
        udid=udid,
        backend=backend,
        wda_url=wda_url,
        timeout=timeout,
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
) -> Optional[GenericResponse]:
    """UI source hierarchy

     Return the current view hierarchy (XML for WDA; backend Content-Type preserved).

    Args:
        udid (str):
        backend (Union[Unset, str]):
        wda_url (Union[Unset, str]):
        timeout (Union[Unset, int]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        GenericResponse
    """

    return sync_detailed(
        udid=udid,
        client=client,
        backend=backend,
        wda_url=wda_url,
        timeout=timeout,
    ).parsed


async def asyncio_detailed(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    backend: Union[Unset, str] = UNSET,
    wda_url: Union[Unset, str] = UNSET,
    timeout: Union[Unset, int] = UNSET,
) -> Response[GenericResponse]:
    """UI source hierarchy

     Return the current view hierarchy (XML for WDA; backend Content-Type preserved).

    Args:
        udid (str):
        backend (Union[Unset, str]):
        wda_url (Union[Unset, str]):
        timeout (Union[Unset, int]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[GenericResponse]
    """

    kwargs = _get_kwargs(
        udid=udid,
        backend=backend,
        wda_url=wda_url,
        timeout=timeout,
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
) -> Optional[GenericResponse]:
    """UI source hierarchy

     Return the current view hierarchy (XML for WDA; backend Content-Type preserved).

    Args:
        udid (str):
        backend (Union[Unset, str]):
        wda_url (Union[Unset, str]):
        timeout (Union[Unset, int]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        GenericResponse
    """

    return (
        await asyncio_detailed(
            udid=udid,
            client=client,
            backend=backend,
            wda_url=wda_url,
            timeout=timeout,
        )
    ).parsed
