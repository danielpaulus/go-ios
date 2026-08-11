from collections.abc import Mapping
from typing import (
    Any,
    TypeVar,
    Union,
)

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

T = TypeVar("T", bound="UISwipeRequest")


@_attrs_define
class UISwipeRequest:
    """`POST /device/{udid}/ui/swipe` request — drag from (x1,y1) to (x2,y2).

    Attributes:
        x1 (int):
        y1 (int):
        x2 (int):
        y2 (int):
        duration (Union[Unset, float]): Gesture duration in seconds.
    """

    x1: int
    y1: int
    x2: int
    y2: int
    duration: Union[Unset, float] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        x1 = self.x1

        y1 = self.y1

        x2 = self.x2

        y2 = self.y2

        duration = self.duration

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "x1": x1,
                "y1": y1,
                "x2": x2,
                "y2": y2,
            }
        )
        if duration is not UNSET:
            field_dict["duration"] = duration

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        x1 = d.pop("x1")

        y1 = d.pop("y1")

        x2 = d.pop("x2")

        y2 = d.pop("y2")

        duration = d.pop("duration", UNSET)

        ui_swipe_request = cls(
            x1=x1,
            y1=y1,
            x2=x2,
            y2=y2,
            duration=duration,
        )

        ui_swipe_request.additional_properties = d
        return ui_swipe_request

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
