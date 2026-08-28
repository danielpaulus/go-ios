from http import HTTPStatus
from typing import Any, Optional, Union

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.file_domain_type_1 import FileDomainType1
from ...models.file_push_result import FilePushResult
from ...models.generic_response import GenericResponse
from ...types import UNSET, Response, Unset


def _get_kwargs(
    udid: str,
    *,
    body: Any,
    domain: Union[FileDomainType1, str],
    identifier: Union[Unset, str] = UNSET,
    remote: str,
) -> dict[str, Any]:
    headers: dict[str, Any] = {}

    params: dict[str, Any] = {}

    json_domain: str
    if isinstance(domain, FileDomainType1):
        json_domain = domain.value
    else:
        json_domain = domain
    params["domain"] = json_domain

    params["identifier"] = identifier

    params["remote"] = remote

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "post",
        "url": "/api/v1/device/{udid}/files/push".format(
            udid=udid,
        ),
        "params": params,
    }

    _kwargs["content"] = body.payload

    headers["Content-Type"] = "application/octet-stream"

    _kwargs["headers"] = headers
    return _kwargs


def _parse_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Optional[Union[FilePushResult, GenericResponse]]:
    if response.status_code == 200:
        response_200 = FilePushResult.from_dict(response.json())

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
) -> Response[Union[FilePushResult, GenericResponse]]:
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
    body: Any,
    domain: Union[FileDomainType1, str],
    identifier: Union[Unset, str] = UNSET,
    remote: str,
) -> Response[Union[FilePushResult, GenericResponse]]:
    """Push file

     Upload the request body to a device path (CLI: `ios file push`). A
    `Content-Length` header is required.

    Args:
        udid (str):
        domain (Union[FileDomainType1, str]): Domain of the on-device file service.
        identifier (Union[Unset, str]):
        remote (str):
        body (Any):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[FilePushResult, GenericResponse]]
    """

    kwargs = _get_kwargs(
        udid=udid,
        body=body,
        domain=domain,
        identifier=identifier,
        remote=remote,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    body: Any,
    domain: Union[FileDomainType1, str],
    identifier: Union[Unset, str] = UNSET,
    remote: str,
) -> Optional[Union[FilePushResult, GenericResponse]]:
    """Push file

     Upload the request body to a device path (CLI: `ios file push`). A
    `Content-Length` header is required.

    Args:
        udid (str):
        domain (Union[FileDomainType1, str]): Domain of the on-device file service.
        identifier (Union[Unset, str]):
        remote (str):
        body (Any):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[FilePushResult, GenericResponse]
    """

    return sync_detailed(
        udid=udid,
        client=client,
        body=body,
        domain=domain,
        identifier=identifier,
        remote=remote,
    ).parsed


async def asyncio_detailed(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    body: Any,
    domain: Union[FileDomainType1, str],
    identifier: Union[Unset, str] = UNSET,
    remote: str,
) -> Response[Union[FilePushResult, GenericResponse]]:
    """Push file

     Upload the request body to a device path (CLI: `ios file push`). A
    `Content-Length` header is required.

    Args:
        udid (str):
        domain (Union[FileDomainType1, str]): Domain of the on-device file service.
        identifier (Union[Unset, str]):
        remote (str):
        body (Any):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[FilePushResult, GenericResponse]]
    """

    kwargs = _get_kwargs(
        udid=udid,
        body=body,
        domain=domain,
        identifier=identifier,
        remote=remote,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    body: Any,
    domain: Union[FileDomainType1, str],
    identifier: Union[Unset, str] = UNSET,
    remote: str,
) -> Optional[Union[FilePushResult, GenericResponse]]:
    """Push file

     Upload the request body to a device path (CLI: `ios file push`). A
    `Content-Length` header is required.

    Args:
        udid (str):
        domain (Union[FileDomainType1, str]): Domain of the on-device file service.
        identifier (Union[Unset, str]):
        remote (str):
        body (Any):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[FilePushResult, GenericResponse]
    """

    return (
        await asyncio_detailed(
            udid=udid,
            client=client,
            body=body,
            domain=domain,
            identifier=identifier,
            remote=remote,
        )
    ).parsed
