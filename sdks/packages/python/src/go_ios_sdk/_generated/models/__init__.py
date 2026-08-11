"""Contains all the data models used in inputs/outputs"""

from .agent_shutdown import AgentShutdown
from .app_info import AppInfo
from .app_state_notification import AppStateNotification
from .assistive_touch_state import AssistiveTouchState
from .attach_detach_event import AttachDetachEvent
from .battery_info import BatteryInfo
from .cpu_usage_sample import CpuUsageSample
from .crash_listing import CrashListing
from .dev_mode_request import DevModeRequest
from .dev_mode_state import DevModeState
from .device_date import DeviceDate
from .device_entry import DeviceEntry
from .device_info import DeviceInfo
from .device_list import DeviceList
from .device_name import DeviceName
from .device_properties import DeviceProperties
from .devices_add_profile_body import DevicesAddProfileBody
from .devices_install_app_body import DevicesInstallAppBody
from .devices_mdm_clear_passcode_body import DevicesMdmClearPasscodeBody
from .devices_mdm_clear_screen_time_password_body import (
    DevicesMdmClearScreenTimePasswordBody,
)
from .devices_mdm_fetch_unlock_token_body import DevicesMdmFetchUnlockTokenBody
from .devices_mdm_security_info_body import DevicesMdmSecurityInfoBody
from .devices_pair_body import DevicesPairBody
from .devices_set_http_proxy_body import DevicesSetHttpProxyBody
from .devices_set_wallpaper_body import DevicesSetWallpaperBody
from .diagnostics import Diagnostics
from .enabled_request import EnabledRequest
from .file_domain_type_1 import FileDomainType1
from .file_entry import FileEntry
from .file_listing import FileListing
from .file_push_result import FilePushResult
from .forward_request import ForwardRequest
from .generic_response import GenericResponse
from .heartbeat import Heartbeat
from .icon_layout import IconLayout
from .installed_profiles import InstalledProfiles
from .job import Job
from .job_log_line import JobLogLine
from .job_status_type_1 import JobStatusType1
from .language_configuration import LanguageConfiguration
from .lockdown_values import LockdownValues
from .mem_limit_request import MemLimitRequest
from .mem_limit_result import MemLimitResult
from .mobile_gestalt import MobileGestalt
from .mounted_images import MountedImages
from .os_trace_entry import OsTraceEntry
from .pasteboard_content import PasteboardContent
from .process_info import ProcessInfo
from .profile import Profile
from .profile_type import ProfileType
from .run_test_request import RunTestRequest
from .run_test_request_env import RunTestRequestEnv
from .security_info import SecurityInfo
from .set_language_request import SetLanguageRequest
from .status_ok import StatusOk
from .syslog_message import SyslogMessage
from .time_format_request import TimeFormatRequest
from .time_format_state import TimeFormatState
from .tunnel import Tunnel
from .tunnel_stopped import TunnelStopped
from .unlock_token import UnlockToken
from .wda_config import WdaConfig
from .wda_config_env import WdaConfigEnv
from .wda_session import WdaSession
from .wifi_request import WifiRequest

__all__ = (
    "AgentShutdown",
    "AppInfo",
    "AppStateNotification",
    "AssistiveTouchState",
    "AttachDetachEvent",
    "BatteryInfo",
    "CpuUsageSample",
    "CrashListing",
    "DeviceDate",
    "DeviceEntry",
    "DeviceInfo",
    "DeviceList",
    "DeviceName",
    "DeviceProperties",
    "DevicesAddProfileBody",
    "DevicesInstallAppBody",
    "DevicesMdmClearPasscodeBody",
    "DevicesMdmClearScreenTimePasswordBody",
    "DevicesMdmFetchUnlockTokenBody",
    "DevicesMdmSecurityInfoBody",
    "DevicesPairBody",
    "DevicesSetHttpProxyBody",
    "DevicesSetWallpaperBody",
    "DevModeRequest",
    "DevModeState",
    "Diagnostics",
    "EnabledRequest",
    "FileDomainType1",
    "FileEntry",
    "FileListing",
    "FilePushResult",
    "ForwardRequest",
    "GenericResponse",
    "Heartbeat",
    "IconLayout",
    "InstalledProfiles",
    "Job",
    "JobLogLine",
    "JobStatusType1",
    "LanguageConfiguration",
    "LockdownValues",
    "MemLimitRequest",
    "MemLimitResult",
    "MobileGestalt",
    "MountedImages",
    "OsTraceEntry",
    "PasteboardContent",
    "ProcessInfo",
    "Profile",
    "ProfileType",
    "RunTestRequest",
    "RunTestRequestEnv",
    "SecurityInfo",
    "SetLanguageRequest",
    "StatusOk",
    "SyslogMessage",
    "TimeFormatRequest",
    "TimeFormatState",
    "Tunnel",
    "TunnelStopped",
    "UnlockToken",
    "WdaConfig",
    "WdaConfigEnv",
    "WdaSession",
    "WifiRequest",
)
