from collections.abc import Mapping
from typing import (
    Any,
    TypeVar,
    Union,
)

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

T = TypeVar("T", bound="BatteryInfo")


@_attrs_define
class BatteryInfo:
    """`GET /device/{udid}/battery` — battery diagnostics (`ios.BatteryInfo`).
    Open map; commonly-present keys are surfaced for discoverability.

        Attributes:
            current_capacity (Union[Unset, int]):
            external_connected (Union[Unset, bool]):
            fully_charged (Union[Unset, bool]):
            is_charging (Union[Unset, bool]):
            temperature (Union[Unset, int]):
    """

    current_capacity: Union[Unset, int] = UNSET
    external_connected: Union[Unset, bool] = UNSET
    fully_charged: Union[Unset, bool] = UNSET
    is_charging: Union[Unset, bool] = UNSET
    temperature: Union[Unset, int] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        current_capacity = self.current_capacity

        external_connected = self.external_connected

        fully_charged = self.fully_charged

        is_charging = self.is_charging

        temperature = self.temperature

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update({})
        if current_capacity is not UNSET:
            field_dict["CurrentCapacity"] = current_capacity
        if external_connected is not UNSET:
            field_dict["ExternalConnected"] = external_connected
        if fully_charged is not UNSET:
            field_dict["FullyCharged"] = fully_charged
        if is_charging is not UNSET:
            field_dict["IsCharging"] = is_charging
        if temperature is not UNSET:
            field_dict["Temperature"] = temperature

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        current_capacity = d.pop("CurrentCapacity", UNSET)

        external_connected = d.pop("ExternalConnected", UNSET)

        fully_charged = d.pop("FullyCharged", UNSET)

        is_charging = d.pop("IsCharging", UNSET)

        temperature = d.pop("Temperature", UNSET)

        battery_info = cls(
            current_capacity=current_capacity,
            external_connected=external_connected,
            fully_charged=fully_charged,
            is_charging=is_charging,
            temperature=temperature,
        )

        battery_info.additional_properties = d
        return battery_info

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
