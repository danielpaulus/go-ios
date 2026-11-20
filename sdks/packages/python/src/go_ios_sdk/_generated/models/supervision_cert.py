from collections.abc import Mapping
from typing import Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field

T = TypeVar("T", bound="SupervisionCert")


@_attrs_define
class SupervisionCert:
    """`POST /prepare/create-cert` — a generated self-signed supervision identity,
    returned as DER (base64) and PEM for both the certificate and private key.
    Host-scoped (device-free).

        Attributes:
            cert_der_base_64 (str): Certificate in DER form, base64-encoded.
            cert_pem (str): Certificate in PEM form.
            private_key_der_base_64 (str): Private key in DER form, base64-encoded.
            private_key_pem (str): Private key in PEM form.
    """

    cert_der_base_64: str
    cert_pem: str
    private_key_der_base_64: str
    private_key_pem: str
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        cert_der_base_64 = self.cert_der_base_64

        cert_pem = self.cert_pem

        private_key_der_base_64 = self.private_key_der_base_64

        private_key_pem = self.private_key_pem

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "certDerBase64": cert_der_base_64,
                "certPem": cert_pem,
                "privateKeyDerBase64": private_key_der_base_64,
                "privateKeyPem": private_key_pem,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        cert_der_base_64 = d.pop("certDerBase64")

        cert_pem = d.pop("certPem")

        private_key_der_base_64 = d.pop("privateKeyDerBase64")

        private_key_pem = d.pop("privateKeyPem")

        supervision_cert = cls(
            cert_der_base_64=cert_der_base_64,
            cert_pem=cert_pem,
            private_key_der_base_64=private_key_der_base_64,
            private_key_pem=private_key_pem,
        )

        supervision_cert.additional_properties = d
        return supervision_cert

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
