from collections.abc import Mapping
from typing import (
    Any,
    TypeVar,
    cast,
)

from attrs import define as _attrs_define
from attrs import field as _attrs_field

T = TypeVar("T", bound="PrepareSkipOptions")


@_attrs_define
class PrepareSkipOptions:
    """`GET /prepare/skip-options` — the static list of setup-pane skip options
    usable when preparing a device. Host-scoped (device-free).

        Attributes:
            options (list[str]): All available skip-option identifiers.
            count (int): Number of options.
    """

    options: list[str]
    count: int
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        options = self.options

        count = self.count

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "options": options,
                "count": count,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        options = cast(list[str], d.pop("options"))

        count = d.pop("count")

        prepare_skip_options = cls(
            options=options,
            count=count,
        )

        prepare_skip_options.additional_properties = d
        return prepare_skip_options

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
