from collections.abc import Mapping
from typing import (
    Any,
    TypeVar,
    Union,
)

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

T = TypeVar("T", bound="RsdServiceEntry")


@_attrs_define
class RsdServiceEntry:
    """A single RSD (Remote Service Discovery) service entry.

    Attributes:
        port (Union[Unset, int]): TCP port the service is reachable on over the tunnel.
        protocol_type (Union[Unset, str]): Wire protocol (e.g. `tcp`).
    """

    port: Union[Unset, int] = UNSET
    protocol_type: Union[Unset, str] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        port = self.port

        protocol_type = self.protocol_type

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update({})
        if port is not UNSET:
            field_dict["Port"] = port
        if protocol_type is not UNSET:
            field_dict["ProtocolType"] = protocol_type

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        port = d.pop("Port", UNSET)

        protocol_type = d.pop("ProtocolType", UNSET)

        rsd_service_entry = cls(
            port=port,
            protocol_type=protocol_type,
        )

        rsd_service_entry.additional_properties = d
        return rsd_service_entry

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
