from collections.abc import Mapping
from typing import (
    TYPE_CHECKING,
    Any,
    TypeVar,
)

from attrs import define as _attrs_define
from attrs import field as _attrs_field

if TYPE_CHECKING:
    from ..models.fsync_tree_entry import FsyncTreeEntry


T = TypeVar("T", bound="FsyncTreeListing")


@_attrs_define
class FsyncTreeListing:
    """`GET /device/{udid}/fsync/tree` — a recursive directory walk over AFC.

    Attributes:
        path (str): The root (cleaned) device path.
        entries (list['FsyncTreeEntry']): Flattened list of entries in the subtree.
        count (int): Number of entries.
    """

    path: str
    entries: list["FsyncTreeEntry"]
    count: int
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        path = self.path

        entries = []
        for entries_item_data in self.entries:
            entries_item = entries_item_data.to_dict()
            entries.append(entries_item)

        count = self.count

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "path": path,
                "entries": entries,
                "count": count,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.fsync_tree_entry import FsyncTreeEntry

        d = dict(src_dict)
        path = d.pop("path")

        entries = []
        _entries = d.pop("entries")
        for entries_item_data in _entries:
            entries_item = FsyncTreeEntry.from_dict(entries_item_data)

            entries.append(entries_item)

        count = d.pop("count")

        fsync_tree_listing = cls(
            path=path,
            entries=entries,
            count=count,
        )

        fsync_tree_listing.additional_properties = d
        return fsync_tree_listing

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
