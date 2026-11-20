from http import HTTPStatus
from typing import Any, Optional, Union

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.generic_response import GenericResponse
from ...models.language_configuration import LanguageConfiguration
from ...models.set_language_request import SetLanguageRequest
from ...types import Response


def _get_kwargs(
    udid: str,
    *,
    body: SetLanguageRequest,
) -> dict[str, Any]:
    headers: dict[str, Any] = {}

    _kwargs: dict[str, Any] = {
        "method": "put",
        "url": "/api/v1/device/{udid}/lang".format(
            udid=udid,
        ),
    }

    _kwargs["json"] = body.to_dict()

    headers["Content-Type"] = "application/json"

    _kwargs["headers"] = headers
    return _kwargs


def _parse_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Optional[Union[GenericResponse, LanguageConfiguration]]:
    if response.status_code == 200:
        response_200 = LanguageConfiguration.from_dict(response.json())

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
) -> Response[Union[GenericResponse, LanguageConfiguration]]:
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
    body: SetLanguageRequest,
) -> Response[Union[GenericResponse, LanguageConfiguration]]:
    """Set language

     Set the device language and/or locale (CLI: `ios lang --setlang --setlocale`).
    Returns the resulting configuration.

    Args:
        udid (str):
        body (SetLanguageRequest): `PUT /device/{udid}/lang` request.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[GenericResponse, LanguageConfiguration]]
    """

    kwargs = _get_kwargs(
        udid=udid,
        body=body,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    body: SetLanguageRequest,
) -> Optional[Union[GenericResponse, LanguageConfiguration]]:
    """Set language

     Set the device language and/or locale (CLI: `ios lang --setlang --setlocale`).
    Returns the resulting configuration.

    Args:
        udid (str):
        body (SetLanguageRequest): `PUT /device/{udid}/lang` request.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[GenericResponse, LanguageConfiguration]
    """

    return sync_detailed(
        udid=udid,
        client=client,
        body=body,
    ).parsed


async def asyncio_detailed(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    body: SetLanguageRequest,
) -> Response[Union[GenericResponse, LanguageConfiguration]]:
    """Set language

     Set the device language and/or locale (CLI: `ios lang --setlang --setlocale`).
    Returns the resulting configuration.

    Args:
        udid (str):
        body (SetLanguageRequest): `PUT /device/{udid}/lang` request.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[GenericResponse, LanguageConfiguration]]
    """

    kwargs = _get_kwargs(
        udid=udid,
        body=body,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    body: SetLanguageRequest,
) -> Optional[Union[GenericResponse, LanguageConfiguration]]:
    """Set language

     Set the device language and/or locale (CLI: `ios lang --setlang --setlocale`).
    Returns the resulting configuration.

    Args:
        udid (str):
        body (SetLanguageRequest): `PUT /device/{udid}/lang` request.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[GenericResponse, LanguageConfiguration]
    """

    return (
        await asyncio_detailed(
            udid=udid,
            client=client,
            body=body,
        )
    ).parsed
