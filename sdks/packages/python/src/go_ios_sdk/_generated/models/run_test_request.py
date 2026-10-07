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
    from ..models.run_test_request_env import RunTestRequestEnv


T = TypeVar("T", bound="RunTestRequest")


@_attrs_define
class RunTestRequest:
    """`POST /device/{udid}/jobs/runtest` (and `runwda`) request.

    Attributes:
        bundle_id (Union[Unset, str]): Bundle id of the app under test.
        test_runner_bundle_id (Union[Unset, str]): Bundle id of the test runner. Defaults to `bundleId` if omitted.
        xctest_config (Union[Unset, str]): Name of the `.xctestconfiguration`.
        env (Union[Unset, RunTestRequestEnv]): Extra environment variables for the test runner.
        args (Union[Unset, list[str]]): Extra process arguments for the test runner.
        tests_to_run (Union[Unset, list[str]]): Only run these tests (`Class/method` identifiers).
        tests_to_skip (Union[Unset, list[str]]): Skip these tests.
        xctest (Union[Unset, bool]): Run as a plain XCTest (vs XCUITest).
    """

    bundle_id: Union[Unset, str] = UNSET
    test_runner_bundle_id: Union[Unset, str] = UNSET
    xctest_config: Union[Unset, str] = UNSET
    env: Union[Unset, "RunTestRequestEnv"] = UNSET
    args: Union[Unset, list[str]] = UNSET
    tests_to_run: Union[Unset, list[str]] = UNSET
    tests_to_skip: Union[Unset, list[str]] = UNSET
    xctest: Union[Unset, bool] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        bundle_id = self.bundle_id

        test_runner_bundle_id = self.test_runner_bundle_id

        xctest_config = self.xctest_config

        env: Union[Unset, dict[str, Any]] = UNSET
        if not isinstance(self.env, Unset):
            env = self.env.to_dict()

        args: Union[Unset, list[str]] = UNSET
        if not isinstance(self.args, Unset):
            args = self.args

        tests_to_run: Union[Unset, list[str]] = UNSET
        if not isinstance(self.tests_to_run, Unset):
            tests_to_run = self.tests_to_run

        tests_to_skip: Union[Unset, list[str]] = UNSET
        if not isinstance(self.tests_to_skip, Unset):
            tests_to_skip = self.tests_to_skip

        xctest = self.xctest

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update({})
        if bundle_id is not UNSET:
            field_dict["bundleId"] = bundle_id
        if test_runner_bundle_id is not UNSET:
            field_dict["testRunnerBundleId"] = test_runner_bundle_id
        if xctest_config is not UNSET:
            field_dict["xctestConfig"] = xctest_config
        if env is not UNSET:
            field_dict["env"] = env
        if args is not UNSET:
            field_dict["args"] = args
        if tests_to_run is not UNSET:
            field_dict["testsToRun"] = tests_to_run
        if tests_to_skip is not UNSET:
            field_dict["testsToSkip"] = tests_to_skip
        if xctest is not UNSET:
            field_dict["xctest"] = xctest

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.run_test_request_env import RunTestRequestEnv

        d = dict(src_dict)
        bundle_id = d.pop("bundleId", UNSET)

        test_runner_bundle_id = d.pop("testRunnerBundleId", UNSET)

        xctest_config = d.pop("xctestConfig", UNSET)

        _env = d.pop("env", UNSET)
        env: Union[Unset, RunTestRequestEnv]
        if isinstance(_env, Unset):
            env = UNSET
        else:
            env = RunTestRequestEnv.from_dict(_env)

        args = cast(list[str], d.pop("args", UNSET))

        tests_to_run = cast(list[str], d.pop("testsToRun", UNSET))

        tests_to_skip = cast(list[str], d.pop("testsToSkip", UNSET))

        xctest = d.pop("xctest", UNSET)

        run_test_request = cls(
            bundle_id=bundle_id,
            test_runner_bundle_id=test_runner_bundle_id,
            xctest_config=xctest_config,
            env=env,
            args=args,
            tests_to_run=tests_to_run,
            tests_to_skip=tests_to_skip,
            xctest=xctest,
        )

        run_test_request.additional_properties = d
        return run_test_request

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
