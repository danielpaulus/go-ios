import datetime
from collections.abc import Mapping
from typing import (
    Any,
    TypeVar,
    Union,
)

from attrs import define as _attrs_define
from attrs import field as _attrs_field
from dateutil.parser import isoparse

from ..types import UNSET, Unset

T = TypeVar("T", bound="ProcessInfo")


@_attrs_define
class ProcessInfo:
    """A running process entry (`instruments.ProcessInfo`) from
    `GET /device/{udid}/processes`.

        Attributes:
            pid (int):
            name (str):
            real_app_name (Union[Unset, str]):
            is_application (Union[Unset, bool]):
            start_date (Union[Unset, datetime.datetime]): Process start time, ISO-8601.
    """

    pid: int
    name: str
    real_app_name: Union[Unset, str] = UNSET
    is_application: Union[Unset, bool] = UNSET
    start_date: Union[Unset, datetime.datetime] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        pid = self.pid

        name = self.name

        real_app_name = self.real_app_name

        is_application = self.is_application

        start_date: Union[Unset, str] = UNSET
        if not isinstance(self.start_date, Unset):
            start_date = self.start_date.isoformat()

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "pid": pid,
                "name": name,
            }
        )
        if real_app_name is not UNSET:
            field_dict["realAppName"] = real_app_name
        if is_application is not UNSET:
            field_dict["isApplication"] = is_application
        if start_date is not UNSET:
            field_dict["startDate"] = start_date

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        pid = d.pop("pid")

        name = d.pop("name")

        real_app_name = d.pop("realAppName", UNSET)

        is_application = d.pop("isApplication", UNSET)

        _start_date = d.pop("startDate", UNSET)
        start_date: Union[Unset, datetime.datetime]
        if isinstance(_start_date, Unset):
            start_date = UNSET
        else:
            start_date = isoparse(_start_date)

        process_info = cls(
            pid=pid,
            name=name,
            real_app_name=real_app_name,
            is_application=is_application,
            start_date=start_date,
        )

        process_info.additional_properties = d
        return process_info

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
