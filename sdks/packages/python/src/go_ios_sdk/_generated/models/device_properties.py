from collections.abc import Mapping
from typing import (
    Any,
    TypeVar,
    Union,
)

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

T = TypeVar("T", bound="DeviceProperties")


@_attrs_define
class DeviceProperties:
    """Low-level device properties reported by usbmuxd / lockdown.

    Attributes:
        serial_number (str): The device udid (serial number). This is what device-scoped routes key on.
        connection_speed (Union[Unset, int]):
        connection_type (Union[Unset, str]):
        device_id (Union[Unset, int]):
        location_id (Union[Unset, int]):
        product_id (Union[Unset, int]):
    """

    serial_number: str
    connection_speed: Union[Unset, int] = UNSET
    connection_type: Union[Unset, str] = UNSET
    device_id: Union[Unset, int] = UNSET
    location_id: Union[Unset, int] = UNSET
    product_id: Union[Unset, int] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        serial_number = self.serial_number

        connection_speed = self.connection_speed

        connection_type = self.connection_type

        device_id = self.device_id

        location_id = self.location_id

        product_id = self.product_id

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "serialNumber": serial_number,
            }
        )
        if connection_speed is not UNSET:
            field_dict["connectionSpeed"] = connection_speed
        if connection_type is not UNSET:
            field_dict["connectionType"] = connection_type
        if device_id is not UNSET:
            field_dict["deviceID"] = device_id
        if location_id is not UNSET:
            field_dict["locationID"] = location_id
        if product_id is not UNSET:
            field_dict["productID"] = product_id

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        serial_number = d.pop("serialNumber")

        connection_speed = d.pop("connectionSpeed", UNSET)

        connection_type = d.pop("connectionType", UNSET)

        device_id = d.pop("deviceID", UNSET)

        location_id = d.pop("locationID", UNSET)

        product_id = d.pop("productID", UNSET)

        device_properties = cls(
            serial_number=serial_number,
            connection_speed=connection_speed,
            connection_type=connection_type,
            device_id=device_id,
            location_id=location_id,
            product_id=product_id,
        )

        device_properties.additional_properties = d
        return device_properties

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
