"""Contains all the data models used in inputs/outputs"""

from .accessibility_set_location_gpx_body import AccessibilitySetLocationGpxBody
from .agent_shutdown import AgentShutdown
from .app_info import AppInfo
from .app_state_notification import AppStateNotification
from .assistive_touch_state import AssistiveTouchState
from .attach_detach_event import AttachDetachEvent
from .ax_audit_issue import AXAuditIssue
from .ax_element import AXElement
from .ax_enabled_request import AXEnabledRequest
from .battery_info import BatteryInfo
from .battery_registry import BatteryRegistry
from .cloud_config import CloudConfig
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
from .disk_space_info import DiskSpaceInfo
from .enabled_request import EnabledRequest
from .file_domain_type_1 import FileDomainType1
from .file_entry import FileEntry
from .file_listing import FileListing
from .file_push_result import FilePushResult
from .forward_request import ForwardRequest
from .fsync_listing import FsyncListing
from .fsync_message import FsyncMessage
from .fsync_push_result import FsyncPushResult
from .fsync_tree_entry import FsyncTreeEntry
from .fsync_tree_listing import FsyncTreeListing
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
from .network_info import NetworkInfo
from .os_trace_entry import OsTraceEntry
from .pasteboard_content import PasteboardContent
from .prepare_prepare_device_body import PreparePrepareDeviceBody
from .prepare_result import PrepareResult
from .prepare_skip_options import PrepareSkipOptions
from .process_info import ProcessInfo
from .profile import Profile
from .profile_type import ProfileType
from .provisioning_result import ProvisioningResult
from .rsd_service_entry import RsdServiceEntry
from .rsd_services import RsdServices
from .run_test_request import RunTestRequest
from .run_test_request_env import RunTestRequestEnv
from .security_info import SecurityInfo
from .set_language_request import SetLanguageRequest
from .sign_app_body import SignAppBody
from .sign_certificate_body import SignCertificateBody
from .sign_provision_body import SignProvisionBody
from .status_ok import StatusOk
from .supervision_cert import SupervisionCert
from .syslog_message import SyslogMessage
from .time_format_request import TimeFormatRequest
from .time_format_state import TimeFormatState
from .tunnel import Tunnel
from .tunnel_stopped import TunnelStopped
from .ui_app_request import UIAppRequest
from .ui_button_request import UIButtonRequest
from .ui_long_press_request import UILongPressRequest
from .ui_orientation_request import UIOrientationRequest
from .ui_response import UIResponse
from .ui_swipe_request import UISwipeRequest
from .ui_tap_request import UITapRequest
from .ui_type_request import UITypeRequest
from .uiapi_request import UIAPIRequest
from .unlock_token import UnlockToken
from .voice_over_state import VoiceOverState
from .wda_config import WdaConfig
from .wda_config_env import WdaConfigEnv
from .wda_session import WdaSession
from .web_inspector_eval_request import WebInspectorEvalRequest
from .web_inspector_eval_result import WebInspectorEvalResult
from .web_inspector_launch_request import WebInspectorLaunchRequest
from .web_inspector_launch_result import WebInspectorLaunchResult
from .web_inspector_page import WebInspectorPage
from .wifi_request import WifiRequest
from .zoom_touch_state import ZoomTouchState

__all__ = (
    "AccessibilitySetLocationGpxBody",
    "AgentShutdown",
    "AppInfo",
    "AppStateNotification",
    "AssistiveTouchState",
    "AttachDetachEvent",
    "AXAuditIssue",
    "AXElement",
    "AXEnabledRequest",
    "BatteryInfo",
    "BatteryRegistry",
    "CloudConfig",
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
    "DiskSpaceInfo",
    "EnabledRequest",
    "FileDomainType1",
    "FileEntry",
    "FileListing",
    "FilePushResult",
    "ForwardRequest",
    "FsyncListing",
    "FsyncMessage",
    "FsyncPushResult",
    "FsyncTreeEntry",
    "FsyncTreeListing",
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
    "NetworkInfo",
    "OsTraceEntry",
    "PasteboardContent",
    "PreparePrepareDeviceBody",
    "PrepareResult",
    "PrepareSkipOptions",
    "ProcessInfo",
    "Profile",
    "ProfileType",
    "ProvisioningResult",
    "RsdServiceEntry",
    "RsdServices",
    "RunTestRequest",
    "RunTestRequestEnv",
    "SecurityInfo",
    "SetLanguageRequest",
    "SignAppBody",
    "SignCertificateBody",
    "SignProvisionBody",
    "StatusOk",
    "SupervisionCert",
    "SyslogMessage",
    "TimeFormatRequest",
    "TimeFormatState",
    "Tunnel",
    "TunnelStopped",
    "UIAPIRequest",
    "UIAppRequest",
    "UIButtonRequest",
    "UILongPressRequest",
    "UIOrientationRequest",
    "UIResponse",
    "UISwipeRequest",
    "UITapRequest",
    "UITypeRequest",
    "UnlockToken",
    "VoiceOverState",
    "WdaConfig",
    "WdaConfigEnv",
    "WdaSession",
    "WebInspectorEvalRequest",
    "WebInspectorEvalResult",
    "WebInspectorLaunchRequest",
    "WebInspectorLaunchResult",
    "WebInspectorPage",
    "WifiRequest",
    "ZoomTouchState",
)
