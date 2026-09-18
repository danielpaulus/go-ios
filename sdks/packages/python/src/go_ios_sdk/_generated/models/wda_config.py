from collections.abc import Mapping
from typing import (
    TYPE_CHECKING,
    Any,
    TypeVar,
    Union,
    cast,
)

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

if TYPE_CHECKING:
    from ..models.wda_config_env import WdaConfigEnv


T = TypeVar("T", bound="WdaConfig")


@_attrs_define
class WdaConfig:
    """Configuration for launching a WebDriverAgent (XCUITest) runner session.

    Attributes:
        bundle_id (str): Bundle id of the WDA runner host app (e.g. `com.facebook.WebDriverAgentRunner.xctrunner`).
        test_bundle_id (str): Bundle id of the XCTest test bundle.
        xc_test_config (str): Path/name of the `.xctestconfiguration` to use.
        args (Union[Unset, list[str]]): Extra process arguments passed to the runner.
        env (Union[Unset, WdaConfigEnv]): Extra environment variables passed to the runner.
    """

    bundle_id: str
    test_bundle_id: str
    xc_test_config: str
    args: Union[Unset, list[str]] = UNSET
    env: Union[Unset, "WdaConfigEnv"] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        bundle_id = self.bundle_id

        test_bundle_id = self.test_bundle_id

        xc_test_config = self.xc_test_config

        args: Union[Unset, list[str]] = UNSET
        if not isinstance(self.args, Unset):
            args = self.args

        env: Union[Unset, dict[str, Any]] = UNSET
        if not isinstance(self.env, Unset):
            env = self.env.to_dict()

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "bundleId": bundle_id,
                "testBundleId": test_bundle_id,
                "xcTestConfig": xc_test_config,
            }
        )
        if args is not UNSET:
            field_dict["args"] = args
        if env is not UNSET:
            field_dict["env"] = env

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.wda_config_env import WdaConfigEnv

        d = dict(src_dict)
        bundle_id = d.pop("bundleId")

        test_bundle_id = d.pop("testBundleId")

        xc_test_config = d.pop("xcTestConfig")

        args = cast(list[str], d.pop("args", UNSET))

        _env = d.pop("env", UNSET)
        env: Union[Unset, WdaConfigEnv]
        if isinstance(_env, Unset):
            env = UNSET
        else:
            env = WdaConfigEnv.from_dict(_env)

        wda_config = cls(
            bundle_id=bundle_id,
            test_bundle_id=test_bundle_id,
            xc_test_config=xc_test_config,
            args=args,
            env=env,
        )

        wda_config.additional_properties = d
        return wda_config

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
