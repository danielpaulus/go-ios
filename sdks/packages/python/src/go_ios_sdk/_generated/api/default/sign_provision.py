from http import HTTPStatus
from typing import Any, Optional, Union

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.generic_response import GenericResponse
from ...models.provisioning_result import ProvisioningResult
from ...models.sign_provision_body import SignProvisionBody
from ...types import Response


def _get_kwargs(
    *,
    body: SignProvisionBody,
) -> dict[str, Any]:
    headers: dict[str, Any] = {}

    _kwargs: dict[str, Any] = {
        "method": "post",
        "url": "/api/v1/sign/provision",
    }

    _kwargs["files"] = body.to_multipart()

    _kwargs["headers"] = headers
    return _kwargs


def _parse_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Optional[Union[GenericResponse, ProvisioningResult]]:
    if response.status_code == 200:
        response_200 = ProvisioningResult.from_dict(response.json())

        return response_200

    if response.status_code == 400:
        response_400 = GenericResponse.from_dict(response.json())

        return response_400

    if response.status_code == 401:
        response_401 = GenericResponse.from_dict(response.json())

        return response_401

    if response.status_code == 500:
        response_500 = GenericResponse.from_dict(response.json())

        return response_500

    if response.status_code == 502:
        response_502 = GenericResponse.from_dict(response.json())

        return response_502

    if client.raise_on_unexpected_status:
        raise errors.UnexpectedStatus(response.status_code, response.content)
    else:
        return None


def _build_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Response[Union[GenericResponse, ProvisioningResult]]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    *,
    client: Union[AuthenticatedClient, Client],
    body: SignProvisionBody,
) -> Response[Union[GenericResponse, ProvisioningResult]]:
    """Create a provisioning profile + P12

     Create a bundle id, development certificate and provisioning profile via App
    Store Connect and return both artifacts base64-encoded in a JSON envelope.
    The target device udid is supplied as a form field. Host-scoped.

    Args:
        body (SignProvisionBody):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[GenericResponse, ProvisioningResult]]
    """

    kwargs = _get_kwargs(
        body=body,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    *,
    client: Union[AuthenticatedClient, Client],
    body: SignProvisionBody,
) -> Optional[Union[GenericResponse, ProvisioningResult]]:
    """Create a provisioning profile + P12

     Create a bundle id, development certificate and provisioning profile via App
    Store Connect and return both artifacts base64-encoded in a JSON envelope.
    The target device udid is supplied as a form field. Host-scoped.

    Args:
        body (SignProvisionBody):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[GenericResponse, ProvisioningResult]
    """

    return sync_detailed(
        client=client,
        body=body,
    ).parsed


async def asyncio_detailed(
    *,
    client: Union[AuthenticatedClient, Client],
    body: SignProvisionBody,
) -> Response[Union[GenericResponse, ProvisioningResult]]:
    """Create a provisioning profile + P12

     Create a bundle id, development certificate and provisioning profile via App
    Store Connect and return both artifacts base64-encoded in a JSON envelope.
    The target device udid is supplied as a form field. Host-scoped.

    Args:
        body (SignProvisionBody):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[GenericResponse, ProvisioningResult]]
    """

    kwargs = _get_kwargs(
        body=body,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    *,
    client: Union[AuthenticatedClient, Client],
    body: SignProvisionBody,
) -> Optional[Union[GenericResponse, ProvisioningResult]]:
    """Create a provisioning profile + P12

     Create a bundle id, development certificate and provisioning profile via App
    Store Connect and return both artifacts base64-encoded in a JSON envelope.
    The target device udid is supplied as a form field. Host-scoped.

    Args:
        body (SignProvisionBody):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[GenericResponse, ProvisioningResult]
    """

    return (
        await asyncio_detailed(
            client=client,
            body=body,
        )
    ).parsed
