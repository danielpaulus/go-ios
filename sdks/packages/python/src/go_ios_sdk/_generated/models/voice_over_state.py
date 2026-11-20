from collections.abc import Mapping
from typing import Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field

T = TypeVar("T", bound="VoiceOverState")


@_attrs_define
class VoiceOverState:
    """`GET|PUT /device/{udid}/voiceover` — VoiceOver enabled state.

    Attributes:
        voice_over_enabled (bool):
    """

    voice_over_enabled: bool
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        voice_over_enabled = self.voice_over_enabled

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "VoiceOverEnabled": voice_over_enabled,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        voice_over_enabled = d.pop("VoiceOverEnabled")

        voice_over_state = cls(
            voice_over_enabled=voice_over_enabled,
        )

        voice_over_state.additional_properties = d
        return voice_over_state

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
