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


T = TypeVar("T", bound="AttachDetachEvent")


@_attrs_define
class AttachDetachEvent:
    """A device was attached to or detached from the host.

    Attributes:
        event (str): Event kind.
            `attached` when a device connects, `detached` when it disconnects,
            `paired` when a pairing record appears.
        device_id (Union[Unset, int]): usbmuxd device id.
        udid (Union[Unset, str]): The device udid (serial number), when known.
        properties (Union[Unset, DeviceProperties]): Low-level device properties reported by usbmuxd / lockdown.
    """

    event: str
    device_id: Union[Unset, int] = UNSET
    udid: Union[Unset, str] = UNSET
    properties: Union[Unset, "DeviceProperties"] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        event = self.event

        device_id = self.device_id

        udid = self.udid

        properties: Union[Unset, dict[str, Any]] = UNSET
        if not isinstance(self.properties, Unset):
            properties = self.properties.to_dict()

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "event": event,
            }
        )
        if device_id is not UNSET:
            field_dict["deviceID"] = device_id
        if udid is not UNSET:
            field_dict["udid"] = udid
        if properties is not UNSET:
            field_dict["properties"] = properties

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.device_properties import DeviceProperties

        d = dict(src_dict)
        event = d.pop("event")

        device_id = d.pop("deviceID", UNSET)

        udid = d.pop("udid", UNSET)

        _properties = d.pop("properties", UNSET)
        properties: Union[Unset, DeviceProperties]
        if isinstance(_properties, Unset):
            properties = UNSET
        else:
            properties = DeviceProperties.from_dict(_properties)

        attach_detach_event = cls(
            event=event,
            device_id=device_id,
            udid=udid,
            properties=properties,
        )

        attach_detach_event.additional_properties = d
        return attach_detach_event

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
