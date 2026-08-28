from collections.abc import Mapping
from typing import (
    Any,
    TypeVar,
    Union,
)

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

T = TypeVar("T", bound="NetworkInfo")


@_attrs_define
class NetworkInfo:
    """`GET /device/{udid}/ip` — device network info discovered over pcapd
    (`pcap.NetworkInfo`).

        Attributes:
            mac_address (Union[Unset, str]): Hardware (MAC) address.
            i_pv_4 (Union[Unset, str]): IPv4 address, when discovered.
            i_pv_6 (Union[Unset, str]): IPv6 address, when discovered.
    """

    mac_address: Union[Unset, str] = UNSET
    i_pv_4: Union[Unset, str] = UNSET
    i_pv_6: Union[Unset, str] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        mac_address = self.mac_address

        i_pv_4 = self.i_pv_4

        i_pv_6 = self.i_pv_6

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update({})
        if mac_address is not UNSET:
            field_dict["MacAddress"] = mac_address
        if i_pv_4 is not UNSET:
            field_dict["IPv4"] = i_pv_4
        if i_pv_6 is not UNSET:
            field_dict["IPv6"] = i_pv_6

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        mac_address = d.pop("MacAddress", UNSET)

        i_pv_4 = d.pop("IPv4", UNSET)

        i_pv_6 = d.pop("IPv6", UNSET)

        network_info = cls(
            mac_address=mac_address,
            i_pv_4=i_pv_4,
            i_pv_6=i_pv_6,
        )

        network_info.additional_properties = d
        return network_info

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
