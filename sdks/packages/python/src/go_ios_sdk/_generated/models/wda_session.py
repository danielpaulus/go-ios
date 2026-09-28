from collections.abc import Mapping
from typing import (
    TYPE_CHECKING,
    Any,
    TypeVar,
)

from attrs import define as _attrs_define
from attrs import field as _attrs_field

if TYPE_CHECKING:
    from ..models.wda_config import WdaConfig


T = TypeVar("T", bound="WdaSession")


@_attrs_define
class WdaSession:
    """A running WebDriverAgent session.

    Attributes:
        config (WdaConfig): Configuration for launching a WebDriverAgent (XCUITest) runner session.
        session_id (str): Opaque session identifier.
        udid (str): The device udid the session runs on.
    """

    config: "WdaConfig"
    session_id: str
    udid: str
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        config = self.config.to_dict()

        session_id = self.session_id

        udid = self.udid

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "config": config,
                "sessionId": session_id,
                "udid": udid,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.wda_config import WdaConfig

        d = dict(src_dict)
        config = WdaConfig.from_dict(d.pop("config"))

        session_id = d.pop("sessionId")

        udid = d.pop("udid")

        wda_session = cls(
            config=config,
            session_id=session_id,
            udid=udid,
        )

        wda_session.additional_properties = d
        return wda_session

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
