from collections.abc import Mapping
from typing import Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field

T = TypeVar("T", bound="FsyncTreeEntry")


@_attrs_define
class FsyncTreeEntry:
    """One entry returned by the recursive `GET /device/{udid}/fsync/tree` walk.

    Attributes:
        path (str): Full device-side path of this entry.
        name (str): Base name of the entry.
        is_dir (bool): Whether the entry is a directory.
        size (int): Size in bytes.
    """

    path: str
    name: str
    is_dir: bool
    size: int
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        path = self.path

        name = self.name

        is_dir = self.is_dir

        size = self.size

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "path": path,
                "name": name,
                "isDir": is_dir,
                "size": size,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        path = d.pop("path")

        name = d.pop("name")

        is_dir = d.pop("isDir")

        size = d.pop("size")

        fsync_tree_entry = cls(
            path=path,
            name=name,
            is_dir=is_dir,
            size=size,
        )

        fsync_tree_entry.additional_properties = d
        return fsync_tree_entry

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
