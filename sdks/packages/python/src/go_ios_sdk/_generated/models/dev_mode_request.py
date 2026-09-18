from collections.abc import Mapping
from typing import (
    Any,
    TypeVar,
    Union,
)

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

T = TypeVar("T", bound="DevModeRequest")


@_attrs_define
class DevModeRequest:
    """`POST /device/{udid}/devmode` request.

    Attributes:
        action (str): `enable` to turn developer mode on, `reveal` to expose the settings menu.
        enable_post_restart (Union[Unset, bool]): When enabling, also arm developer mode to persist across the next
            reboot.
    """

    action: str
    enable_post_restart: Union[Unset, bool] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        action = self.action

        enable_post_restart = self.enable_post_restart

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "action": action,
            }
        )
        if enable_post_restart is not UNSET:
            field_dict["enablePostRestart"] = enable_post_restart

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        action = d.pop("action")

        enable_post_restart = d.pop("enablePostRestart", UNSET)

        dev_mode_request = cls(
            action=action,
            enable_post_restart=enable_post_restart,
        )

        dev_mode_request.additional_properties = d
        return dev_mode_request

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
