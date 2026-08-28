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

T = TypeVar("T", bound="DevicesSetWallpaperBody")


@_attrs_define
class DevicesSetWallpaperBody:
    """
    Attributes:
        image (Any): The wallpaper image.
        p12 (Any): The supervision identity certificate.
        password (Union[Unset, str]): Passphrase for the `.p12` identity.
        screen (Union[Unset, str]): Target screen (`home`, `lock`, `both`).
    """

    image: Any
    p12: Any
    password: Union[Unset, str] = UNSET
    screen: Union[Unset, str] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        image = self.image

        p12 = self.p12

        password = self.password

        screen = self.screen

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "image": image,
                "p12": p12,
            }
        )
        if password is not UNSET:
            field_dict["password"] = password
        if screen is not UNSET:
            field_dict["screen"] = screen

        return field_dict

    def to_multipart(self) -> types.RequestFiles:
        files: types.RequestFiles = []

        files.append(("image", (None, str(self.image).encode(), "text/plain")))

        files.append(("p12", (None, str(self.p12).encode(), "text/plain")))

        if not isinstance(self.password, Unset):
            files.append(
                ("password", (None, str(self.password).encode(), "text/plain"))
            )

        if not isinstance(self.screen, Unset):
            files.append(("screen", (None, str(self.screen).encode(), "text/plain")))

        for prop_name, prop in self.additional_properties.items():
            files.append((prop_name, (None, str(prop).encode(), "text/plain")))

        return files

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        image = d.pop("image")

        p12 = d.pop("p12")

        password = d.pop("password", UNSET)

        screen = d.pop("screen", UNSET)

        devices_set_wallpaper_body = cls(
            image=image,
            p12=p12,
            password=password,
            screen=screen,
        )

        devices_set_wallpaper_body.additional_properties = d
        return devices_set_wallpaper_body

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
