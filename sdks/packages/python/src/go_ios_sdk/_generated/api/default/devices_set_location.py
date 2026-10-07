from http import HTTPStatus
from typing import Any, Optional, Union

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.generic_response import GenericResponse
from ...types import UNSET, Response


def _get_kwargs(
    udid: str,
    *,
    latitude: str,
    longitude: str,
) -> dict[str, Any]:
    params: dict[str, Any] = {}

    params["latitude"] = latitude

    params["longitude"] = longitude

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "put",
        "url": "/api/v1/device/{udid}/setlocation".format(
            udid=udid,
        ),
        "params": params,
    }

    return _kwargs


def _parse_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Optional[GenericResponse]:
    if response.status_code == 200:
        response_200 = GenericResponse.from_dict(response.json())

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
    latitude: str,
    longitude: str,
) -> Response[GenericResponse]:
    """Set simulated location

     Simulate a GPS location on the device.

    NOTE: the longitude parameter was historically misspelled `longtitude` on
    the wire. This spec fixes it to `longitude`; the go-ios server accepts
    `longitude` (and may keep `longtitude` as a deprecated alias).

    Args:
        udid (str):
        latitude (str):
        longitude (str):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[GenericResponse]
    """

    kwargs = _get_kwargs(
        udid=udid,
        latitude=latitude,
        longitude=longitude,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    latitude: str,
    longitude: str,
) -> Optional[GenericResponse]:
    """Set simulated location

     Simulate a GPS location on the device.

    NOTE: the longitude parameter was historically misspelled `longtitude` on
    the wire. This spec fixes it to `longitude`; the go-ios server accepts
    `longitude` (and may keep `longtitude` as a deprecated alias).

    Args:
        udid (str):
        latitude (str):
        longitude (str):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        GenericResponse
    """

    return sync_detailed(
        udid=udid,
        client=client,
        latitude=latitude,
        longitude=longitude,
    ).parsed


async def asyncio_detailed(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    latitude: str,
    longitude: str,
) -> Response[GenericResponse]:
    """Set simulated location

     Simulate a GPS location on the device.

    NOTE: the longitude parameter was historically misspelled `longtitude` on
    the wire. This spec fixes it to `longitude`; the go-ios server accepts
    `longitude` (and may keep `longtitude` as a deprecated alias).

    Args:
        udid (str):
        latitude (str):
        longitude (str):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[GenericResponse]
    """

    kwargs = _get_kwargs(
        udid=udid,
        latitude=latitude,
        longitude=longitude,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    latitude: str,
    longitude: str,
) -> Optional[GenericResponse]:
    """Set simulated location

     Simulate a GPS location on the device.

    NOTE: the longitude parameter was historically misspelled `longtitude` on
    the wire. This spec fixes it to `longitude`; the go-ios server accepts
    `longitude` (and may keep `longtitude` as a deprecated alias).

    Args:
        udid (str):
        latitude (str):
        longitude (str):

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
            latitude=latitude,
            longitude=longitude,
        )
    ).parsed
