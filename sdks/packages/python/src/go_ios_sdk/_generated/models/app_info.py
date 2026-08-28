from collections.abc import Mapping
from typing import (
    Any,
    TypeVar,
    Union,
)

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

T = TypeVar("T", bound="AppInfo")


@_attrs_define
class AppInfo:
    """Installed application metadata. This is an open map: keys come straight from
    the app's Info.plist. Common keys are surfaced for discoverability but any
    additional keys may be present.

        Attributes:
            cf_bundle_identifier (Union[Unset, str]):
            cf_bundle_executable (Union[Unset, str]):
            cf_bundle_name (Union[Unset, str]):
            cf_bundle_short_version_string (Union[Unset, str]):
            path (Union[Unset, str]):
            ui_file_sharing_enabled (Union[Unset, bool]):
    """

    cf_bundle_identifier: Union[Unset, str] = UNSET
    cf_bundle_executable: Union[Unset, str] = UNSET
    cf_bundle_name: Union[Unset, str] = UNSET
    cf_bundle_short_version_string: Union[Unset, str] = UNSET
    path: Union[Unset, str] = UNSET
    ui_file_sharing_enabled: Union[Unset, bool] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        cf_bundle_identifier = self.cf_bundle_identifier

        cf_bundle_executable = self.cf_bundle_executable

        cf_bundle_name = self.cf_bundle_name

        cf_bundle_short_version_string = self.cf_bundle_short_version_string

        path = self.path

        ui_file_sharing_enabled = self.ui_file_sharing_enabled

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update({})
        if cf_bundle_identifier is not UNSET:
            field_dict["CFBundleIdentifier"] = cf_bundle_identifier
        if cf_bundle_executable is not UNSET:
            field_dict["CFBundleExecutable"] = cf_bundle_executable
        if cf_bundle_name is not UNSET:
            field_dict["CFBundleName"] = cf_bundle_name
        if cf_bundle_short_version_string is not UNSET:
            field_dict["CFBundleShortVersionString"] = cf_bundle_short_version_string
        if path is not UNSET:
            field_dict["Path"] = path
        if ui_file_sharing_enabled is not UNSET:
            field_dict["UIFileSharingEnabled"] = ui_file_sharing_enabled

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        cf_bundle_identifier = d.pop("CFBundleIdentifier", UNSET)

        cf_bundle_executable = d.pop("CFBundleExecutable", UNSET)

        cf_bundle_name = d.pop("CFBundleName", UNSET)

        cf_bundle_short_version_string = d.pop("CFBundleShortVersionString", UNSET)

        path = d.pop("Path", UNSET)

        ui_file_sharing_enabled = d.pop("UIFileSharingEnabled", UNSET)

        app_info = cls(
            cf_bundle_identifier=cf_bundle_identifier,
            cf_bundle_executable=cf_bundle_executable,
            cf_bundle_name=cf_bundle_name,
            cf_bundle_short_version_string=cf_bundle_short_version_string,
            path=path,
            ui_file_sharing_enabled=ui_file_sharing_enabled,
        )

        app_info.additional_properties = d
        return app_info

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
