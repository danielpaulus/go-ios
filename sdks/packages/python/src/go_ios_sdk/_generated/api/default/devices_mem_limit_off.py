from http import HTTPStatus
from typing import Any, Optional, Union

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.generic_response import GenericResponse
from ...models.mem_limit_request import MemLimitRequest
from ...models.mem_limit_result import MemLimitResult
from ...types import UNSET, Response, Unset


def _get_kwargs(
    udid: str,
    *,
    body: MemLimitRequest,
    process: Union[Unset, str] = UNSET,
) -> dict[str, Any]:
    headers: dict[str, Any] = {}

    params: dict[str, Any] = {}

    params["process"] = process

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "post",
        "url": "/api/v1/device/{udid}/memlimitoff".format(
            udid=udid,
        ),
        "params": params,
    }

    _kwargs["json"] = body.to_dict()

    headers["Content-Type"] = "application/json"

    _kwargs["headers"] = headers
    return _kwargs


def _parse_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Optional[Union[GenericResponse, MemLimitResult]]:
    if response.status_code == 200:
        response_200 = MemLimitResult.from_dict(response.json())

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
) -> Response[Union[GenericResponse, MemLimitResult]]:
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
    body: MemLimitRequest,
    process: Union[Unset, str] = UNSET,
) -> Response[Union[GenericResponse, MemLimitResult]]:
    """Waive memory limit

     Waive the memory limit for a process (CLI: `ios memlimitoff`). The process
    name may be given via the `process` query param or the JSON body.

    Args:
        udid (str):
        process (Union[Unset, str]):
        body (MemLimitRequest): `POST /device/{udid}/memlimitoff` request.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[GenericResponse, MemLimitResult]]
    """

    kwargs = _get_kwargs(
        udid=udid,
        body=body,
        process=process,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    body: MemLimitRequest,
    process: Union[Unset, str] = UNSET,
) -> Optional[Union[GenericResponse, MemLimitResult]]:
    """Waive memory limit

     Waive the memory limit for a process (CLI: `ios memlimitoff`). The process
    name may be given via the `process` query param or the JSON body.

    Args:
        udid (str):
        process (Union[Unset, str]):
        body (MemLimitRequest): `POST /device/{udid}/memlimitoff` request.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[GenericResponse, MemLimitResult]
    """

    return sync_detailed(
        udid=udid,
        client=client,
        body=body,
        process=process,
    ).parsed


async def asyncio_detailed(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    body: MemLimitRequest,
    process: Union[Unset, str] = UNSET,
) -> Response[Union[GenericResponse, MemLimitResult]]:
    """Waive memory limit

     Waive the memory limit for a process (CLI: `ios memlimitoff`). The process
    name may be given via the `process` query param or the JSON body.

    Args:
        udid (str):
        process (Union[Unset, str]):
        body (MemLimitRequest): `POST /device/{udid}/memlimitoff` request.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[GenericResponse, MemLimitResult]]
    """

    kwargs = _get_kwargs(
        udid=udid,
        body=body,
        process=process,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    body: MemLimitRequest,
    process: Union[Unset, str] = UNSET,
) -> Optional[Union[GenericResponse, MemLimitResult]]:
    """Waive memory limit

     Waive the memory limit for a process (CLI: `ios memlimitoff`). The process
    name may be given via the `process` query param or the JSON body.

    Args:
        udid (str):
        process (Union[Unset, str]):
        body (MemLimitRequest): `POST /device/{udid}/memlimitoff` request.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[GenericResponse, MemLimitResult]
    """

    return (
        await asyncio_detailed(
            udid=udid,
            client=client,
            body=body,
            process=process,
        )
    ).parsed
