from http import HTTPStatus
from typing import Any, Optional, Union

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.ax_audit_issue import AXAuditIssue
from ...models.generic_response import GenericResponse
from ...types import UNSET, Response, Unset


def _get_kwargs(
    udid: str,
    *,
    timeout: Union[Unset, int] = UNSET,
) -> dict[str, Any]:
    params: dict[str, Any] = {}

    params["timeout"] = timeout

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "post",
        "url": "/api/v1/device/{udid}/ax/audit".format(
            udid=udid,
        ),
        "params": params,
    }

    return _kwargs


def _parse_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Optional[Union[GenericResponse, list["AXAuditIssue"]]]:
    if response.status_code == 200:
        response_200 = []
        _response_200 = response.json()
        for response_200_item_data in _response_200:
            response_200_item = AXAuditIssue.from_dict(response_200_item_data)

            response_200.append(response_200_item)

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
) -> Response[Union[GenericResponse, list["AXAuditIssue"]]]:
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
    timeout: Union[Unset, int] = UNSET,
) -> Response[Union[GenericResponse, list["AXAuditIssue"]]]:
    """Run accessibility audit

     Run the accessibility audit against the focused app and return the issues
    found (CLI: `ios ax audit`). Bounded by `timeout` (seconds, default 60).

    Args:
        udid (str):
        timeout (Union[Unset, int]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[GenericResponse, list['AXAuditIssue']]]
    """

    kwargs = _get_kwargs(
        udid=udid,
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
    timeout: Union[Unset, int] = UNSET,
) -> Optional[Union[GenericResponse, list["AXAuditIssue"]]]:
    """Run accessibility audit

     Run the accessibility audit against the focused app and return the issues
    found (CLI: `ios ax audit`). Bounded by `timeout` (seconds, default 60).

    Args:
        udid (str):
        timeout (Union[Unset, int]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[GenericResponse, list['AXAuditIssue']]
    """

    return sync_detailed(
        udid=udid,
        client=client,
        timeout=timeout,
    ).parsed


async def asyncio_detailed(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    timeout: Union[Unset, int] = UNSET,
) -> Response[Union[GenericResponse, list["AXAuditIssue"]]]:
    """Run accessibility audit

     Run the accessibility audit against the focused app and return the issues
    found (CLI: `ios ax audit`). Bounded by `timeout` (seconds, default 60).

    Args:
        udid (str):
        timeout (Union[Unset, int]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[GenericResponse, list['AXAuditIssue']]]
    """

    kwargs = _get_kwargs(
        udid=udid,
        timeout=timeout,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    timeout: Union[Unset, int] = UNSET,
) -> Optional[Union[GenericResponse, list["AXAuditIssue"]]]:
    """Run accessibility audit

     Run the accessibility audit against the focused app and return the issues
    found (CLI: `ios ax audit`). Bounded by `timeout` (seconds, default 60).

    Args:
        udid (str):
        timeout (Union[Unset, int]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[GenericResponse, list['AXAuditIssue']]
    """

    return (
        await asyncio_detailed(
            udid=udid,
            client=client,
            timeout=timeout,
        )
    ).parsed
