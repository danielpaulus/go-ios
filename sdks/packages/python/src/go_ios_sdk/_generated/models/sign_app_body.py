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

T = TypeVar("T", bound="SignAppBody")


@_attrs_define
class SignAppBody:
    """
    Attributes:
        ipa (Any): App or `.ipa` to resign.
        p12file (Any): Signing identity (`.p12`).
        profile (Any): Provisioning profile (`.mobileprovision`).
        p12password (Union[Unset, str]): P12 password.
        bundleid (Union[Unset, str]): Override bundle id.
    """

    ipa: Any
    p12file: Any
    profile: Any
    p12password: Union[Unset, str] = UNSET
    bundleid: Union[Unset, str] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        ipa = self.ipa

        p12file = self.p12file

        profile = self.profile

        p12password = self.p12password

        bundleid = self.bundleid

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "ipa": ipa,
                "p12file": p12file,
                "profile": profile,
            }
        )
        if p12password is not UNSET:
            field_dict["p12password"] = p12password
        if bundleid is not UNSET:
            field_dict["bundleid"] = bundleid

        return field_dict

    def to_multipart(self) -> types.RequestFiles:
        files: types.RequestFiles = []

        files.append(("ipa", (None, str(self.ipa).encode(), "text/plain")))

        files.append(("p12file", (None, str(self.p12file).encode(), "text/plain")))

        files.append(("profile", (None, str(self.profile).encode(), "text/plain")))

        if not isinstance(self.p12password, Unset):
            files.append(
                ("p12password", (None, str(self.p12password).encode(), "text/plain"))
            )

        if not isinstance(self.bundleid, Unset):
            files.append(
                ("bundleid", (None, str(self.bundleid).encode(), "text/plain"))
            )

        for prop_name, prop in self.additional_properties.items():
            files.append((prop_name, (None, str(prop).encode(), "text/plain")))

        return files

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        ipa = d.pop("ipa")

        p12file = d.pop("p12file")

        profile = d.pop("profile")

        p12password = d.pop("p12password", UNSET)

        bundleid = d.pop("bundleid", UNSET)

        sign_app_body = cls(
            ipa=ipa,
            p12file=p12file,
            profile=profile,
            p12password=p12password,
            bundleid=bundleid,
        )

        sign_app_body.additional_properties = d
        return sign_app_body

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
