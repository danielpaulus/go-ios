from collections.abc import Mapping
from typing import (
    Any,
    TypeVar,
    Union,
    cast,
)

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from .. import types
from ..types import UNSET, Unset

T = TypeVar("T", bound="PreparePrepareDeviceBody")


@_attrs_define
class PreparePrepareDeviceBody:
    """
    Attributes:
        cert (Union[Unset, Any]): Supervision identity (DER/PEM/P12). Omit to prepare unsupervised.
        p12password (Union[Unset, str]): P12 password (when `cert` is a P12).
        skip (Union[Unset, list[str]]): Setup panes to skip (see /prepare/skip-options). Repeatable.
        orgname (Union[Unset, str]): Supervision organization name.
        locale (Union[Unset, str]): Device locale (default en_US).
        lang (Union[Unset, str]): Device language (default en).
    """

    cert: Union[Unset, Any] = UNSET
    p12password: Union[Unset, str] = UNSET
    skip: Union[Unset, list[str]] = UNSET
    orgname: Union[Unset, str] = UNSET
    locale: Union[Unset, str] = UNSET
    lang: Union[Unset, str] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        cert = self.cert

        p12password = self.p12password

        skip: Union[Unset, list[str]] = UNSET
        if not isinstance(self.skip, Unset):
            skip = self.skip

        orgname = self.orgname

        locale = self.locale

        lang = self.lang

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update({})
        if cert is not UNSET:
            field_dict["cert"] = cert
        if p12password is not UNSET:
            field_dict["p12password"] = p12password
        if skip is not UNSET:
            field_dict["skip"] = skip
        if orgname is not UNSET:
            field_dict["orgname"] = orgname
        if locale is not UNSET:
            field_dict["locale"] = locale
        if lang is not UNSET:
            field_dict["lang"] = lang

        return field_dict

    def to_multipart(self) -> types.RequestFiles:
        files: types.RequestFiles = []

        if not isinstance(self.cert, Unset):
            files.append(("cert", (None, str(self.cert).encode(), "text/plain")))

        if not isinstance(self.p12password, Unset):
            files.append(
                ("p12password", (None, str(self.p12password).encode(), "text/plain"))
            )

        if not isinstance(self.skip, Unset):
            for skip_item_element in self.skip:
                files.append(
                    ("skip", (None, str(skip_item_element).encode(), "text/plain"))
                )

        if not isinstance(self.orgname, Unset):
            files.append(("orgname", (None, str(self.orgname).encode(), "text/plain")))

        if not isinstance(self.locale, Unset):
            files.append(("locale", (None, str(self.locale).encode(), "text/plain")))

        if not isinstance(self.lang, Unset):
            files.append(("lang", (None, str(self.lang).encode(), "text/plain")))

        for prop_name, prop in self.additional_properties.items():
            files.append((prop_name, (None, str(prop).encode(), "text/plain")))

        return files

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        cert = d.pop("cert", UNSET)

        p12password = d.pop("p12password", UNSET)

        skip = cast(list[str], d.pop("skip", UNSET))

        orgname = d.pop("orgname", UNSET)

        locale = d.pop("locale", UNSET)

        lang = d.pop("lang", UNSET)

        prepare_prepare_device_body = cls(
            cert=cert,
            p12password=p12password,
            skip=skip,
            orgname=orgname,
            locale=locale,
            lang=lang,
        )

        prepare_prepare_device_body.additional_properties = d
        return prepare_prepare_device_body

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
