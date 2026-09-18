from collections.abc import Mapping
from typing import (
    Any,
    TypeVar,
    Union,
)

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

T = TypeVar("T", bound="CpuUsageSample")


@_attrs_define
class CpuUsageSample:
    """A single sysmontap CPU-usage sample. Open map; sampler keys vary by OS.

    Attributes:
        cpu_total_load (Union[Unset, float]): Total CPU load across all cores (0–100).
        system_load (Union[Unset, float]): System (kernel) CPU load.
        user_load (Union[Unset, float]): User CPU load.
    """

    cpu_total_load: Union[Unset, float] = UNSET
    system_load: Union[Unset, float] = UNSET
    user_load: Union[Unset, float] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        cpu_total_load = self.cpu_total_load

        system_load = self.system_load

        user_load = self.user_load

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update({})
        if cpu_total_load is not UNSET:
            field_dict["CPU_TotalLoad"] = cpu_total_load
        if system_load is not UNSET:
            field_dict["SystemLoad"] = system_load
        if user_load is not UNSET:
            field_dict["UserLoad"] = user_load

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        cpu_total_load = d.pop("CPU_TotalLoad", UNSET)

        system_load = d.pop("SystemLoad", UNSET)

        user_load = d.pop("UserLoad", UNSET)

        cpu_usage_sample = cls(
            cpu_total_load=cpu_total_load,
            system_load=system_load,
            user_load=user_load,
        )

        cpu_usage_sample.additional_properties = d
        return cpu_usage_sample

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
