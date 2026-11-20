import datetime
from collections.abc import Mapping
from typing import (
    Any,
    TypeVar,
    Union,
    cast,
)

from attrs import define as _attrs_define
from attrs import field as _attrs_field
from dateutil.parser import isoparse

from ..models.job_status_type_1 import JobStatusType1
from ..types import UNSET, Unset

T = TypeVar("T", bound="Job")


@_attrs_define
class Job:
    """A long-running operation started via the REST API (test run, WDA runner,
    port forward). Mirrors the server's `jobView`.

        Attributes:
            id (str): Opaque job id, e.g. `runtest-3`.
            kind (str): Job kind: `runtest`, `runwda` or `forward`.
            udid (str): The device udid the job runs on.
            status (Union[JobStatusType1, str]): Job lifecycle state.
            started_at (datetime.datetime): When the job started (ISO-8601).
            finished_at (Union[Unset, datetime.datetime]): When the job reached a terminal state (absent while running).
            error (Union[Unset, str]): Error message when `status` is `failed`.
            result (Union[Unset, Any]): Terminal result payload (e.g. test suites) when the job succeeded.
    """

    id: str
    kind: str
    udid: str
    status: Union[JobStatusType1, str]
    started_at: datetime.datetime
    finished_at: Union[Unset, datetime.datetime] = UNSET
    error: Union[Unset, str] = UNSET
    result: Union[Unset, Any] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        id = self.id

        kind = self.kind

        udid = self.udid

        status: str
        if isinstance(self.status, JobStatusType1):
            status = self.status.value
        else:
            status = self.status

        started_at = self.started_at.isoformat()

        finished_at: Union[Unset, str] = UNSET
        if not isinstance(self.finished_at, Unset):
            finished_at = self.finished_at.isoformat()

        error = self.error

        result = self.result

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "kind": kind,
                "udid": udid,
                "status": status,
                "startedAt": started_at,
            }
        )
        if finished_at is not UNSET:
            field_dict["finishedAt"] = finished_at
        if error is not UNSET:
            field_dict["error"] = error
        if result is not UNSET:
            field_dict["result"] = result

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        id = d.pop("id")

        kind = d.pop("kind")

        udid = d.pop("udid")

        def _parse_status(data: object) -> Union[JobStatusType1, str]:
            try:
                if not isinstance(data, str):
                    raise TypeError()
                componentsschemas_job_status_type_1 = JobStatusType1(data)

                return componentsschemas_job_status_type_1
            except:  # noqa: E722
                pass
            return cast(Union[JobStatusType1, str], data)

        status = _parse_status(d.pop("status"))

        started_at = isoparse(d.pop("startedAt"))

        _finished_at = d.pop("finishedAt", UNSET)
        finished_at: Union[Unset, datetime.datetime]
        if isinstance(_finished_at, Unset):
            finished_at = UNSET
        else:
            finished_at = isoparse(_finished_at)

        error = d.pop("error", UNSET)

        result = d.pop("result", UNSET)

        job = cls(
            id=id,
            kind=kind,
            udid=udid,
            status=status,
            started_at=started_at,
            finished_at=finished_at,
            error=error,
            result=result,
        )

        job.additional_properties = d
        return job

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
