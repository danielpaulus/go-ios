from collections.abc import Mapping
from typing import (
    Any,
    TypeVar,
    Union,
)

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

T = TypeVar("T", bound="DiskSpaceInfo")


@_attrs_define
class DiskSpaceInfo:
    """`GET /device/{udid}/diskspace` — AFC filesystem info (`afc.DeviceInfo`).
    Total/free/used bytes and block size. Open map; common keys surfaced.

        Attributes:
            fs_total_bytes (Union[Unset, int]): Total filesystem capacity in bytes.
            fs_free_bytes (Union[Unset, int]): Free filesystem space in bytes.
            fs_block_size (Union[Unset, int]): Filesystem block size in bytes.
            model (Union[Unset, str]): AFC model identifier reported by the device.
    """

    fs_total_bytes: Union[Unset, int] = UNSET
    fs_free_bytes: Union[Unset, int] = UNSET
    fs_block_size: Union[Unset, int] = UNSET
    model: Union[Unset, str] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        fs_total_bytes = self.fs_total_bytes

        fs_free_bytes = self.fs_free_bytes

        fs_block_size = self.fs_block_size

        model = self.model

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update({})
        if fs_total_bytes is not UNSET:
            field_dict["FSTotalBytes"] = fs_total_bytes
        if fs_free_bytes is not UNSET:
            field_dict["FSFreeBytes"] = fs_free_bytes
        if fs_block_size is not UNSET:
            field_dict["FSBlockSize"] = fs_block_size
        if model is not UNSET:
            field_dict["Model"] = model

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        fs_total_bytes = d.pop("FSTotalBytes", UNSET)

        fs_free_bytes = d.pop("FSFreeBytes", UNSET)

        fs_block_size = d.pop("FSBlockSize", UNSET)

        model = d.pop("Model", UNSET)

        disk_space_info = cls(
            fs_total_bytes=fs_total_bytes,
            fs_free_bytes=fs_free_bytes,
            fs_block_size=fs_block_size,
            model=model,
        )

        disk_space_info.additional_properties = d
        return disk_space_info

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
