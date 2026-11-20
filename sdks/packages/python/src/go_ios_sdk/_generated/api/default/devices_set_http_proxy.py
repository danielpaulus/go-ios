from http import HTTPStatus
from typing import Any, Optional, Union

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.devices_set_http_proxy_body import DevicesSetHttpProxyBody
from ...models.generic_response import GenericResponse
from ...types import Response


def _get_kwargs(
    udid: str,
    *,
    body: DevicesSetHttpProxyBody,
) -> dict[str, Any]:
    headers: dict[str, Any] = {}

    _kwargs: dict[str, Any] = {
        "method": "put",
        "url": "/api/v1/device/{udid}/httpproxy".format(
            udid=udid,
        ),
    }

    _kwargs["files"] = body.to_multipart()

    _kwargs["headers"] = headers
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
    body: DevicesSetHttpProxyBody,
) -> Response[GenericResponse]:
    """Set HTTP proxy (supervised)

     Configure a global HTTP proxy (CLI: `ios httpproxy`). Supervised: send
    multipart form-data with `host`, `port`, a `p12` supervisor identity and
    optional `user`/`pass`/`password` fields.

    Args:
        udid (str):
        body (DevicesSetHttpProxyBody):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[GenericResponse]
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
    body: DevicesSetHttpProxyBody,
) -> Optional[GenericResponse]:
    """Set HTTP proxy (supervised)

     Configure a global HTTP proxy (CLI: `ios httpproxy`). Supervised: send
    multipart form-data with `host`, `port`, a `p12` supervisor identity and
    optional `user`/`pass`/`password` fields.

    Args:
        udid (str):
        body (DevicesSetHttpProxyBody):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        GenericResponse
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
    body: DevicesSetHttpProxyBody,
) -> Response[GenericResponse]:
    """Set HTTP proxy (supervised)

     Configure a global HTTP proxy (CLI: `ios httpproxy`). Supervised: send
    multipart form-data with `host`, `port`, a `p12` supervisor identity and
    optional `user`/`pass`/`password` fields.

    Args:
        udid (str):
        body (DevicesSetHttpProxyBody):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[GenericResponse]
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
    body: DevicesSetHttpProxyBody,
) -> Optional[GenericResponse]:
    """Set HTTP proxy (supervised)

     Configure a global HTTP proxy (CLI: `ios httpproxy`). Supervised: send
    multipart form-data with `host`, `port`, a `p12` supervisor identity and
    optional `user`/`pass`/`password` fields.

    Args:
        udid (str):
        body (DevicesSetHttpProxyBody):

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
            body=body,
        )
    ).parsed
