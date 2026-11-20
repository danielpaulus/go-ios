from http import HTTPStatus
from typing import Any, Optional, Union

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.devices_pair_body import DevicesPairBody
from ...models.generic_response import GenericResponse
from ...types import UNSET, Response, Unset


def _get_kwargs(
    udid: str,
    *,
    body: DevicesPairBody,
    supervised: bool,
    supervision_password: Union[Unset, str] = UNSET,
) -> dict[str, Any]:
    headers: dict[str, Any] = {}
    if not isinstance(supervision_password, Unset):
        headers["Supervision-Password"] = supervision_password

    params: dict[str, Any] = {}

    params["supervised"] = supervised

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "post",
        "url": "/api/v1/device/{udid}/pair".format(
            udid=udid,
        ),
        "params": params,
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

    if response.status_code == 401:
        response_401 = GenericResponse.from_dict(response.json())

        return response_401

    if response.status_code == 404:
        response_404 = GenericResponse.from_dict(response.json())

        return response_404

    if response.status_code == 422:
        response_422 = GenericResponse.from_dict(response.json())

        return response_422

    if response.status_code == 423:
        response_423 = GenericResponse.from_dict(response.json())

        return response_423

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
    body: DevicesPairBody,
    supervised: bool,
    supervision_password: Union[Unset, str] = UNSET,
) -> Response[GenericResponse]:
    """Pair device

     Pair with the device.

    For a supervised pairing (`supervised=true`) upload the supervision
    identity as `p12file` (multipart) and supply the passphrase in the
    `Supervision-Password` header.

    Returns `423` when the device is locked and pairing cannot proceed.

    Args:
        udid (str):
        supervised (bool):
        supervision_password (Union[Unset, str]):
        body (DevicesPairBody):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[GenericResponse]
    """

    kwargs = _get_kwargs(
        udid=udid,
        body=body,
        supervised=supervised,
        supervision_password=supervision_password,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    body: DevicesPairBody,
    supervised: bool,
    supervision_password: Union[Unset, str] = UNSET,
) -> Optional[GenericResponse]:
    """Pair device

     Pair with the device.

    For a supervised pairing (`supervised=true`) upload the supervision
    identity as `p12file` (multipart) and supply the passphrase in the
    `Supervision-Password` header.

    Returns `423` when the device is locked and pairing cannot proceed.

    Args:
        udid (str):
        supervised (bool):
        supervision_password (Union[Unset, str]):
        body (DevicesPairBody):

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
        supervised=supervised,
        supervision_password=supervision_password,
    ).parsed


async def asyncio_detailed(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    body: DevicesPairBody,
    supervised: bool,
    supervision_password: Union[Unset, str] = UNSET,
) -> Response[GenericResponse]:
    """Pair device

     Pair with the device.

    For a supervised pairing (`supervised=true`) upload the supervision
    identity as `p12file` (multipart) and supply the passphrase in the
    `Supervision-Password` header.

    Returns `423` when the device is locked and pairing cannot proceed.

    Args:
        udid (str):
        supervised (bool):
        supervision_password (Union[Unset, str]):
        body (DevicesPairBody):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[GenericResponse]
    """

    kwargs = _get_kwargs(
        udid=udid,
        body=body,
        supervised=supervised,
        supervision_password=supervision_password,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    udid: str,
    *,
    client: Union[AuthenticatedClient, Client],
    body: DevicesPairBody,
    supervised: bool,
    supervision_password: Union[Unset, str] = UNSET,
) -> Optional[GenericResponse]:
    """Pair device

     Pair with the device.

    For a supervised pairing (`supervised=true`) upload the supervision
    identity as `p12file` (multipart) and supply the passphrase in the
    `Supervision-Password` header.

    Returns `423` when the device is locked and pairing cannot proceed.

    Args:
        udid (str):
        supervised (bool):
        supervision_password (Union[Unset, str]):
        body (DevicesPairBody):

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
            supervised=supervised,
            supervision_password=supervision_password,
        )
    ).parsed
