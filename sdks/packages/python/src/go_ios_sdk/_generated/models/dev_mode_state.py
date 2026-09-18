from collections.abc import Mapping
from typing import Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field

T = TypeVar("T", bound="DevModeState")


@_attrs_define
class DevModeState:
    """`GET /device/{udid}/devmode` — developer mode state.

    Attributes:
        developer_mode_enabled (bool):
    """

    developer_mode_enabled: bool
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        developer_mode_enabled = self.developer_mode_enabled

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "DeveloperModeEnabled": developer_mode_enabled,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        developer_mode_enabled = d.pop("DeveloperModeEnabled")

        dev_mode_state = cls(
            developer_mode_enabled=developer_mode_enabled,
        )

        dev_mode_state.additional_properties = d
        return dev_mode_state

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
