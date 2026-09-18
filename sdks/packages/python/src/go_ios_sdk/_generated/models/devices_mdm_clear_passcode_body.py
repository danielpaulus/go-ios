from collections.abc import Mapping
from typing import (
    Any,
    TypeVar,
    Union,
)

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from .. import types
from ..types import UNSET, Unset

T = TypeVar("T", bound="DevicesMdmClearPasscodeBody")


@_attrs_define
class DevicesMdmClearPasscodeBody:
    """
    Attributes:
        p12 (Any): The supervision identity certificate.
        token (str): Base64-encoded escrow unlock token.
        password (Union[Unset, str]): Passphrase for the `.p12` identity.
    """

    p12: Any
    token: str
    password: Union[Unset, str] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        p12 = self.p12

        token = self.token

        password = self.password

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "p12": p12,
                "token": token,
            }
        )
        if password is not UNSET:
            field_dict["password"] = password

        return field_dict

    def to_multipart(self) -> types.RequestFiles:
        files: types.RequestFiles = []

        files.append(("p12", (None, str(self.p12).encode(), "text/plain")))

        files.append(("token", (None, str(self.token).encode(), "text/plain")))

        if not isinstance(self.password, Unset):
            files.append(
                ("password", (None, str(self.password).encode(), "text/plain"))
            )

        for prop_name, prop in self.additional_properties.items():
            files.append((prop_name, (None, str(prop).encode(), "text/plain")))

        return files

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        p12 = d.pop("p12")

        token = d.pop("token")

        password = d.pop("password", UNSET)

        devices_mdm_clear_passcode_body = cls(
            p12=p12,
            token=token,
            password=password,
        )

        devices_mdm_clear_passcode_body.additional_properties = d
        return devices_mdm_clear_passcode_body

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
