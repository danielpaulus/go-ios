from http import HTTPStatus
from typing import Any, Optional, Union

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.generic_response import GenericResponse
from ...models.web_inspector_eval_request import WebInspectorEvalRequest
from ...models.web_inspector_eval_result import WebInspectorEvalResult
from ...types import Response


def _get_kwargs(
    udid: str,
    *,
    body: WebInspectorEvalRequest,
) -> dict[str, Any]:
    headers: dict[str, Any] = {}

    _kwargs: dict[str, Any] = {
        "method": "post",
        "url": "/api/v1/device/{udid}/webinspector/eval".format(
            udid=udid,
        ),
    }

    _kwargs["json"] = body.to_dict()

    headers["Content-Type"] = "application/json"

    _kwargs["headers"] = headers
    return _kwargs


def _parse_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Optional[Union["GenericResponse", GenericResponse, WebInspectorEvalResult]]:
    if response.status_code == 200:
        response_200 = WebInspectorEvalResult.from_dict(response.json())

        return response_200

    if response.status_code == 400:
        response_400 = GenericResponse.from_dict(response.json())

        return response_400

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

    if response.status_code == 424:
        response_424 = GenericResponse.from_dict(response.json())

        return response_424

    if response.status_code == 500:
        response_500 = GenericResponse.from_dict(response.json())

        return response_500

    if client.raise_on_unexpected_status:
        raise errors.UnexpectedStatus(response.status_code, response.content)
    else:
        return None


def _build_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Response[Union["GenericResponse", GenericResponse, WebInspectorEvalResult]]:
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
    body: WebInspectorEvalRequest,
) -> Response[Union["GenericResponse", GenericResponse, WebInspectorEvalResult]]:
    """Evaluate JavaScript in a page

     Evaluate JavaScript in an inspectable page and return the result
    (CLI: `ios webinspector eval`). `404` when no matching page exists.

    Args:
        udid (str):
        body (WebInspectorEvalRequest): `POST /device/{udid}/webinspector/eval` request body.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union['GenericResponse', GenericResponse, WebInspectorEvalResult]]
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
    body: WebInspectorEvalRequest,
) -> Optional[Union["GenericResponse", GenericResponse, WebInspectorEvalResult]]:
    """Evaluate JavaScript in a page

     Evaluate JavaScript in an inspectable page and return the result
    (CLI: `ios webinspector eval`). `404` when no matching page exists.

    Args:
        udid (str):
        body (WebInspectorEvalRequest): `POST /device/{udid}/webinspector/eval` request body.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union['GenericResponse', GenericResponse, WebInspectorEvalResult]
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
    body: WebInspectorEvalRequest,
) -> Response[Union["GenericResponse", GenericResponse, WebInspectorEvalResult]]:
    """Evaluate JavaScript in a page

     Evaluate JavaScript in an inspectable page and return the result
    (CLI: `ios webinspector eval`). `404` when no matching page exists.

    Args:
        udid (str):
        body (WebInspectorEvalRequest): `POST /device/{udid}/webinspector/eval` request body.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union['GenericResponse', GenericResponse, WebInspectorEvalResult]]
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
    body: WebInspectorEvalRequest,
) -> Optional[Union["GenericResponse", GenericResponse, WebInspectorEvalResult]]:
    """Evaluate JavaScript in a page

     Evaluate JavaScript in an inspectable page and return the result
    (CLI: `ios webinspector eval`). `404` when no matching page exists.

    Args:
        udid (str):
        body (WebInspectorEvalRequest): `POST /device/{udid}/webinspector/eval` request body.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union['GenericResponse', GenericResponse, WebInspectorEvalResult]
    """

    return (
        await asyncio_detailed(
            udid=udid,
            client=client,
            body=body,
        )
    ).parsed
