from collections.abc import Mapping
from typing import Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field

T = TypeVar("T", bound="DeviceDate")


@_attrs_define
class DeviceDate:
    """`GET /device/{udid}/date`.

    Attributes:
        formated_date (str): Human-readable RFC850 date on the device.
        time_interval_since_1970 (float): Device clock as Unix epoch seconds.
    """

    formated_date: str
    time_interval_since_1970: float
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        formated_date = self.formated_date

        time_interval_since_1970 = self.time_interval_since_1970

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "formatedDate": formated_date,
                "TimeIntervalSince1970": time_interval_since_1970,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        formated_date = d.pop("formatedDate")

        time_interval_since_1970 = d.pop("TimeIntervalSince1970")

        device_date = cls(
            formated_date=formated_date,
            time_interval_since_1970=time_interval_since_1970,
        )

        device_date.additional_properties = d
        return device_date

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
