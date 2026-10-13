from collections.abc import Mapping
from typing import (
    Any,
    TypeVar,
    Union,
    cast,
)

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

T = TypeVar("T", bound="LanguageConfiguration")


@_attrs_define
class LanguageConfiguration:
    """Language/locale configuration (`ios.LanguageConfiguration`), returned by
    `GET/PUT /device/{udid}/lang`.

        Attributes:
            language (Union[Unset, str]):
            locale (Union[Unset, str]):
            supported_locales (Union[Unset, list[str]]): Supported locales advertised by the device.
            supported_languages (Union[Unset, list[str]]): Supported UI languages advertised by the device.
    """

    language: Union[Unset, str] = UNSET
    locale: Union[Unset, str] = UNSET
    supported_locales: Union[Unset, list[str]] = UNSET
    supported_languages: Union[Unset, list[str]] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        language = self.language

        locale = self.locale

        supported_locales: Union[Unset, list[str]] = UNSET
        if not isinstance(self.supported_locales, Unset):
            supported_locales = self.supported_locales

        supported_languages: Union[Unset, list[str]] = UNSET
        if not isinstance(self.supported_languages, Unset):
            supported_languages = self.supported_languages

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update({})
        if language is not UNSET:
            field_dict["Language"] = language
        if locale is not UNSET:
            field_dict["Locale"] = locale
        if supported_locales is not UNSET:
            field_dict["SupportedLocales"] = supported_locales
        if supported_languages is not UNSET:
            field_dict["SupportedLanguages"] = supported_languages

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        language = d.pop("Language", UNSET)

        locale = d.pop("Locale", UNSET)

        supported_locales = cast(list[str], d.pop("SupportedLocales", UNSET))

        supported_languages = cast(list[str], d.pop("SupportedLanguages", UNSET))

        language_configuration = cls(
            language=language,
            locale=locale,
            supported_locales=supported_locales,
            supported_languages=supported_languages,
        )

        language_configuration.additional_properties = d
        return language_configuration

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
