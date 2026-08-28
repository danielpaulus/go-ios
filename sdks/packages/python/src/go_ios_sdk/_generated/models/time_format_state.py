from collections.abc import Mapping
from typing import Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field

T = TypeVar("T", bound="TimeFormatState")


@_attrs_define
class TimeFormatState:
    """`GET /device/{udid}/timeformat` — 24-hour clock state.

    Attributes:
        uses_24_hour_clock (bool):
    """

    uses_24_hour_clock: bool
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        uses_24_hour_clock = self.uses_24_hour_clock

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "Uses24HourClock": uses_24_hour_clock,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        uses_24_hour_clock = d.pop("Uses24HourClock")

        time_format_state = cls(
            uses_24_hour_clock=uses_24_hour_clock,
        )

        time_format_state.additional_properties = d
        return time_format_state

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
