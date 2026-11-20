from collections.abc import Mapping
from typing import Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field

T = TypeVar("T", bound="MemLimitResult")


@_attrs_define
class MemLimitResult:
    """`POST /device/{udid}/memlimitoff` response.

    Attributes:
        process (str):
        pid (int):
        disabled (bool):
    """

    process: str
    pid: int
    disabled: bool
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        process = self.process

        pid = self.pid

        disabled = self.disabled

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "process": process,
                "pid": pid,
                "disabled": disabled,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        process = d.pop("process")

        pid = d.pop("pid")

        disabled = d.pop("disabled")

        mem_limit_result = cls(
            process=process,
            pid=pid,
            disabled=disabled,
        )

        mem_limit_result.additional_properties = d
        return mem_limit_result

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
