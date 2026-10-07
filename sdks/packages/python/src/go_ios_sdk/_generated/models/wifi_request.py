from collections.abc import Mapping
from typing import (
    Any,
    TypeVar,
    Union,
)

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

T = TypeVar("T", bound="WifiRequest")


@_attrs_define
class WifiRequest:
    """`PUT /device/{udid}/wifi` request.

    Attributes:
        ssid (str):
        password (Union[Unset, str]):
        enc_type (Union[Unset, str]): Encryption type, e.g. `WPA2`, `WPA`, `WEP`, `None`.
    """

    ssid: str
    password: Union[Unset, str] = UNSET
    enc_type: Union[Unset, str] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        ssid = self.ssid

        password = self.password

        enc_type = self.enc_type

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "ssid": ssid,
            }
        )
        if password is not UNSET:
            field_dict["password"] = password
        if enc_type is not UNSET:
            field_dict["encType"] = enc_type

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        ssid = d.pop("ssid")

        password = d.pop("password", UNSET)

        enc_type = d.pop("encType", UNSET)

        wifi_request = cls(
            ssid=ssid,
            password=password,
            enc_type=enc_type,
        )

        wifi_request.additional_properties = d
        return wifi_request

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
