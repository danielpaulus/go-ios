from collections.abc import Mapping
from typing import (
    TYPE_CHECKING,
    Any,
    TypeVar,
    Union,
)

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

if TYPE_CHECKING:
    from ..models.device_properties import DeviceProperties


T = TypeVar("T", bound="DeviceEntry")


@_attrs_define
class DeviceEntry:
    """A single device as returned by `GET /list`.

    Attributes:
        device_id (int):
        properties (DeviceProperties): Low-level device properties reported by usbmuxd / lockdown.
        message_type (Union[Unset, str]):
        address (Union[Unset, str]): Network address for a device reached over the network / tunnel.
        userspace_tun (Union[Unset, bool]): True if reachable via the userspace TUN tunnel.
        userspace_tun_host (Union[Unset, str]):
        userspace_tun_port (Union[Unset, int]):
    """

    device_id: int
    properties: "DeviceProperties"
    message_type: Union[Unset, str] = UNSET
    address: Union[Unset, str] = UNSET
    userspace_tun: Union[Unset, bool] = UNSET
    userspace_tun_host: Union[Unset, str] = UNSET
    userspace_tun_port: Union[Unset, int] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        device_id = self.device_id

        properties = self.properties.to_dict()

        message_type = self.message_type

        address = self.address

        userspace_tun = self.userspace_tun

        userspace_tun_host = self.userspace_tun_host

        userspace_tun_port = self.userspace_tun_port

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "deviceID": device_id,
                "properties": properties,
            }
        )
        if message_type is not UNSET:
            field_dict["messageType"] = message_type
        if address is not UNSET:
            field_dict["address"] = address
        if userspace_tun is not UNSET:
            field_dict["userspaceTUN"] = userspace_tun
        if userspace_tun_host is not UNSET:
            field_dict["userspaceTUNHost"] = userspace_tun_host
        if userspace_tun_port is not UNSET:
            field_dict["userspaceTUNPort"] = userspace_tun_port

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.device_properties import DeviceProperties

        d = dict(src_dict)
        device_id = d.pop("deviceID")

        properties = DeviceProperties.from_dict(d.pop("properties"))

        message_type = d.pop("messageType", UNSET)

        address = d.pop("address", UNSET)

        userspace_tun = d.pop("userspaceTUN", UNSET)

        userspace_tun_host = d.pop("userspaceTUNHost", UNSET)

        userspace_tun_port = d.pop("userspaceTUNPort", UNSET)

        device_entry = cls(
            device_id=device_id,
            properties=properties,
            message_type=message_type,
            address=address,
            userspace_tun=userspace_tun,
            userspace_tun_host=userspace_tun_host,
            userspace_tun_port=userspace_tun_port,
        )

        device_entry.additional_properties = d
        return device_entry

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
