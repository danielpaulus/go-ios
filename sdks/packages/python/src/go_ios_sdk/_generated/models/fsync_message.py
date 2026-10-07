from collections.abc import Mapping
from typing import Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field

T = TypeVar("T", bound="FsyncMessage")


@_attrs_define
class FsyncMessage:
    """`POST /device/{udid}/fsync/mkdir` and `DELETE /device/{udid}/fsync/rm` —
    simple message + path acknowledgement.

        Attributes:
            message (str): Human-readable result message (e.g. `created`, `removed`).
            path (str): The (cleaned) device path acted on.
    """

    message: str
    path: str
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        message = self.message

        path = self.path

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "message": message,
                "path": path,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        message = d.pop("message")

        path = d.pop("path")

        fsync_message = cls(
            message=message,
            path=path,
        )

        fsync_message.additional_properties = d
        return fsync_message

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
