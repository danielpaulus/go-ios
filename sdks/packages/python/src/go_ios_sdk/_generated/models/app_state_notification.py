from collections.abc import Mapping
from typing import (
    Any,
    TypeVar,
    Union,
)

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

T = TypeVar("T", bound="AppStateNotification")


@_attrs_define
class AppStateNotification:
    """An app foreground/background/lifecycle state change.

    Attributes:
        bundle_id (str): Bundle id of the app whose state changed.
        state (str): New application state.
            Typical values: `foreground`, `background`, `suspended`, `terminated`,
            `unknown`.
        timestamp (Union[Unset, int]): Unix epoch milliseconds when the change was observed.
    """

    bundle_id: str
    state: str
    timestamp: Union[Unset, int] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        bundle_id = self.bundle_id

        state = self.state

        timestamp = self.timestamp

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "bundleId": bundle_id,
                "state": state,
            }
        )
        if timestamp is not UNSET:
            field_dict["timestamp"] = timestamp

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        bundle_id = d.pop("bundleId")

        state = d.pop("state")

        timestamp = d.pop("timestamp", UNSET)

        app_state_notification = cls(
            bundle_id=bundle_id,
            state=state,
            timestamp=timestamp,
        )

        app_state_notification.additional_properties = d
        return app_state_notification

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
