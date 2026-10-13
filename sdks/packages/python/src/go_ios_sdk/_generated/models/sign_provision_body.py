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

T = TypeVar("T", bound="SignProvisionBody")


@_attrs_define
class SignProvisionBody:
    """
    Attributes:
        asc_private_key (Any): `.p8` App Store Connect API key.
        asc_key_id (str): App Store Connect key id.
        asc_issuer_id (str): App Store Connect issuer id.
        bundleid (str): App bundle identifier.
        udid (str): Target device udid to register against the profile.
        bundlename (Union[Unset, str]): Bundle display name.
        profilename (Union[Unset, str]): Provisioning profile name.
        devicename (Union[Unset, str]): Device display name.
        certificate_id (Union[Unset, str]): Reuse an existing certificate (no new P12 is generated).
        revoke_existing (Union[Unset, str]): Revoke existing certificates first.
        p12password (Union[Unset, str]): Password to protect the generated P12.
    """

    asc_private_key: Any
    asc_key_id: str
    asc_issuer_id: str
    bundleid: str
    udid: str
    bundlename: Union[Unset, str] = UNSET
    profilename: Union[Unset, str] = UNSET
    devicename: Union[Unset, str] = UNSET
    certificate_id: Union[Unset, str] = UNSET
    revoke_existing: Union[Unset, str] = UNSET
    p12password: Union[Unset, str] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        asc_private_key = self.asc_private_key

        asc_key_id = self.asc_key_id

        asc_issuer_id = self.asc_issuer_id

        bundleid = self.bundleid

        udid = self.udid

        bundlename = self.bundlename

        profilename = self.profilename

        devicename = self.devicename

        certificate_id = self.certificate_id

        revoke_existing = self.revoke_existing

        p12password = self.p12password

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "asc-private-key": asc_private_key,
                "asc-key-id": asc_key_id,
                "asc-issuer-id": asc_issuer_id,
                "bundleid": bundleid,
                "udid": udid,
            }
        )
        if bundlename is not UNSET:
            field_dict["bundlename"] = bundlename
        if profilename is not UNSET:
            field_dict["profilename"] = profilename
        if devicename is not UNSET:
            field_dict["devicename"] = devicename
        if certificate_id is not UNSET:
            field_dict["certificate-id"] = certificate_id
        if revoke_existing is not UNSET:
            field_dict["revoke-existing"] = revoke_existing
        if p12password is not UNSET:
            field_dict["p12password"] = p12password

        return field_dict

    def to_multipart(self) -> types.RequestFiles:
        files: types.RequestFiles = []

        files.append(
            (
                "asc-private-key",
                (None, str(self.asc_private_key).encode(), "text/plain"),
            )
        )

        files.append(
            ("asc-key-id", (None, str(self.asc_key_id).encode(), "text/plain"))
        )

        files.append(
            ("asc-issuer-id", (None, str(self.asc_issuer_id).encode(), "text/plain"))
        )

        files.append(("bundleid", (None, str(self.bundleid).encode(), "text/plain")))

        files.append(("udid", (None, str(self.udid).encode(), "text/plain")))

        if not isinstance(self.bundlename, Unset):
            files.append(
                ("bundlename", (None, str(self.bundlename).encode(), "text/plain"))
            )

        if not isinstance(self.profilename, Unset):
            files.append(
                ("profilename", (None, str(self.profilename).encode(), "text/plain"))
            )

        if not isinstance(self.devicename, Unset):
            files.append(
                ("devicename", (None, str(self.devicename).encode(), "text/plain"))
            )

        if not isinstance(self.certificate_id, Unset):
            files.append(
                (
                    "certificate-id",
                    (None, str(self.certificate_id).encode(), "text/plain"),
                )
            )

        if not isinstance(self.revoke_existing, Unset):
            files.append(
                (
                    "revoke-existing",
                    (None, str(self.revoke_existing).encode(), "text/plain"),
                )
            )

        if not isinstance(self.p12password, Unset):
            files.append(
                ("p12password", (None, str(self.p12password).encode(), "text/plain"))
            )

        for prop_name, prop in self.additional_properties.items():
            files.append((prop_name, (None, str(prop).encode(), "text/plain")))

        return files

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        asc_private_key = d.pop("asc-private-key")

        asc_key_id = d.pop("asc-key-id")

        asc_issuer_id = d.pop("asc-issuer-id")

        bundleid = d.pop("bundleid")

        udid = d.pop("udid")

        bundlename = d.pop("bundlename", UNSET)

        profilename = d.pop("profilename", UNSET)

        devicename = d.pop("devicename", UNSET)

        certificate_id = d.pop("certificate-id", UNSET)

        revoke_existing = d.pop("revoke-existing", UNSET)

        p12password = d.pop("p12password", UNSET)

        sign_provision_body = cls(
            asc_private_key=asc_private_key,
            asc_key_id=asc_key_id,
            asc_issuer_id=asc_issuer_id,
            bundleid=bundleid,
            udid=udid,
            bundlename=bundlename,
            profilename=profilename,
            devicename=devicename,
            certificate_id=certificate_id,
            revoke_existing=revoke_existing,
            p12password=p12password,
        )

        sign_provision_body.additional_properties = d
        return sign_provision_body

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
