from collections.abc import Mapping
from typing import (
    Any,
    TypeVar,
    Union,
)

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

T = TypeVar("T", bound="ProvisioningResult")


@_attrs_define
class ProvisioningResult:
    """`POST /sign/provision` — provisioning assets envelope. The mobileprovision
    (and optionally the P12) are base64-encoded so one JSON response can carry
    both binary artifacts. Host-scoped (device-free).

        Attributes:
            bundle_id (str): The app bundle identifier registered with App Store Connect.
            certificate_id (str): The signing certificate resource id.
            mobileprovision_base_64 (str): The `.mobileprovision`, base64-encoded.
            p_12_base_64 (Union[Unset, str]): The generated `.p12`, base64-encoded (absent when reusing a certificate).
            p_12_password (Union[Unset, str]): The password protecting `p12Base64`, echoed back (client-supplied).
    """

    bundle_id: str
    certificate_id: str
    mobileprovision_base_64: str
    p_12_base_64: Union[Unset, str] = UNSET
    p_12_password: Union[Unset, str] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        bundle_id = self.bundle_id

        certificate_id = self.certificate_id

        mobileprovision_base_64 = self.mobileprovision_base_64

        p_12_base_64 = self.p_12_base_64

        p_12_password = self.p_12_password

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "bundleId": bundle_id,
                "certificateId": certificate_id,
                "mobileprovisionBase64": mobileprovision_base_64,
            }
        )
        if p_12_base_64 is not UNSET:
            field_dict["p12Base64"] = p_12_base_64
        if p_12_password is not UNSET:
            field_dict["p12Password"] = p_12_password

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        bundle_id = d.pop("bundleId")

        certificate_id = d.pop("certificateId")

        mobileprovision_base_64 = d.pop("mobileprovisionBase64")

        p_12_base_64 = d.pop("p12Base64", UNSET)

        p_12_password = d.pop("p12Password", UNSET)

        provisioning_result = cls(
            bundle_id=bundle_id,
            certificate_id=certificate_id,
            mobileprovision_base_64=mobileprovision_base_64,
            p_12_base_64=p_12_base_64,
            p_12_password=p_12_password,
        )

        provisioning_result.additional_properties = d
        return provisioning_result

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
