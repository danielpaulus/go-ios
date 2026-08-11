from collections.abc import Mapping
from typing import (
    TYPE_CHECKING,
    Any,
    TypeVar,
    Union,
)

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

if TYPE_CHECKING:
    from ..models.profile import Profile


T = TypeVar("T", bound="ProfileType")


@_attrs_define
class ProfileType:
    """A condition inducer profile type (e.g. thermal, network) with its variants.

    Attributes:
        identifier (str):
        name (str):
        profiles (list['Profile']):
        active_profile (Union[Unset, str]):
        profiles_sorted (Union[Unset, bool]):
        is_active (Union[Unset, bool]):
        is_destructive (Union[Unset, bool]):
        is_internal (Union[Unset, bool]):
    """

    identifier: str
    name: str
    profiles: list["Profile"]
    active_profile: Union[Unset, str] = UNSET
    profiles_sorted: Union[Unset, bool] = UNSET
    is_active: Union[Unset, bool] = UNSET
    is_destructive: Union[Unset, bool] = UNSET
    is_internal: Union[Unset, bool] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        identifier = self.identifier

        name = self.name

        profiles = []
        for profiles_item_data in self.profiles:
            profiles_item = profiles_item_data.to_dict()
            profiles.append(profiles_item)

        active_profile = self.active_profile

        profiles_sorted = self.profiles_sorted

        is_active = self.is_active

        is_destructive = self.is_destructive

        is_internal = self.is_internal

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "identifier": identifier,
                "name": name,
                "profiles": profiles,
            }
        )
        if active_profile is not UNSET:
            field_dict["activeProfile"] = active_profile
        if profiles_sorted is not UNSET:
            field_dict["profilesSorted"] = profiles_sorted
        if is_active is not UNSET:
            field_dict["isActive"] = is_active
        if is_destructive is not UNSET:
            field_dict["isDestructive"] = is_destructive
        if is_internal is not UNSET:
            field_dict["isInternal"] = is_internal

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.profile import Profile

        d = dict(src_dict)
        identifier = d.pop("identifier")

        name = d.pop("name")

        profiles = []
        _profiles = d.pop("profiles")
        for profiles_item_data in _profiles:
            profiles_item = Profile.from_dict(profiles_item_data)

            profiles.append(profiles_item)

        active_profile = d.pop("activeProfile", UNSET)

        profiles_sorted = d.pop("profilesSorted", UNSET)

        is_active = d.pop("isActive", UNSET)

        is_destructive = d.pop("isDestructive", UNSET)

        is_internal = d.pop("isInternal", UNSET)

        profile_type = cls(
            identifier=identifier,
            name=name,
            profiles=profiles,
            active_profile=active_profile,
            profiles_sorted=profiles_sorted,
            is_active=is_active,
            is_destructive=is_destructive,
            is_internal=is_internal,
        )

        profile_type.additional_properties = d
        return profile_type

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
