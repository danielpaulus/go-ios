from collections.abc import Mapping
from typing import Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field

T = TypeVar("T", bound="WebInspectorLaunchResult")


@_attrs_define
class WebInspectorLaunchResult:
    """`POST /device/{udid}/webinspector/launch` — result of opening a URL.

    Attributes:
        bundle_id (str): Bundle id the page was opened in.
        url (str): The resolved current URL after navigation.
        title (str): The page title after navigation.
    """

    bundle_id: str
    url: str
    title: str
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        bundle_id = self.bundle_id

        url = self.url

        title = self.title

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "bundleId": bundle_id,
                "url": url,
                "title": title,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        bundle_id = d.pop("bundleId")

        url = d.pop("url")

        title = d.pop("title")

        web_inspector_launch_result = cls(
            bundle_id=bundle_id,
            url=url,
            title=title,
        )

        web_inspector_launch_result.additional_properties = d
        return web_inspector_launch_result

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
