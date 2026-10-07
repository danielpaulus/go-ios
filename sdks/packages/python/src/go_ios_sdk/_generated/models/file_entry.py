from collections.abc import Mapping
from typing import (
    Any,
    TypeVar,
    Union,
)

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

T = TypeVar("T", bound="FileEntry")


@_attrs_define
class FileEntry:
    """A single entry in a device directory listing.

    Attributes:
        name (Union[Unset, str]):
        path (Union[Unset, str]):
        is_dir (Union[Unset, bool]):
        size (Union[Unset, int]):
    """

    name: Union[Unset, str] = UNSET
    path: Union[Unset, str] = UNSET
    is_dir: Union[Unset, bool] = UNSET
    size: Union[Unset, int] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        name = self.name

        path = self.path

        is_dir = self.is_dir

        size = self.size

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update({})
        if name is not UNSET:
            field_dict["name"] = name
        if path is not UNSET:
            field_dict["path"] = path
        if is_dir is not UNSET:
            field_dict["isDir"] = is_dir
        if size is not UNSET:
            field_dict["size"] = size

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        name = d.pop("name", UNSET)

        path = d.pop("path", UNSET)

        is_dir = d.pop("isDir", UNSET)

        size = d.pop("size", UNSET)

        file_entry = cls(
            name=name,
            path=path,
            is_dir=is_dir,
            size=size,
        )

        file_entry.additional_properties = d
        return file_entry

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
