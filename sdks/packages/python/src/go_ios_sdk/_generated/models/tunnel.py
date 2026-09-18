from collections.abc import Mapping
from typing import (
    Any,
    TypeVar,
    Union,
)

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

T = TypeVar("T", bound="Tunnel")


@_attrs_define
class Tunnel:
    """A running device tunnel as reported by the tunnel agent
    (`GET /tunnels`, `POST /tunnels/{udid}/refresh`). Mirrors `tunnel.Tunnel`.

        Attributes:
            udid (str): The device udid this tunnel serves.
            address (str): Tunnel address (IPv6) reachable for RemoteXPC/RSD.
            rsd_port (int): RemoteServiceDiscovery port on the tunnel.
            userspace_tun (Union[Unset, bool]): Whether this tunnel is a userspace TUN.
            userspace_tun_port (Union[Unset, int]): Userspace TUN port, when `UserspaceTUN` is true.
    """

    udid: str
    address: str
    rsd_port: int
    userspace_tun: Union[Unset, bool] = UNSET
    userspace_tun_port: Union[Unset, int] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        udid = self.udid

        address = self.address

        rsd_port = self.rsd_port

        userspace_tun = self.userspace_tun

        userspace_tun_port = self.userspace_tun_port

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "Udid": udid,
                "Address": address,
                "RsdPort": rsd_port,
            }
        )
        if userspace_tun is not UNSET:
            field_dict["UserspaceTUN"] = userspace_tun
        if userspace_tun_port is not UNSET:
            field_dict["UserspaceTUNPort"] = userspace_tun_port

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        udid = d.pop("Udid")

        address = d.pop("Address")

        rsd_port = d.pop("RsdPort")

        userspace_tun = d.pop("UserspaceTUN", UNSET)

        userspace_tun_port = d.pop("UserspaceTUNPort", UNSET)

        tunnel = cls(
            udid=udid,
            address=address,
            rsd_port=rsd_port,
            userspace_tun=userspace_tun,
            userspace_tun_port=userspace_tun_port,
        )

        tunnel.additional_properties = d
        return tunnel

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
