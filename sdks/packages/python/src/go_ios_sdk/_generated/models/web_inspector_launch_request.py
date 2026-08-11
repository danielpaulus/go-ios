from collections.abc import Mapping
from typing import (
    Any,
    TypeVar,
    Union,
)

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

T = TypeVar("T", bound="WebInspectorLaunchRequest")


@_attrs_define
class WebInspectorLaunchRequest:
    """`POST /device/{udid}/webinspector/launch` request body.

    Attributes:
        url (Union[Unset, str]): URL to open. May alternatively be supplied as the `url` query param.
        bundle_id (Union[Unset, str]): Bundle id to open the URL in. Defaults to Safari.
    """

    url: Union[Unset, str] = UNSET
    bundle_id: Union[Unset, str] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        url = self.url

        bundle_id = self.bundle_id

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update({})
        if url is not UNSET:
            field_dict["url"] = url
        if bundle_id is not UNSET:
            field_dict["bundleId"] = bundle_id

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        url = d.pop("url", UNSET)

        bundle_id = d.pop("bundleId", UNSET)

        web_inspector_launch_request = cls(
            url=url,
            bundle_id=bundle_id,
        )

        web_inspector_launch_request.additional_properties = d
        return web_inspector_launch_request

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
