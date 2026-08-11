from http import HTTPStatus
from typing import Any, Optional, Union

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.ax_enabled_request import AXEnabledRequest
from ...models.generic_response import GenericResponse
from ...models.zoom_touch_state import ZoomTouchState
from ...types import UNSET, Response, Unset


def _get_kwargs(
    udid: str,
    *,
    body: AXEnabledRequest,
    enabled: Union[Unset, bool] = UNSET,
) -> dict[str, Any]:
    headers: dict[str, Any] = {}

    params: dict[str, Any] = {}

    params["enabled"] = enabled

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "put",
        "url": "/api/v1/device/{udid}/zoom".format(
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
) -> Optional[Union[GenericResponse, ZoomTouchState]]:
    if response.status_code == 200:
        response_200 = ZoomTouchState.from_dict(response.json())

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
) -> Response[Union[GenericResponse, ZoomTouchState]]:
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
    body: AXEnabledRequest,
    enabled: Union[Unset, bool] = UNSET,
) -> Response[Union[GenericResponse, ZoomTouchState]]:
    """Set ZoomTouch state

     Enable/disable ZoomTouch (CLI: `ios zoomtouch enable|disable`). The desired
    state comes from the JSON body or the `enabled` query param.

    Args:
        udid (str):
        enabled (Union[Unset, bool]):
        body (AXEnabledRequest): Body for the accessibility toggle PUTs (`/voiceover`, `/zoom`).
            The desired
            state may also be supplied as an `enabled` query param; a parseable body wins.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[GenericResponse, ZoomTouchState]]
    """

    kwargs = _get_kwargs(
        udid=udid,
        body=body,
        enabled=enabled,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    body: AXEnabledRequest,
    enabled: Union[Unset, bool] = UNSET,
) -> Optional[Union[GenericResponse, ZoomTouchState]]:
    """Set ZoomTouch state

     Enable/disable ZoomTouch (CLI: `ios zoomtouch enable|disable`). The desired
    state comes from the JSON body or the `enabled` query param.

    Args:
        udid (str):
        enabled (Union[Unset, bool]):
        body (AXEnabledRequest): Body for the accessibility toggle PUTs (`/voiceover`, `/zoom`).
            The desired
            state may also be supplied as an `enabled` query param; a parseable body wins.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[GenericResponse, ZoomTouchState]
    """

    return sync_detailed(
        udid=udid,
        client=client,
        body=body,
        enabled=enabled,
    ).parsed


async def asyncio_detailed(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    body: AXEnabledRequest,
    enabled: Union[Unset, bool] = UNSET,
) -> Response[Union[GenericResponse, ZoomTouchState]]:
    """Set ZoomTouch state

     Enable/disable ZoomTouch (CLI: `ios zoomtouch enable|disable`). The desired
    state comes from the JSON body or the `enabled` query param.

    Args:
        udid (str):
        enabled (Union[Unset, bool]):
        body (AXEnabledRequest): Body for the accessibility toggle PUTs (`/voiceover`, `/zoom`).
            The desired
            state may also be supplied as an `enabled` query param; a parseable body wins.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[GenericResponse, ZoomTouchState]]
    """

    kwargs = _get_kwargs(
        udid=udid,
        body=body,
        enabled=enabled,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    body: AXEnabledRequest,
    enabled: Union[Unset, bool] = UNSET,
) -> Optional[Union[GenericResponse, ZoomTouchState]]:
    """Set ZoomTouch state

     Enable/disable ZoomTouch (CLI: `ios zoomtouch enable|disable`). The desired
    state comes from the JSON body or the `enabled` query param.

    Args:
        udid (str):
        enabled (Union[Unset, bool]):
        body (AXEnabledRequest): Body for the accessibility toggle PUTs (`/voiceover`, `/zoom`).
            The desired
            state may also be supplied as an `enabled` query param; a parseable body wins.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[GenericResponse, ZoomTouchState]
    """

    return (
        await asyncio_detailed(
            udid=udid,
            client=client,
            body=body,
            enabled=enabled,
        )
    ).parsed
