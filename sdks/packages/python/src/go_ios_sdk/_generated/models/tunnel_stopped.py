from collections.abc import Mapping
from typing import Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field

T = TypeVar("T", bound="TunnelStopped")


@_attrs_define
class TunnelStopped:
    """`DELETE /tunnels/{udid}` — acknowledgement that the tunnel was stopped.

    Attributes:
        udid (str):
        status (str): Always `stopped`.
    """

    udid: str
    status: str
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        udid = self.udid

        status = self.status

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "udid": udid,
                "status": status,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        udid = d.pop("udid")

        status = d.pop("status")

        tunnel_stopped = cls(
            udid=udid,
            status=status,
        )

        tunnel_stopped.additional_properties = d
        return tunnel_stopped

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
