from collections.abc import Mapping
from typing import (
    Any,
    TypeVar,
    Union,
)

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

T = TypeVar("T", bound="BatteryRegistry")


@_attrs_define
class BatteryRegistry:
    """`GET /device/{udid}/battery/registry` — battery IORegistry stats
    (`diagnostics.IORegistry`). Open map; common keys surfaced.

        Attributes:
            temperature (Union[Unset, int]):
            voltage (Union[Unset, int]):
            current_capacity (Union[Unset, int]):
            instant_amperage (Union[Unset, int]):
            is_charging (Union[Unset, bool]):
            fully_charged (Union[Unset, bool]):
    """

    temperature: Union[Unset, int] = UNSET
    voltage: Union[Unset, int] = UNSET
    current_capacity: Union[Unset, int] = UNSET
    instant_amperage: Union[Unset, int] = UNSET
    is_charging: Union[Unset, bool] = UNSET
    fully_charged: Union[Unset, bool] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        temperature = self.temperature

        voltage = self.voltage

        current_capacity = self.current_capacity

        instant_amperage = self.instant_amperage

        is_charging = self.is_charging

        fully_charged = self.fully_charged

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update({})
        if temperature is not UNSET:
            field_dict["Temperature"] = temperature
        if voltage is not UNSET:
            field_dict["Voltage"] = voltage
        if current_capacity is not UNSET:
            field_dict["CurrentCapacity"] = current_capacity
        if instant_amperage is not UNSET:
            field_dict["InstantAmperage"] = instant_amperage
        if is_charging is not UNSET:
            field_dict["IsCharging"] = is_charging
        if fully_charged is not UNSET:
            field_dict["FullyCharged"] = fully_charged

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        temperature = d.pop("Temperature", UNSET)

        voltage = d.pop("Voltage", UNSET)

        current_capacity = d.pop("CurrentCapacity", UNSET)

        instant_amperage = d.pop("InstantAmperage", UNSET)

        is_charging = d.pop("IsCharging", UNSET)

        fully_charged = d.pop("FullyCharged", UNSET)

        battery_registry = cls(
            temperature=temperature,
            voltage=voltage,
            current_capacity=current_capacity,
            instant_amperage=instant_amperage,
            is_charging=is_charging,
            fully_charged=fully_charged,
        )

        battery_registry.additional_properties = d
        return battery_registry

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
