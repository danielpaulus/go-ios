from collections.abc import Mapping
from typing import (
    TYPE_CHECKING,
    Any,
    TypeVar,
)

from attrs import define as _attrs_define
from attrs import field as _attrs_field

if TYPE_CHECKING:
    from ..models.device_entry import DeviceEntry


T = TypeVar("T", bound="DeviceList")


@_attrs_define
class DeviceList:
    """Response of `GET /list`.

    Attributes:
        device_list (list['DeviceEntry']):
    """

    device_list: list["DeviceEntry"]
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        device_list = []
        for device_list_item_data in self.device_list:
            device_list_item = device_list_item_data.to_dict()
            device_list.append(device_list_item)

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "deviceList": device_list,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.device_entry import DeviceEntry

        d = dict(src_dict)
        device_list = []
        _device_list = d.pop("deviceList")
        for device_list_item_data in _device_list:
            device_list_item = DeviceEntry.from_dict(device_list_item_data)

            device_list.append(device_list_item)

        device_list = cls(
            device_list=device_list,
        )

        device_list.additional_properties = d
        return device_list

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
