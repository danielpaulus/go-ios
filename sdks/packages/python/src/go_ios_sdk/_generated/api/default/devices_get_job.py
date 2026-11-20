from http import HTTPStatus
from typing import Any, Optional, Union

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.generic_response import GenericResponse
from ...models.job import Job
from ...types import Response


def _get_kwargs(
    udid: str,
    id: str,
) -> dict[str, Any]:
    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/api/v1/device/{udid}/jobs/{id}".format(
            udid=udid,
            id=id,
        ),
    }

    return _kwargs


def _parse_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Optional[Union["GenericResponse", GenericResponse, Job]]:
    if response.status_code == 200:
        response_200 = Job.from_dict(response.json())

        return response_200

    if response.status_code == 401:
        response_401 = GenericResponse.from_dict(response.json())

        return response_401

    if response.status_code == 404:

        def _parse_response_404(data: object) -> "GenericResponse":
            try:
                if not isinstance(data, dict):
                    raise TypeError()
                response_404_type_0 = GenericResponse.from_dict(data)

                return response_404_type_0
            except:  # noqa: E722
                pass
            if not isinstance(data, dict):
                raise TypeError()
            response_404_type_1 = GenericResponse.from_dict(data)

            return response_404_type_1

        response_404 = _parse_response_404(response.json())

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
) -> Response[Union["GenericResponse", GenericResponse, Job]]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    udid: str,
    id: str,
    *,
    client: Union[AuthenticatedClient, Client],
) -> Response[Union["GenericResponse", GenericResponse, Job]]:
    """Get job

     Get a job's status. Returns `404` for an unknown job on this device.

    Args:
        udid (str):
        id (str):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union['GenericResponse', GenericResponse, Job]]
    """

    kwargs = _get_kwargs(
        udid=udid,
        id=id,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    udid: str,
    id: str,
    *,
    client: Union[AuthenticatedClient, Client],
) -> Optional[Union["GenericResponse", GenericResponse, Job]]:
    """Get job

     Get a job's status. Returns `404` for an unknown job on this device.

    Args:
        udid (str):
        id (str):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union['GenericResponse', GenericResponse, Job]
    """

    return sync_detailed(
        udid=udid,
        id=id,
        client=client,
    ).parsed


async def asyncio_detailed(
    udid: str,
    id: str,
    *,
    client: Union[AuthenticatedClient, Client],
) -> Response[Union["GenericResponse", GenericResponse, Job]]:
    """Get job

     Get a job's status. Returns `404` for an unknown job on this device.

    Args:
        udid (str):
        id (str):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union['GenericResponse', GenericResponse, Job]]
    """

    kwargs = _get_kwargs(
        udid=udid,
        id=id,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    udid: str,
    id: str,
    *,
    client: Union[AuthenticatedClient, Client],
) -> Optional[Union["GenericResponse", GenericResponse, Job]]:
    """Get job

     Get a job's status. Returns `404` for an unknown job on this device.

    Args:
        udid (str):
        id (str):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union['GenericResponse', GenericResponse, Job]
    """

    return (
        await asyncio_detailed(
            udid=udid,
            id=id,
            client=client,
        )
    ).parsed
