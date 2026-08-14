from collections.abc import Mapping
from typing import (
    Any,
    TypeVar,
    Union,
)

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

T = TypeVar("T", bound="UIAPIRequest")


@_attrs_define
class UIAPIRequest:
    """`POST /device/{udid}/ui/api` request — raw passthrough to the backend
    (`uidriver.APIRequest`). For WDA supply `method`/`path`/`body`; for DeviceKit
    supply `rpcMethod`/`rpcParams`.

        Attributes:
            method (Union[Unset, str]): HTTP method for a WDA passthrough (defaults to GET).
            path (Union[Unset, str]): HTTP path for a WDA passthrough (required for the wda backend).
            body (Union[Unset, str]): Raw HTTP request body for a WDA passthrough (base64 bytes on the wire).
            rpc_method (Union[Unset, str]): JSON-RPC method name for a DeviceKit passthrough.
            rpc_params (Union[Unset, Any]): JSON-RPC params for a DeviceKit passthrough.
    """

    method: Union[Unset, str] = UNSET
    path: Union[Unset, str] = UNSET
    body: Union[Unset, str] = UNSET
    rpc_method: Union[Unset, str] = UNSET
    rpc_params: Union[Unset, Any] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        method = self.method

        path = self.path

        body = self.body

        rpc_method = self.rpc_method

        rpc_params = self.rpc_params

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update({})
        if method is not UNSET:
            field_dict["method"] = method
        if path is not UNSET:
            field_dict["path"] = path
        if body is not UNSET:
            field_dict["body"] = body
        if rpc_method is not UNSET:
            field_dict["rpcMethod"] = rpc_method
        if rpc_params is not UNSET:
            field_dict["rpcParams"] = rpc_params

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        method = d.pop("method", UNSET)

        path = d.pop("path", UNSET)

        body = d.pop("body", UNSET)

        rpc_method = d.pop("rpcMethod", UNSET)

        rpc_params = d.pop("rpcParams", UNSET)

        uiapi_request = cls(
            method=method,
            path=path,
            body=body,
            rpc_method=rpc_method,
            rpc_params=rpc_params,
        )

        uiapi_request.additional_properties = d
        return uiapi_request

    @property
    def additional_keys(self) -> list[str]:
        return list(self.additional_properties.keys())

    def __getitem__(self, key: str) -> Any:
        return self.additional_properties[key]

    def __setitem__(self, key: str, value: Any) -> None:
        self.additional_properties[key] = value

    def __delitem__(self, key: str) -> None:
        del self.additional_properties[key]

    def __contains__(self, key: str) -> bool:
        return key in self.additional_properties
