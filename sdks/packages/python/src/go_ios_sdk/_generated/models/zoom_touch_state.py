from collections.abc import Mapping
from typing import Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field

T = TypeVar("T", bound="ZoomTouchState")


@_attrs_define
class ZoomTouchState:
    """`GET|PUT /device/{udid}/zoom` — ZoomTouch enabled state.

    Attributes:
        zoom_touch_enabled (bool):
    """

    zoom_touch_enabled: bool
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        zoom_touch_enabled = self.zoom_touch_enabled

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "ZoomTouchEnabled": zoom_touch_enabled,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        zoom_touch_enabled = d.pop("ZoomTouchEnabled")

        zoom_touch_state = cls(
            zoom_touch_enabled=zoom_touch_enabled,
        )

        zoom_touch_state.additional_properties = d
        return zoom_touch_state

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
