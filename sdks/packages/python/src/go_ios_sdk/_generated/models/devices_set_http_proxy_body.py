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

T = TypeVar("T", bound="DevicesSetHttpProxyBody")


@_attrs_define
class DevicesSetHttpProxyBody:
    """
    Attributes:
        host (str): Proxy host.
        port (str): Proxy port.
        p12 (Any): Supervision identity certificate.
        user (Union[Unset, str]): Proxy username.
        pass_ (Union[Unset, str]): Proxy password.
        password (Union[Unset, str]): Passphrase for the `.p12` identity.
    """

    host: str
    port: str
    p12: Any
    user: Union[Unset, str] = UNSET
    pass_: Union[Unset, str] = UNSET
    password: Union[Unset, str] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        host = self.host

        port = self.port

        p12 = self.p12

        user = self.user

        pass_ = self.pass_

        password = self.password

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "host": host,
                "port": port,
                "p12": p12,
            }
        )
        if user is not UNSET:
            field_dict["user"] = user
        if pass_ is not UNSET:
            field_dict["pass"] = pass_
        if password is not UNSET:
            field_dict["password"] = password

        return field_dict

    def to_multipart(self) -> types.RequestFiles:
        files: types.RequestFiles = []

        files.append(("host", (None, str(self.host).encode(), "text/plain")))

        files.append(("port", (None, str(self.port).encode(), "text/plain")))

        files.append(("p12", (None, str(self.p12).encode(), "text/plain")))

        if not isinstance(self.user, Unset):
            files.append(("user", (None, str(self.user).encode(), "text/plain")))

        if not isinstance(self.pass_, Unset):
            files.append(("pass", (None, str(self.pass_).encode(), "text/plain")))

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
        host = d.pop("host")

        port = d.pop("port")

        p12 = d.pop("p12")

        user = d.pop("user", UNSET)

        pass_ = d.pop("pass", UNSET)

        password = d.pop("password", UNSET)

        devices_set_http_proxy_body = cls(
            host=host,
            port=port,
            p12=p12,
            user=user,
            pass_=pass_,
            password=password,
        )

        devices_set_http_proxy_body.additional_properties = d
        return devices_set_http_proxy_body

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
