from collections.abc import Mapping
from typing import Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field

T = TypeVar("T", bound="PrepareResult")


@_attrs_define
class PrepareResult:
    """`POST /device/{udid}/prepare` — device preparation acknowledgement.

    Attributes:
        status (str): Always `prepared`.
        supervised (bool): Whether the device was supervised (a supervision cert was supplied).
    """

    status: str
    supervised: bool
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        status = self.status

        supervised = self.supervised

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "status": status,
                "supervised": supervised,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        status = d.pop("status")

        supervised = d.pop("supervised")

        prepare_result = cls(
            status=status,
            supervised=supervised,
        )

        prepare_result.additional_properties = d
        return prepare_result

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
