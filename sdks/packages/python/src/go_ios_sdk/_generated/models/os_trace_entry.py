from collections.abc import Mapping
from typing import (
    Any,
    TypeVar,
    Union,
)

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

T = TypeVar("T", bound="OsTraceEntry")


@_attrs_define
class OsTraceEntry:
    """A structured os_log trace entry.

    Attributes:
        message (str): The formatted log message.
        pid (Union[Unset, int]): Process id that emitted the entry.
        process_name (Union[Unset, str]): Emitting process/executable name.
        level (Union[Unset, str]): Log level, e.g. `default`, `info`, `debug`, `error`, `fault`.
        subsystem (Union[Unset, str]): Subsystem string (e.g. `com.apple.network`).
        category (Union[Unset, str]): Category within the subsystem.
        timestamp (Union[Unset, int]): Unix epoch milliseconds when the entry was emitted, if known.
    """

    message: str
    pid: Union[Unset, int] = UNSET
    process_name: Union[Unset, str] = UNSET
    level: Union[Unset, str] = UNSET
    subsystem: Union[Unset, str] = UNSET
    category: Union[Unset, str] = UNSET
    timestamp: Union[Unset, int] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        message = self.message

        pid = self.pid

        process_name = self.process_name

        level = self.level

        subsystem = self.subsystem

        category = self.category

        timestamp = self.timestamp

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "message": message,
            }
        )
        if pid is not UNSET:
            field_dict["pid"] = pid
        if process_name is not UNSET:
            field_dict["processName"] = process_name
        if level is not UNSET:
            field_dict["level"] = level
        if subsystem is not UNSET:
            field_dict["subsystem"] = subsystem
        if category is not UNSET:
            field_dict["category"] = category
        if timestamp is not UNSET:
            field_dict["timestamp"] = timestamp

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        message = d.pop("message")

        pid = d.pop("pid", UNSET)

        process_name = d.pop("processName", UNSET)

        level = d.pop("level", UNSET)

        subsystem = d.pop("subsystem", UNSET)

        category = d.pop("category", UNSET)

        timestamp = d.pop("timestamp", UNSET)

        os_trace_entry = cls(
            message=message,
            pid=pid,
            process_name=process_name,
            level=level,
            subsystem=subsystem,
            category=category,
            timestamp=timestamp,
        )

        os_trace_entry.additional_properties = d
        return os_trace_entry

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
