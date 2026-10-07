from http import HTTPStatus
from typing import Any, Optional, Union

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.generic_response import GenericResponse
from ...models.supervision_cert import SupervisionCert
from ...types import Response


def _get_kwargs() -> dict[str, Any]:
    _kwargs: dict[str, Any] = {
        "method": "post",
        "url": "/api/v1/prepare/create-cert",
    }

    return _kwargs


def _parse_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Optional[Union[GenericResponse, SupervisionCert]]:
    if response.status_code == 200:
        response_200 = SupervisionCert.from_dict(response.json())

        return response_200

    if response.status_code == 401:
        response_401 = GenericResponse.from_dict(response.json())

        return response_401

    if response.status_code == 500:
        response_500 = GenericResponse.from_dict(response.json())

        return response_500

    if client.raise_on_unexpected_status:
        raise errors.UnexpectedStatus(response.status_code, response.content)
    else:
        return None


def _build_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Response[Union[GenericResponse, SupervisionCert]]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    *,
    client: Union[AuthenticatedClient, Client],
) -> Response[Union[GenericResponse, SupervisionCert]]:
    """Generate a supervision certificate

     Generate a self-signed supervision identity (CLI: `ios prepare create-cert`)
    and return the DER (base64) and PEM for both the certificate and private key.
    Host-scoped (device-free).

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[GenericResponse, SupervisionCert]]
    """

    kwargs = _get_kwargs()

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    *,
    client: Union[AuthenticatedClient, Client],
) -> Optional[Union[GenericResponse, SupervisionCert]]:
    """Generate a supervision certificate

     Generate a self-signed supervision identity (CLI: `ios prepare create-cert`)
    and return the DER (base64) and PEM for both the certificate and private key.
    Host-scoped (device-free).

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[GenericResponse, SupervisionCert]
    """

    return sync_detailed(
        client=client,
    ).parsed


async def asyncio_detailed(
    *,
    client: Union[AuthenticatedClient, Client],
) -> Response[Union[GenericResponse, SupervisionCert]]:
    """Generate a supervision certificate

     Generate a self-signed supervision identity (CLI: `ios prepare create-cert`)
    and return the DER (base64) and PEM for both the certificate and private key.
    Host-scoped (device-free).

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[GenericResponse, SupervisionCert]]
    """

    kwargs = _get_kwargs()

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    *,
    client: Union[AuthenticatedClient, Client],
) -> Optional[Union[GenericResponse, SupervisionCert]]:
    """Generate a supervision certificate

     Generate a self-signed supervision identity (CLI: `ios prepare create-cert`)
    and return the DER (base64) and PEM for both the certificate and private key.
    Host-scoped (device-free).

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[GenericResponse, SupervisionCert]
    """

    return (
        await asyncio_detailed(
            client=client,
        )
    ).parsed
