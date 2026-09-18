from collections.abc import Mapping
from typing import (
    Any,
    TypeVar,
    Union,
)

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

T = TypeVar("T", bound="WebInspectorEvalRequest")


@_attrs_define
class WebInspectorEvalRequest:
    """`POST /device/{udid}/webinspector/eval` request body.

    Attributes:
        script (str): JavaScript source to evaluate. Required.
        page (Union[Unset, str]): Inspectable page key. When empty the first matching web/javascript page
            (optionally scoped by `bundleId`) is used.
        bundle_id (Union[Unset, str]): Optional bundle id to scope page selection.
    """

    script: str
    page: Union[Unset, str] = UNSET
    bundle_id: Union[Unset, str] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        script = self.script

        page = self.page

        bundle_id = self.bundle_id

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "script": script,
            }
        )
        if page is not UNSET:
            field_dict["page"] = page
        if bundle_id is not UNSET:
            field_dict["bundleId"] = bundle_id

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        script = d.pop("script")

        page = d.pop("page", UNSET)

        bundle_id = d.pop("bundleId", UNSET)

        web_inspector_eval_request = cls(
            script=script,
            page=page,
            bundle_id=bundle_id,
        )

        web_inspector_eval_request.additional_properties = d
        return web_inspector_eval_request

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
