# GoIos.Sdk.Generated.Api.DefaultApi

All URIs are relative to *http://localhost:60105*

| Method | HTTP request | Description |
|--------|--------------|-------------|
| [**AccessibilityGetAxSnapshot**](DefaultApi.md#accessibilitygetaxsnapshot) | **GET** /api/v1/device/{udid}/ax | Get accessibility element snapshot |
| [**AccessibilityGetVoiceOver**](DefaultApi.md#accessibilitygetvoiceover) | **GET** /api/v1/device/{udid}/voiceover | Get VoiceOver state |
| [**AccessibilityGetZoomTouch**](DefaultApi.md#accessibilitygetzoomtouch) | **GET** /api/v1/device/{udid}/zoom | Get ZoomTouch state |
| [**AccessibilityRunAxAudit**](DefaultApi.md#accessibilityrunaxaudit) | **POST** /api/v1/device/{udid}/ax/audit | Run accessibility audit |
| [**AccessibilitySetLocationGpx**](DefaultApi.md#accessibilitysetlocationgpx) | **PUT** /api/v1/device/{udid}/setlocation/gpx | Simulate location from a GPX file |
| [**AccessibilitySetVoiceOver**](DefaultApi.md#accessibilitysetvoiceover) | **PUT** /api/v1/device/{udid}/voiceover | Set VoiceOver state |
| [**AccessibilitySetZoomTouch**](DefaultApi.md#accessibilitysetzoomtouch) | **PUT** /api/v1/device/{udid}/zoom | Set ZoomTouch state |
| [**DevicesActivate**](DefaultApi.md#devicesactivate) | **POST** /api/v1/device/{udid}/activate | Activate device |
| [**DevicesAddProfile**](DefaultApi.md#devicesaddprofile) | **POST** /api/v1/device/{udid}/profiles | Install profile |
| [**DevicesCreateWdaSession**](DefaultApi.md#devicescreatewdasession) | **POST** /api/v1/device/{udid}/wda/session | Start WDA session |
| [**DevicesDeleteWdaSession**](DefaultApi.md#devicesdeletewdasession) | **DELETE** /api/v1/device/{udid}/wda/session/{sessionId} | Stop WDA session |
| [**DevicesDisableCondition**](DefaultApi.md#devicesdisablecondition) | **POST** /api/v1/device/{udid}/disable-condition | Disable condition |
| [**DevicesEnableCondition**](DefaultApi.md#devicesenablecondition) | **PUT** /api/v1/device/{udid}/enable-condition | Enable condition |
| [**DevicesErase**](DefaultApi.md#deviceserase) | **POST** /api/v1/device/{udid}/erase | Erase device |
| [**DevicesGetAssistiveTouch**](DefaultApi.md#devicesgetassistivetouch) | **GET** /api/v1/device/{udid}/assistivetouch | Get AssistiveTouch |
| [**DevicesGetBattery**](DefaultApi.md#devicesgetbattery) | **GET** /api/v1/device/{udid}/battery | Get battery info |
| [**DevicesGetDevMode**](DefaultApi.md#devicesgetdevmode) | **GET** /api/v1/device/{udid}/devmode | Get developer mode |
| [**DevicesGetDeviceDate**](DefaultApi.md#devicesgetdevicedate) | **GET** /api/v1/device/{udid}/date | Get device date |
| [**DevicesGetDeviceName**](DefaultApi.md#devicesgetdevicename) | **GET** /api/v1/device/{udid}/devicename | Get device name |
| [**DevicesGetDiagnostics**](DefaultApi.md#devicesgetdiagnostics) | **GET** /api/v1/device/{udid}/diagnostics | List diagnostics |
| [**DevicesGetIconLayout**](DefaultApi.md#devicesgeticonlayout) | **GET** /api/v1/device/{udid}/icon-layout | Get icon layout |
| [**DevicesGetInfo**](DefaultApi.md#devicesgetinfo) | **GET** /api/v1/device/{udid}/info | Get device info |
| [**DevicesGetJob**](DefaultApi.md#devicesgetjob) | **GET** /api/v1/device/{udid}/jobs/{id} | Get job |
| [**DevicesGetLanguage**](DefaultApi.md#devicesgetlanguage) | **GET** /api/v1/device/{udid}/lang | Get language |
| [**DevicesGetLockdownValues**](DefaultApi.md#devicesgetlockdownvalues) | **GET** /api/v1/device/{udid}/lockdown | Get lockdown values |
| [**DevicesGetMobileGestalt**](DefaultApi.md#devicesgetmobilegestalt) | **GET** /api/v1/device/{udid}/mobilegestalt | Query MobileGestalt |
| [**DevicesGetPasteboard**](DefaultApi.md#devicesgetpasteboard) | **GET** /api/v1/device/{udid}/pasteboard | Get pasteboard |
| [**DevicesGetProcesses**](DefaultApi.md#devicesgetprocesses) | **GET** /api/v1/device/{udid}/processes | List processes |
| [**DevicesGetProfiles**](DefaultApi.md#devicesgetprofiles) | **GET** /api/v1/device/{udid}/profiles | List configuration profiles |
| [**DevicesGetTimeFormat**](DefaultApi.md#devicesgettimeformat) | **GET** /api/v1/device/{udid}/timeformat | Get time format |
| [**DevicesGetWallpaper**](DefaultApi.md#devicesgetwallpaper) | **GET** /api/v1/device/{udid}/wallpaper | Get wallpaper |
| [**DevicesGetWdaSession**](DefaultApi.md#devicesgetwdasession) | **GET** /api/v1/device/{udid}/wda/session/{sessionId} | Get WDA session |
| [**DevicesInstallApp**](DefaultApi.md#devicesinstallapp) | **POST** /api/v1/device/{udid}/apps/install | Install app |
| [**DevicesKillApp**](DefaultApi.md#deviceskillapp) | **POST** /api/v1/device/{udid}/apps/kill | Kill app |
| [**DevicesLaunchApp**](DefaultApi.md#deviceslaunchapp) | **POST** /api/v1/device/{udid}/apps/launch | Launch app |
| [**DevicesListApps**](DefaultApi.md#deviceslistapps) | **GET** /api/v1/device/{udid}/apps/ | List apps |
| [**DevicesListConditions**](DefaultApi.md#deviceslistconditions) | **GET** /api/v1/device/{udid}/conditions | List conditions |
| [**DevicesListCrashes**](DefaultApi.md#deviceslistcrashes) | **GET** /api/v1/device/{udid}/crashes | List crash reports |
| [**DevicesListFiles**](DefaultApi.md#deviceslistfiles) | **GET** /api/v1/device/{udid}/files | List files |
| [**DevicesListImages**](DefaultApi.md#deviceslistimages) | **GET** /api/v1/device/{udid}/image | List mounted developer images |
| [**DevicesListJobs**](DefaultApi.md#deviceslistjobs) | **GET** /api/v1/device/{udid}/jobs | List jobs |
| [**DevicesListMountedImages**](DefaultApi.md#deviceslistmountedimages) | **GET** /api/v1/device/{udid}/image/list | List mounted images |
| [**DevicesMdmClearPasscode**](DefaultApi.md#devicesmdmclearpasscode) | **POST** /api/v1/device/{udid}/mdm/clear-passcode | Clear passcode (supervised) |
| [**DevicesMdmClearScreenTimePassword**](DefaultApi.md#devicesmdmclearscreentimepassword) | **POST** /api/v1/device/{udid}/mdm/clear-screen-time-password | Clear Screen Time password (supervised) |
| [**DevicesMdmFetchUnlockToken**](DefaultApi.md#devicesmdmfetchunlocktoken) | **POST** /api/v1/device/{udid}/mdm/fetch-unlock-token | Fetch unlock token (supervised) |
| [**DevicesMdmSecurityInfo**](DefaultApi.md#devicesmdmsecurityinfo) | **POST** /api/v1/device/{udid}/mdm/security-info | Get MDM security info (supervised) |
| [**DevicesMemLimitOff**](DefaultApi.md#devicesmemlimitoff) | **POST** /api/v1/device/{udid}/memlimitoff | Waive memory limit |
| [**DevicesMountImage**](DefaultApi.md#devicesmountimage) | **PUT** /api/v1/device/{udid}/image | Mount a developer image |
| [**DevicesPair**](DefaultApi.md#devicespair) | **POST** /api/v1/device/{udid}/pair | Pair device |
| [**DevicesPullFile**](DefaultApi.md#devicespullfile) | **GET** /api/v1/device/{udid}/files/pull | Pull file |
| [**DevicesPushFile**](DefaultApi.md#devicespushfile) | **POST** /api/v1/device/{udid}/files/push | Push file |
| [**DevicesReboot**](DefaultApi.md#devicesreboot) | **POST** /api/v1/device/{udid}/reboot | Reboot device |
| [**DevicesRemoveCrashes**](DefaultApi.md#devicesremovecrashes) | **DELETE** /api/v1/device/{udid}/crashes | Delete crash reports |
| [**DevicesRemoveHttpProxy**](DefaultApi.md#devicesremovehttpproxy) | **DELETE** /api/v1/device/{udid}/httpproxy | Remove HTTP proxy |
| [**DevicesRemoveProfile**](DefaultApi.md#devicesremoveprofile) | **DELETE** /api/v1/device/{udid}/profiles/{name} | Remove profile |
| [**DevicesRemoveWifi**](DefaultApi.md#devicesremovewifi) | **DELETE** /api/v1/device/{udid}/wifi | Remove wifi |
| [**DevicesResetAccessibility**](DefaultApi.md#devicesresetaccessibility) | **POST** /api/v1/device/{udid}/resetaccessibility | Reset accessibility |
| [**DevicesResetLocation**](DefaultApi.md#devicesresetlocation) | **POST** /api/v1/device/{udid}/resetlocation | Reset simulated location |
| [**DevicesScreenshot**](DefaultApi.md#devicesscreenshot) | **GET** /api/v1/device/{udid}/screenshot | Capture screenshot |
| [**DevicesSetAssistiveTouch**](DefaultApi.md#devicessetassistivetouch) | **PUT** /api/v1/device/{udid}/assistivetouch | Set AssistiveTouch |
| [**DevicesSetDevMode**](DefaultApi.md#devicessetdevmode) | **POST** /api/v1/device/{udid}/devmode | Set developer mode |
| [**DevicesSetHttpProxy**](DefaultApi.md#devicessethttpproxy) | **PUT** /api/v1/device/{udid}/httpproxy | Set HTTP proxy (supervised) |
| [**DevicesSetIconLayout**](DefaultApi.md#devicesseticonlayout) | **PUT** /api/v1/device/{udid}/icon-layout | Set icon layout |
| [**DevicesSetLanguage**](DefaultApi.md#devicessetlanguage) | **PUT** /api/v1/device/{udid}/lang | Set language |
| [**DevicesSetLocation**](DefaultApi.md#devicessetlocation) | **PUT** /api/v1/device/{udid}/setlocation | Set simulated location |
| [**DevicesSetPasteboard**](DefaultApi.md#devicessetpasteboard) | **PUT** /api/v1/device/{udid}/pasteboard | Set pasteboard |
| [**DevicesSetTimeFormat**](DefaultApi.md#devicessettimeformat) | **PUT** /api/v1/device/{udid}/timeformat | Set time format |
| [**DevicesSetWallpaper**](DefaultApi.md#devicessetwallpaper) | **PUT** /api/v1/device/{udid}/wallpaper | Set wallpaper (supervised) |
| [**DevicesSetWifi**](DefaultApi.md#devicessetwifi) | **PUT** /api/v1/device/{udid}/wifi | Provision wifi |
| [**DevicesShutdown**](DefaultApi.md#devicesshutdown) | **POST** /api/v1/device/{udid}/shutdown | Shut down device |
| [**DevicesStartForward**](DefaultApi.md#devicesstartforward) | **POST** /api/v1/device/{udid}/jobs/forward | Start port forward (job) |
| [**DevicesStartRunTest**](DefaultApi.md#devicesstartruntest) | **POST** /api/v1/device/{udid}/jobs/runtest | Start test run (job) |
| [**DevicesStartRunWda**](DefaultApi.md#devicesstartrunwda) | **POST** /api/v1/device/{udid}/jobs/runwda | Start WDA runner (job) |
| [**DevicesStopJob**](DefaultApi.md#devicesstopjob) | **DELETE** /api/v1/device/{udid}/jobs/{id} | Stop or delete job |
| [**DevicesStreamJobLogs**](DefaultApi.md#devicesstreamjoblogs) | **GET** /api/v1/device/{udid}/jobs/{id}/logs | Stream job logs (SSE) |
| [**DevicesStreamListen**](DefaultApi.md#devicesstreamlisten) | **GET** /api/v1/device/{udid}/listen | Stream device attach/detach (SSE) |
| [**DevicesStreamNotifications**](DefaultApi.md#devicesstreamnotifications) | **GET** /api/v1/device/{udid}/notifications | Stream app-state notifications (SSE) |
| [**DevicesStreamOsTrace**](DefaultApi.md#devicesstreamostrace) | **GET** /api/v1/device/{udid}/ostrace | Stream os_log trace (SSE) |
| [**DevicesStreamSyslog**](DefaultApi.md#devicesstreamsyslog) | **GET** /api/v1/device/{udid}/syslog | Stream syslog (SSE) |
| [**DevicesStreamSysmontap**](DefaultApi.md#devicesstreamsysmontap) | **GET** /api/v1/device/{udid}/sysmontap | Stream CPU usage (SSE) |
| [**DevicesUninstallApp**](DefaultApi.md#devicesuninstallapp) | **POST** /api/v1/device/{udid}/apps/uninstall | Uninstall app |
| [**DevicesUnmountImage**](DefaultApi.md#devicesunmountimage) | **DELETE** /api/v1/device/{udid}/image | Unmount developer image |
| [**DiagnosticsNetGetBatteryRegistry**](DefaultApi.md#diagnosticsnetgetbatteryregistry) | **GET** /api/v1/device/{udid}/battery/registry | Get battery IORegistry |
| [**DiagnosticsNetGetDeviceIp**](DefaultApi.md#diagnosticsnetgetdeviceip) | **GET** /api/v1/device/{udid}/ip | Get device IP / network info |
| [**DiagnosticsNetGetDiskSpace**](DefaultApi.md#diagnosticsnetgetdiskspace) | **GET** /api/v1/device/{udid}/diskspace | Get disk space info |
| [**DiagnosticsNetGetRsdServices**](DefaultApi.md#diagnosticsnetgetrsdservices) | **GET** /api/v1/device/{udid}/rsd | Get RSD service list |
| [**FsyncFsyncLs**](DefaultApi.md#fsyncfsyncls) | **GET** /api/v1/device/{udid}/fsync/ls | List a directory over AFC |
| [**FsyncFsyncMkdir**](DefaultApi.md#fsyncfsyncmkdir) | **POST** /api/v1/device/{udid}/fsync/mkdir | Create a directory over AFC |
| [**FsyncFsyncPull**](DefaultApi.md#fsyncfsyncpull) | **GET** /api/v1/device/{udid}/fsync/pull | Download a file over AFC |
| [**FsyncFsyncPush**](DefaultApi.md#fsyncfsyncpush) | **POST** /api/v1/device/{udid}/fsync/push | Upload a file over AFC |
| [**FsyncFsyncRm**](DefaultApi.md#fsyncfsyncrm) | **DELETE** /api/v1/device/{udid}/fsync/rm | Remove a file or directory over AFC |
| [**FsyncFsyncTree**](DefaultApi.md#fsyncfsynctree) | **GET** /api/v1/device/{udid}/fsync/tree | Recursively list a directory over AFC |
| [**FsyncGetCloudConfig**](DefaultApi.md#fsyncgetcloudconfig) | **GET** /api/v1/device/{udid}/cloudconfig | Get device cloud configuration |
| [**GetPrepareSkipOptions**](DefaultApi.md#getprepareskipoptions) | **GET** /api/v1/prepare/skip-options | List setup skip options |
| [**ListDevices**](DefaultApi.md#listdevices) | **GET** /api/v1/list | List devices |
| [**ListTunnels**](DefaultApi.md#listtunnels) | **GET** /api/v1/tunnels | List tunnels |
| [**PrepareCreateCert**](DefaultApi.md#preparecreatecert) | **POST** /api/v1/prepare/create-cert | Generate a supervision certificate |
| [**PreparePrepareDevice**](DefaultApi.md#preparepreparedevice) | **POST** /api/v1/device/{udid}/prepare | Prepare (and optionally supervise) a device |
| [**RefreshTunnel**](DefaultApi.md#refreshtunnel) | **POST** /api/v1/tunnels/{udid}/refresh | Refresh tunnel |
| [**ShutdownTunnelAgent**](DefaultApi.md#shutdowntunnelagent) | **POST** /api/v1/tunnel-agent/shutdown | Shut down tunnel agent |
| [**SignApp**](DefaultApi.md#signapp) | **POST** /api/v1/sign/app | Resign an app/IPA |
| [**SignCertificate**](DefaultApi.md#signcertificate) | **POST** /api/v1/sign/certificate | Create a signing certificate |
| [**SignProvision**](DefaultApi.md#signprovision) | **POST** /api/v1/sign/provision | Create a provisioning profile + P12 |
| [**StopTunnel**](DefaultApi.md#stoptunnel) | **DELETE** /api/v1/tunnels/{udid} | Stop tunnel |
| [**StreamsPcap**](DefaultApi.md#streamspcap) | **GET** /api/v1/device/{udid}/pcap | Stream a live pcap capture (binary) |
| [**StreamsScreenshotStream**](DefaultApi.md#streamsscreenshotstream) | **GET** /api/v1/device/{udid}/screenshot/stream | Stream screenshots as MJPEG (binary) |
| [**StreamsUiStream**](DefaultApi.md#streamsuistream) | **GET** /api/v1/device/{udid}/ui/stream | Stream UI video (binary) |
| [**UIUiApi**](DefaultApi.md#uiuiapi) | **POST** /api/v1/device/{udid}/ui/api | Raw backend passthrough |
| [**UIUiAppForeground**](DefaultApi.md#uiuiappforeground) | **POST** /api/v1/device/{udid}/ui/app/foreground | Foreground app (UI backend) |
| [**UIUiAppLaunch**](DefaultApi.md#uiuiapplaunch) | **POST** /api/v1/device/{udid}/ui/app/launch | Launch app (UI backend) |
| [**UIUiAppTerminate**](DefaultApi.md#uiuiappterminate) | **POST** /api/v1/device/{udid}/ui/app/terminate | Terminate app (UI backend) |
| [**UIUiButton**](DefaultApi.md#uiuibutton) | **POST** /api/v1/device/{udid}/ui/button | Press hardware button |
| [**UIUiGetOrientation**](DefaultApi.md#uiuigetorientation) | **GET** /api/v1/device/{udid}/ui/orientation | Get orientation |
| [**UIUiLongPress**](DefaultApi.md#uiuilongpress) | **POST** /api/v1/device/{udid}/ui/longpress | Long press |
| [**UIUiScreenshot**](DefaultApi.md#uiuiscreenshot) | **GET** /api/v1/device/{udid}/ui/screenshot | UI screenshot (PNG) |
| [**UIUiSetOrientation**](DefaultApi.md#uiuisetorientation) | **PUT** /api/v1/device/{udid}/ui/orientation | Set orientation |
| [**UIUiSource**](DefaultApi.md#uiuisource) | **GET** /api/v1/device/{udid}/ui/source | UI source hierarchy |
| [**UIUiStatus**](DefaultApi.md#uiuistatus) | **GET** /api/v1/device/{udid}/ui/status | UI backend status |
| [**UIUiSwipe**](DefaultApi.md#uiuiswipe) | **POST** /api/v1/device/{udid}/ui/swipe | Swipe |
| [**UIUiTap**](DefaultApi.md#uiuitap) | **POST** /api/v1/device/{udid}/ui/tap | Tap |
| [**UIUiType**](DefaultApi.md#uiuitype) | **POST** /api/v1/device/{udid}/ui/type | Type text |
| [**UIUiWindowSize**](DefaultApi.md#uiuiwindowsize) | **GET** /api/v1/device/{udid}/ui/size | UI window size |
| [**WebInspectorWebInspectorEval**](DefaultApi.md#webinspectorwebinspectoreval) | **POST** /api/v1/device/{udid}/webinspector/eval | Evaluate JavaScript in a page |
| [**WebInspectorWebInspectorLaunch**](DefaultApi.md#webinspectorwebinspectorlaunch) | **POST** /api/v1/device/{udid}/webinspector/launch | Open a URL in a new inspectable page |
| [**WebInspectorWebInspectorPages**](DefaultApi.md#webinspectorwebinspectorpages) | **GET** /api/v1/device/{udid}/webinspector/pages | List inspectable pages |

<a id="accessibilitygetaxsnapshot"></a>
# **AccessibilityGetAxSnapshot**
> Object AccessibilityGetAxSnapshot (string udid)

Get accessibility element snapshot

Get a snapshot of the currently focused accessibility element (CLI: `ios ax`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class AccessibilityGetAxSnapshotExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Get accessibility element snapshot
                Object result = apiInstance.AccessibilityGetAxSnapshot(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.AccessibilityGetAxSnapshot: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the AccessibilityGetAxSnapshotWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get accessibility element snapshot
    ApiResponse<Object> response = apiInstance.AccessibilityGetAxSnapshotWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.AccessibilityGetAxSnapshotWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="accessibilitygetvoiceover"></a>
# **AccessibilityGetVoiceOver**
> VoiceOverState AccessibilityGetVoiceOver (string udid)

Get VoiceOver state

Get VoiceOver enabled state (CLI: `ios voiceover get`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class AccessibilityGetVoiceOverExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Get VoiceOver state
                VoiceOverState result = apiInstance.AccessibilityGetVoiceOver(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.AccessibilityGetVoiceOver: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the AccessibilityGetVoiceOverWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get VoiceOver state
    ApiResponse<VoiceOverState> response = apiInstance.AccessibilityGetVoiceOverWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.AccessibilityGetVoiceOverWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

[**VoiceOverState**](VoiceOverState.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="accessibilitygetzoomtouch"></a>
# **AccessibilityGetZoomTouch**
> ZoomTouchState AccessibilityGetZoomTouch (string udid)

Get ZoomTouch state

Get ZoomTouch enabled state (CLI: `ios zoomtouch get`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class AccessibilityGetZoomTouchExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Get ZoomTouch state
                ZoomTouchState result = apiInstance.AccessibilityGetZoomTouch(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.AccessibilityGetZoomTouch: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the AccessibilityGetZoomTouchWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get ZoomTouch state
    ApiResponse<ZoomTouchState> response = apiInstance.AccessibilityGetZoomTouchWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.AccessibilityGetZoomTouchWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

[**ZoomTouchState**](ZoomTouchState.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="accessibilityrunaxaudit"></a>
# **AccessibilityRunAxAudit**
> List&lt;Object&gt; AccessibilityRunAxAudit (string udid, int? timeout = null)

Run accessibility audit

Run the accessibility audit against the focused app and return the issues found (CLI: `ios ax audit`). Bounded by `timeout` (seconds, default 60).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class AccessibilityRunAxAuditExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var timeout = 56;  // int? | Audit timeout in seconds (default 60). (optional) 

            try
            {
                // Run accessibility audit
                List<Object> result = apiInstance.AccessibilityRunAxAudit(udid, timeout);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.AccessibilityRunAxAudit: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the AccessibilityRunAxAuditWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Run accessibility audit
    ApiResponse<List<Object>> response = apiInstance.AccessibilityRunAxAuditWithHttpInfo(udid, timeout);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.AccessibilityRunAxAuditWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **timeout** | **int?** | Audit timeout in seconds (default 60). | [optional]  |

### Return type

**List<Object>**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="accessibilitysetlocationgpx"></a>
# **AccessibilitySetLocationGpx**
> GenericResponse AccessibilitySetLocationGpx (string udid, Object gpx)

Simulate location from a GPX file

Simulate live location tracking from an uploaded GPX file (CLI: `ios setlocationgpx`). Send multipart/form-data with a `gpx` file.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class AccessibilitySetLocationGpxExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var gpx = new Object(); // Object | 

            try
            {
                // Simulate location from a GPX file
                GenericResponse result = apiInstance.AccessibilitySetLocationGpx(udid, gpx);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.AccessibilitySetLocationGpx: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the AccessibilitySetLocationGpxWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Simulate location from a GPX file
    ApiResponse<GenericResponse> response = apiInstance.AccessibilitySetLocationGpxWithHttpInfo(udid, gpx);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.AccessibilitySetLocationGpxWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **gpx** | [**Object**](Object.md) |  |  |

### Return type

[**GenericResponse**](GenericResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: multipart/form-data
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="accessibilitysetvoiceover"></a>
# **AccessibilitySetVoiceOver**
> VoiceOverState AccessibilitySetVoiceOver (string udid, bool? enabled = null, AXEnabledRequest? aXEnabledRequest = null)

Set VoiceOver state

Enable/disable VoiceOver (CLI: `ios voiceover enable|disable`). The desired state comes from the JSON body or the `enabled` query param.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class AccessibilitySetVoiceOverExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var enabled = true;  // bool? | Desired state (alternative to the request body). (optional) 
            var aXEnabledRequest = new AXEnabledRequest?(); // AXEnabledRequest? |  (optional) 

            try
            {
                // Set VoiceOver state
                VoiceOverState result = apiInstance.AccessibilitySetVoiceOver(udid, enabled, aXEnabledRequest);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.AccessibilitySetVoiceOver: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the AccessibilitySetVoiceOverWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Set VoiceOver state
    ApiResponse<VoiceOverState> response = apiInstance.AccessibilitySetVoiceOverWithHttpInfo(udid, enabled, aXEnabledRequest);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.AccessibilitySetVoiceOverWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **enabled** | **bool?** | Desired state (alternative to the request body). | [optional]  |
| **aXEnabledRequest** | [**AXEnabledRequest?**](AXEnabledRequest?.md) |  | [optional]  |

### Return type

[**VoiceOverState**](VoiceOverState.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="accessibilitysetzoomtouch"></a>
# **AccessibilitySetZoomTouch**
> ZoomTouchState AccessibilitySetZoomTouch (string udid, bool? enabled = null, AXEnabledRequest? aXEnabledRequest = null)

Set ZoomTouch state

Enable/disable ZoomTouch (CLI: `ios zoomtouch enable|disable`). The desired state comes from the JSON body or the `enabled` query param.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class AccessibilitySetZoomTouchExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var enabled = true;  // bool? | Desired state (alternative to the request body). (optional) 
            var aXEnabledRequest = new AXEnabledRequest?(); // AXEnabledRequest? |  (optional) 

            try
            {
                // Set ZoomTouch state
                ZoomTouchState result = apiInstance.AccessibilitySetZoomTouch(udid, enabled, aXEnabledRequest);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.AccessibilitySetZoomTouch: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the AccessibilitySetZoomTouchWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Set ZoomTouch state
    ApiResponse<ZoomTouchState> response = apiInstance.AccessibilitySetZoomTouchWithHttpInfo(udid, enabled, aXEnabledRequest);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.AccessibilitySetZoomTouchWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **enabled** | **bool?** | Desired state (alternative to the request body). | [optional]  |
| **aXEnabledRequest** | [**AXEnabledRequest?**](AXEnabledRequest?.md) |  | [optional]  |

### Return type

[**ZoomTouchState**](ZoomTouchState.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesactivate"></a>
# **DevicesActivate**
> GenericResponse DevicesActivate (string udid)

Activate device

Activate the device (complete Setup Assistant / activation).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesActivateExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Activate device
                GenericResponse result = apiInstance.DevicesActivate(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesActivate: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesActivateWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Activate device
    ApiResponse<GenericResponse> response = apiInstance.DevicesActivateWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesActivateWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

[**GenericResponse**](GenericResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesaddprofile"></a>
# **DevicesAddProfile**
> GenericResponse DevicesAddProfile (string udid, Object profile, Object? p12 = null, string? password = null)

Install profile

Install a configuration profile (CLI: `ios profile add`). Send the profile as the raw request body, or as multipart with a `profile` file plus an optional `p12` supervisor identity and `password` for a supervised install.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesAddProfileExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var profile = new Object(); // Object | 
            var p12 = new Object?(); // Object? |  (optional) 
            var password = "password_example";  // string? | Passphrase for the `.p12` identity. (optional) 

            try
            {
                // Install profile
                GenericResponse result = apiInstance.DevicesAddProfile(udid, profile, p12, password);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesAddProfile: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesAddProfileWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Install profile
    ApiResponse<GenericResponse> response = apiInstance.DevicesAddProfileWithHttpInfo(udid, profile, p12, password);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesAddProfileWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **profile** | [**Object**](Object.md) |  |  |
| **p12** | [**Object?**](Object?.md) |  | [optional]  |
| **password** | **string?** | Passphrase for the &#x60;.p12&#x60; identity. | [optional]  |

### Return type

[**GenericResponse**](GenericResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: multipart/form-data
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicescreatewdasession"></a>
# **DevicesCreateWdaSession**
> WdaSession DevicesCreateWdaSession (string udid, WdaConfig wdaConfig)

Start WDA session

Start a WebDriverAgent (XCUITest) session.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesCreateWdaSessionExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var wdaConfig = new WdaConfig(); // WdaConfig | 

            try
            {
                // Start WDA session
                WdaSession result = apiInstance.DevicesCreateWdaSession(udid, wdaConfig);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesCreateWdaSession: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesCreateWdaSessionWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Start WDA session
    ApiResponse<WdaSession> response = apiInstance.DevicesCreateWdaSessionWithHttpInfo(udid, wdaConfig);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesCreateWdaSessionWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **wdaConfig** | [**WdaConfig**](WdaConfig.md) |  |  |

### Return type

[**WdaSession**](WdaSession.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesdeletewdasession"></a>
# **DevicesDeleteWdaSession**
> WdaSession DevicesDeleteWdaSession (string udid, string sessionId)

Stop WDA session

Stop a running WebDriverAgent session.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesDeleteWdaSessionExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var sessionId = "sessionId_example";  // string | The WDA session id.

            try
            {
                // Stop WDA session
                WdaSession result = apiInstance.DevicesDeleteWdaSession(udid, sessionId);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesDeleteWdaSession: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesDeleteWdaSessionWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Stop WDA session
    ApiResponse<WdaSession> response = apiInstance.DevicesDeleteWdaSessionWithHttpInfo(udid, sessionId);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesDeleteWdaSessionWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **sessionId** | **string** | The WDA session id. |  |

### Return type

[**WdaSession**](WdaSession.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — WDA session id not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesdisablecondition"></a>
# **DevicesDisableCondition**
> GenericResponse DevicesDisableCondition (string udid)

Disable condition

Disable the currently active condition inducer profile.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesDisableConditionExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Disable condition
                GenericResponse result = apiInstance.DevicesDisableCondition(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesDisableCondition: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesDisableConditionWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Disable condition
    ApiResponse<GenericResponse> response = apiInstance.DevicesDisableConditionWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesDisableConditionWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

[**GenericResponse**](GenericResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesenablecondition"></a>
# **DevicesEnableCondition**
> GenericResponse DevicesEnableCondition (string udid, string profileTypeID, string profileID)

Enable condition

Enable a condition inducer profile.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesEnableConditionExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var profileTypeID = "profileTypeID_example";  // string | Identifier of the condition profile type.
            var profileID = "profileID_example";  // string | Identifier of the specific profile to activate.

            try
            {
                // Enable condition
                GenericResponse result = apiInstance.DevicesEnableCondition(udid, profileTypeID, profileID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesEnableCondition: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesEnableConditionWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Enable condition
    ApiResponse<GenericResponse> response = apiInstance.DevicesEnableConditionWithHttpInfo(udid, profileTypeID, profileID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesEnableConditionWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **profileTypeID** | **string** | Identifier of the condition profile type. |  |
| **profileID** | **string** | Identifier of the specific profile to activate. |  |

### Return type

[**GenericResponse**](GenericResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="deviceserase"></a>
# **DevicesErase**
> GenericResponse DevicesErase (string udid, bool confirm)

Erase device

Erase all content and settings (CLI: `ios erase`). Destructive: requires `confirm=true`.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesEraseExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var confirm = true;  // bool | Must be `true` to proceed with the destructive erase.

            try
            {
                // Erase device
                GenericResponse result = apiInstance.DevicesErase(udid, confirm);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesErase: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesEraseWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Erase device
    ApiResponse<GenericResponse> response = apiInstance.DevicesEraseWithHttpInfo(udid, confirm);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesEraseWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **confirm** | **bool** | Must be &#x60;true&#x60; to proceed with the destructive erase. |  |

### Return type

[**GenericResponse**](GenericResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesgetassistivetouch"></a>
# **DevicesGetAssistiveTouch**
> AssistiveTouchState DevicesGetAssistiveTouch (string udid)

Get AssistiveTouch

Get AssistiveTouch state (CLI: `ios assistivetouch get`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesGetAssistiveTouchExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Get AssistiveTouch
                AssistiveTouchState result = apiInstance.DevicesGetAssistiveTouch(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesGetAssistiveTouch: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesGetAssistiveTouchWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get AssistiveTouch
    ApiResponse<AssistiveTouchState> response = apiInstance.DevicesGetAssistiveTouchWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesGetAssistiveTouchWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

[**AssistiveTouchState**](AssistiveTouchState.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesgetbattery"></a>
# **DevicesGetBattery**
> BatteryInfo DevicesGetBattery (string udid)

Get battery info

Get battery diagnostics (CLI: `ios batterycheck`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesGetBatteryExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Get battery info
                BatteryInfo result = apiInstance.DevicesGetBattery(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesGetBattery: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesGetBatteryWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get battery info
    ApiResponse<BatteryInfo> response = apiInstance.DevicesGetBatteryWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesGetBatteryWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

[**BatteryInfo**](BatteryInfo.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesgetdevmode"></a>
# **DevicesGetDevMode**
> DevModeState DevicesGetDevMode (string udid)

Get developer mode

Get developer mode state (CLI: `ios devmode get`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesGetDevModeExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Get developer mode
                DevModeState result = apiInstance.DevicesGetDevMode(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesGetDevMode: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesGetDevModeWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get developer mode
    ApiResponse<DevModeState> response = apiInstance.DevicesGetDevModeWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesGetDevModeWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

[**DevModeState**](DevModeState.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesgetdevicedate"></a>
# **DevicesGetDeviceDate**
> DeviceDate DevicesGetDeviceDate (string udid)

Get device date

Get the device clock (CLI: `ios date`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesGetDeviceDateExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Get device date
                DeviceDate result = apiInstance.DevicesGetDeviceDate(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesGetDeviceDate: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesGetDeviceDateWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get device date
    ApiResponse<DeviceDate> response = apiInstance.DevicesGetDeviceDateWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesGetDeviceDateWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

[**DeviceDate**](DeviceDate.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesgetdevicename"></a>
# **DevicesGetDeviceName**
> DeviceName DevicesGetDeviceName (string udid)

Get device name

Get the device name (CLI: `ios devicename`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesGetDeviceNameExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Get device name
                DeviceName result = apiInstance.DevicesGetDeviceName(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesGetDeviceName: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesGetDeviceNameWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get device name
    ApiResponse<DeviceName> response = apiInstance.DevicesGetDeviceNameWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesGetDeviceNameWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

[**DeviceName**](DeviceName.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesgetdiagnostics"></a>
# **DevicesGetDiagnostics**
> Object DevicesGetDiagnostics (string udid)

List diagnostics

List all IORegistry/diagnostic values (CLI: `ios diagnostics list`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesGetDiagnosticsExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // List diagnostics
                Object result = apiInstance.DevicesGetDiagnostics(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesGetDiagnostics: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesGetDiagnosticsWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // List diagnostics
    ApiResponse<Object> response = apiInstance.DevicesGetDiagnosticsWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesGetDiagnosticsWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesgeticonlayout"></a>
# **DevicesGetIconLayout**
> Object DevicesGetIconLayout (string udid)

Get icon layout

Get the SpringBoard icon layout (CLI: `ios get-icon-layout`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesGetIconLayoutExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Get icon layout
                Object result = apiInstance.DevicesGetIconLayout(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesGetIconLayout: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesGetIconLayoutWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get icon layout
    ApiResponse<Object> response = apiInstance.DevicesGetIconLayoutWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesGetIconLayoutWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesgetinfo"></a>
# **DevicesGetInfo**
> Object DevicesGetInfo (string udid)

Get device info

Get lockdown values plus `instruments:*` keys for the device. Returns an open dictionary of heterogeneous values.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesGetInfoExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Get device info
                Object result = apiInstance.DevicesGetInfo(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesGetInfo: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesGetInfoWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get device info
    ApiResponse<Object> response = apiInstance.DevicesGetInfoWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesGetInfoWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesgetjob"></a>
# **DevicesGetJob**
> Job DevicesGetJob (string udid, string id)

Get job

Get a job's status. Returns `404` for an unknown job on this device.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesGetJobExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var id = "id_example";  // string | The job id.

            try
            {
                // Get job
                Job result = apiInstance.DevicesGetJob(udid, id);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesGetJob: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesGetJobWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get job
    ApiResponse<Job> response = apiInstance.DevicesGetJobWithHttpInfo(udid, id);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesGetJobWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **id** | **string** | The job id. |  |

### Return type

[**Job**](Job.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — the requested resource (e.g. a job) was not found for this device. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesgetlanguage"></a>
# **DevicesGetLanguage**
> LanguageConfiguration DevicesGetLanguage (string udid)

Get language

Get the device language/locale configuration (CLI: `ios lang`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesGetLanguageExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Get language
                LanguageConfiguration result = apiInstance.DevicesGetLanguage(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesGetLanguage: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesGetLanguageWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get language
    ApiResponse<LanguageConfiguration> response = apiInstance.DevicesGetLanguageWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesGetLanguageWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

[**LanguageConfiguration**](LanguageConfiguration.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesgetlockdownvalues"></a>
# **DevicesGetLockdownValues**
> Object DevicesGetLockdownValues (string udid, string? domain = null)

Get lockdown values

Get lockdown values (CLI: `ios lockdown get`). Without `domain` the full set is returned; with `domain` the values are scoped to that lockdown domain.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesGetLockdownValuesExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var domain = "domain_example";  // string? | Optional lockdown domain to scope the returned values. (optional) 

            try
            {
                // Get lockdown values
                Object result = apiInstance.DevicesGetLockdownValues(udid, domain);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesGetLockdownValues: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesGetLockdownValuesWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get lockdown values
    ApiResponse<Object> response = apiInstance.DevicesGetLockdownValuesWithHttpInfo(udid, domain);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesGetLockdownValuesWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **domain** | **string?** | Optional lockdown domain to scope the returned values. | [optional]  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesgetmobilegestalt"></a>
# **DevicesGetMobileGestalt**
> Object DevicesGetMobileGestalt (string udid, List<string> key)

Query MobileGestalt

Query one or more MobileGestalt keys (CLI: `ios mobilegestalt <key>...`). Pass repeated `key` query params.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesGetMobileGestaltExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var key = new List<string>(); // List<string> | One or more MobileGestalt keys to query.

            try
            {
                // Query MobileGestalt
                Object result = apiInstance.DevicesGetMobileGestalt(udid, key);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesGetMobileGestalt: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesGetMobileGestaltWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Query MobileGestalt
    ApiResponse<Object> response = apiInstance.DevicesGetMobileGestaltWithHttpInfo(udid, key);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesGetMobileGestaltWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **key** | [**List&lt;string&gt;**](string.md) | One or more MobileGestalt keys to query. |  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesgetpasteboard"></a>
# **DevicesGetPasteboard**
> PasteboardContent DevicesGetPasteboard (string udid)

Get pasteboard

Get the pasteboard (clipboard) text (CLI: `ios pasteboard get`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesGetPasteboardExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Get pasteboard
                PasteboardContent result = apiInstance.DevicesGetPasteboard(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesGetPasteboard: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesGetPasteboardWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get pasteboard
    ApiResponse<PasteboardContent> response = apiInstance.DevicesGetPasteboardWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesGetPasteboardWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

[**PasteboardContent**](PasteboardContent.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesgetprocesses"></a>
# **DevicesGetProcesses**
> List&lt;ProcessInfo&gt; DevicesGetProcesses (string udid, bool? apps = null)

List processes

List running processes (CLI: `ios ps [- -apps]`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesGetProcessesExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var apps = true;  // bool? | Only return application processes. (optional) 

            try
            {
                // List processes
                List<ProcessInfo> result = apiInstance.DevicesGetProcesses(udid, apps);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesGetProcesses: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesGetProcessesWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // List processes
    ApiResponse<List<ProcessInfo>> response = apiInstance.DevicesGetProcessesWithHttpInfo(udid, apps);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesGetProcessesWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **apps** | **bool?** | Only return application processes. | [optional]  |

### Return type

[**List&lt;ProcessInfo&gt;**](ProcessInfo.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesgetprofiles"></a>
# **DevicesGetProfiles**
> Object DevicesGetProfiles (string udid)

List configuration profiles

List installed configuration profiles. Returns an open dictionary.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesGetProfilesExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // List configuration profiles
                Object result = apiInstance.DevicesGetProfiles(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesGetProfiles: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesGetProfilesWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // List configuration profiles
    ApiResponse<Object> response = apiInstance.DevicesGetProfilesWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesGetProfilesWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesgettimeformat"></a>
# **DevicesGetTimeFormat**
> TimeFormatState DevicesGetTimeFormat (string udid)

Get time format

Get the 24-hour clock state (CLI: `ios timeformat get`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesGetTimeFormatExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Get time format
                TimeFormatState result = apiInstance.DevicesGetTimeFormat(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesGetTimeFormat: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesGetTimeFormatWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get time format
    ApiResponse<TimeFormatState> response = apiInstance.DevicesGetTimeFormatWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesGetTimeFormatWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

[**TimeFormatState**](TimeFormatState.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesgetwallpaper"></a>
# **DevicesGetWallpaper**
> Object DevicesGetWallpaper (string udid)

Get wallpaper

Get the home-screen wallpaper as PNG (CLI: `ios get-wallpaper`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesGetWallpaperExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Get wallpaper
                Object result = apiInstance.DevicesGetWallpaper(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesGetWallpaper: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesGetWallpaperWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get wallpaper
    ApiResponse<Object> response = apiInstance.DevicesGetWallpaperWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesGetWallpaperWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: image/png, application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesgetwdasession"></a>
# **DevicesGetWdaSession**
> WdaSession DevicesGetWdaSession (string udid, string sessionId)

Get WDA session

Get a running WebDriverAgent session. Returns `404` for an unknown session.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesGetWdaSessionExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var sessionId = "sessionId_example";  // string | The WDA session id.

            try
            {
                // Get WDA session
                WdaSession result = apiInstance.DevicesGetWdaSession(udid, sessionId);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesGetWdaSession: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesGetWdaSessionWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get WDA session
    ApiResponse<WdaSession> response = apiInstance.DevicesGetWdaSessionWithHttpInfo(udid, sessionId);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesGetWdaSessionWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **sessionId** | **string** | The WDA session id. |  |

### Return type

[**WdaSession**](WdaSession.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — WDA session id not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesinstallapp"></a>
# **DevicesInstallApp**
> GenericResponse DevicesInstallApp (string udid, Object file)

Install app

Install an application from an uploaded `.ipa`/`.app` archive. The multipart `file` part must be 1 byte–200 MB.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesInstallAppExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var file = new Object(); // Object | 

            try
            {
                // Install app
                GenericResponse result = apiInstance.DevicesInstallApp(udid, file);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesInstallApp: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesInstallAppWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Install app
    ApiResponse<GenericResponse> response = apiInstance.DevicesInstallAppWithHttpInfo(udid, file);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesInstallAppWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **file** | [**Object**](Object.md) |  |  |

### Return type

[**GenericResponse**](GenericResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: multipart/form-data
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="deviceskillapp"></a>
# **DevicesKillApp**
> GenericResponse DevicesKillApp (string udid, string bundleID)

Kill app

Kill a running application by bundle id.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesKillAppExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var bundleID = "bundleID_example";  // string | Bundle id of the app to kill.

            try
            {
                // Kill app
                GenericResponse result = apiInstance.DevicesKillApp(udid, bundleID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesKillApp: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesKillAppWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Kill app
    ApiResponse<GenericResponse> response = apiInstance.DevicesKillAppWithHttpInfo(udid, bundleID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesKillAppWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **bundleID** | **string** | Bundle id of the app to kill. |  |

### Return type

[**GenericResponse**](GenericResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="deviceslaunchapp"></a>
# **DevicesLaunchApp**
> GenericResponse DevicesLaunchApp (string udid, string bundleID)

Launch app

Launch an application by bundle id.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesLaunchAppExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var bundleID = "bundleID_example";  // string | Bundle id of the app to launch.

            try
            {
                // Launch app
                GenericResponse result = apiInstance.DevicesLaunchApp(udid, bundleID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesLaunchApp: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesLaunchAppWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Launch app
    ApiResponse<GenericResponse> response = apiInstance.DevicesLaunchAppWithHttpInfo(udid, bundleID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesLaunchAppWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **bundleID** | **string** | Bundle id of the app to launch. |  |

### Return type

[**GenericResponse**](GenericResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="deviceslistapps"></a>
# **DevicesListApps**
> List&lt;AppInfo&gt; DevicesListApps (string udid)

List apps

List installed applications. Each entry is an open Info.plist map.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesListAppsExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // List apps
                List<AppInfo> result = apiInstance.DevicesListApps(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesListApps: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesListAppsWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // List apps
    ApiResponse<List<AppInfo>> response = apiInstance.DevicesListAppsWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesListAppsWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

[**List&lt;AppInfo&gt;**](AppInfo.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="deviceslistconditions"></a>
# **DevicesListConditions**
> List&lt;ProfileType&gt; DevicesListConditions (string udid)

List conditions

List available condition inducer profile types.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesListConditionsExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // List conditions
                List<ProfileType> result = apiInstance.DevicesListConditions(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesListConditions: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesListConditionsWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // List conditions
    ApiResponse<List<ProfileType>> response = apiInstance.DevicesListConditionsWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesListConditionsWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

[**List&lt;ProfileType&gt;**](ProfileType.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="deviceslistcrashes"></a>
# **DevicesListCrashes**
> CrashListing DevicesListCrashes (string udid, string? pattern = null)

List crash reports

List crash reports (CLI: `ios crash ls`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesListCrashesExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var pattern = "pattern_example";  // string? | Optional glob pattern to filter reports. (optional) 

            try
            {
                // List crash reports
                CrashListing result = apiInstance.DevicesListCrashes(udid, pattern);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesListCrashes: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesListCrashesWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // List crash reports
    ApiResponse<CrashListing> response = apiInstance.DevicesListCrashesWithHttpInfo(udid, pattern);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesListCrashesWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **pattern** | **string?** | Optional glob pattern to filter reports. | [optional]  |

### Return type

[**CrashListing**](CrashListing.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="deviceslistfiles"></a>
# **DevicesListFiles**
> FileListing DevicesListFiles (string udid, FileDomain domain, string? identifier = null, string? path = null)

List files

List a device directory (CLI: `ios file ls`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesListFilesExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var domain = new FileDomain(); // FileDomain | File service domain: `app`, `app-group`, `crash` or `temp`.
            var identifier = "identifier_example";  // string? | Bundle/group id for the `app`/`app-group` domains. (optional) 
            var path = "path_example";  // string? | Directory path to list (defaults to `.`). (optional) 

            try
            {
                // List files
                FileListing result = apiInstance.DevicesListFiles(udid, domain, identifier, path);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesListFiles: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesListFilesWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // List files
    ApiResponse<FileListing> response = apiInstance.DevicesListFilesWithHttpInfo(udid, domain, identifier, path);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesListFilesWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **domain** | [**FileDomain**](FileDomain.md) | File service domain: &#x60;app&#x60;, &#x60;app-group&#x60;, &#x60;crash&#x60; or &#x60;temp&#x60;. |  |
| **identifier** | **string?** | Bundle/group id for the &#x60;app&#x60;/&#x60;app-group&#x60; domains. | [optional]  |
| **path** | **string?** | Directory path to list (defaults to &#x60;.&#x60;). | [optional]  |

### Return type

[**FileListing**](FileListing.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="deviceslistimages"></a>
# **DevicesListImages**
> List&lt;string&gt; DevicesListImages (string udid)

List mounted developer images

List the hex signatures of Developer Disk Images mounted on the device.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesListImagesExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // List mounted developer images
                List<string> result = apiInstance.DevicesListImages(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesListImages: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesListImagesWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // List mounted developer images
    ApiResponse<List<string>> response = apiInstance.DevicesListImagesWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesListImagesWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

**List<string>**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="deviceslistjobs"></a>
# **DevicesListJobs**
> List&lt;Job&gt; DevicesListJobs (string udid)

List jobs

List jobs for a device.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesListJobsExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // List jobs
                List<Job> result = apiInstance.DevicesListJobs(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesListJobs: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesListJobsWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // List jobs
    ApiResponse<List<Job>> response = apiInstance.DevicesListJobsWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesListJobsWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

[**List&lt;Job&gt;**](Job.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="deviceslistmountedimages"></a>
# **DevicesListMountedImages**
> MountedImages DevicesListMountedImages (string udid)

List mounted images

List mounted developer image signatures (CLI: `ios image list`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesListMountedImagesExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // List mounted images
                MountedImages result = apiInstance.DevicesListMountedImages(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesListMountedImages: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesListMountedImagesWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // List mounted images
    ApiResponse<MountedImages> response = apiInstance.DevicesListMountedImagesWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesListMountedImagesWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

[**MountedImages**](MountedImages.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesmdmclearpasscode"></a>
# **DevicesMdmClearPasscode**
> StatusOk DevicesMdmClearPasscode (string udid, Object p12, string token, string? password = null)

Clear passcode (supervised)

Clear the device passcode (CLI: `ios mdm clear-passcode`). Requires the base64 unlock token as an additional `token` form field.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesMdmClearPasscodeExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var p12 = new Object(); // Object | 
            var token = "token_example";  // string | Base64-encoded escrow unlock token.
            var password = "password_example";  // string? | Passphrase for the `.p12` identity. (optional) 

            try
            {
                // Clear passcode (supervised)
                StatusOk result = apiInstance.DevicesMdmClearPasscode(udid, p12, token, password);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesMdmClearPasscode: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesMdmClearPasscodeWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Clear passcode (supervised)
    ApiResponse<StatusOk> response = apiInstance.DevicesMdmClearPasscodeWithHttpInfo(udid, p12, token, password);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesMdmClearPasscodeWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **p12** | [**Object**](Object.md) |  |  |
| **token** | **string** | Base64-encoded escrow unlock token. |  |
| **password** | **string?** | Passphrase for the &#x60;.p12&#x60; identity. | [optional]  |

### Return type

[**StatusOk**](StatusOk.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: multipart/form-data
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesmdmclearscreentimepassword"></a>
# **DevicesMdmClearScreenTimePassword**
> StatusOk DevicesMdmClearScreenTimePassword (string udid, Object p12, string? password = null)

Clear Screen Time password (supervised)

Clear the Screen Time password (CLI: `ios mdm clear-screen-time-password`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesMdmClearScreenTimePasswordExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var p12 = new Object(); // Object | 
            var password = "password_example";  // string? | Passphrase for the `.p12` identity. (optional) 

            try
            {
                // Clear Screen Time password (supervised)
                StatusOk result = apiInstance.DevicesMdmClearScreenTimePassword(udid, p12, password);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesMdmClearScreenTimePassword: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesMdmClearScreenTimePasswordWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Clear Screen Time password (supervised)
    ApiResponse<StatusOk> response = apiInstance.DevicesMdmClearScreenTimePasswordWithHttpInfo(udid, p12, password);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesMdmClearScreenTimePasswordWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **p12** | [**Object**](Object.md) |  |  |
| **password** | **string?** | Passphrase for the &#x60;.p12&#x60; identity. | [optional]  |

### Return type

[**StatusOk**](StatusOk.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: multipart/form-data
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesmdmfetchunlocktoken"></a>
# **DevicesMdmFetchUnlockToken**
> UnlockToken DevicesMdmFetchUnlockToken (string udid, Object p12, string? password = null)

Fetch unlock token (supervised)

Fetch the escrow unlock token, base64-encoded (CLI: `ios mdm fetch-unlock-token`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesMdmFetchUnlockTokenExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var p12 = new Object(); // Object | 
            var password = "password_example";  // string? | Passphrase for the `.p12` identity. (optional) 

            try
            {
                // Fetch unlock token (supervised)
                UnlockToken result = apiInstance.DevicesMdmFetchUnlockToken(udid, p12, password);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesMdmFetchUnlockToken: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesMdmFetchUnlockTokenWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Fetch unlock token (supervised)
    ApiResponse<UnlockToken> response = apiInstance.DevicesMdmFetchUnlockTokenWithHttpInfo(udid, p12, password);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesMdmFetchUnlockTokenWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **p12** | [**Object**](Object.md) |  |  |
| **password** | **string?** | Passphrase for the &#x60;.p12&#x60; identity. | [optional]  |

### Return type

[**UnlockToken**](UnlockToken.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: multipart/form-data
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesmdmsecurityinfo"></a>
# **DevicesMdmSecurityInfo**
> Object DevicesMdmSecurityInfo (string udid, Object p12, string? password = null)

Get MDM security info (supervised)

Get device security info (CLI: `ios mdm security-info`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesMdmSecurityInfoExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var p12 = new Object(); // Object | 
            var password = "password_example";  // string? | Passphrase for the `.p12` identity. (optional) 

            try
            {
                // Get MDM security info (supervised)
                Object result = apiInstance.DevicesMdmSecurityInfo(udid, p12, password);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesMdmSecurityInfo: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesMdmSecurityInfoWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get MDM security info (supervised)
    ApiResponse<Object> response = apiInstance.DevicesMdmSecurityInfoWithHttpInfo(udid, p12, password);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesMdmSecurityInfoWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **p12** | [**Object**](Object.md) |  |  |
| **password** | **string?** | Passphrase for the &#x60;.p12&#x60; identity. | [optional]  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: multipart/form-data
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesmemlimitoff"></a>
# **DevicesMemLimitOff**
> MemLimitResult DevicesMemLimitOff (string udid, string? process = null, MemLimitRequest? memLimitRequest = null)

Waive memory limit

Waive the memory limit for a process (CLI: `ios memlimitoff`). The process name may be given via the `process` query param or the JSON body.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesMemLimitOffExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var process = "process_example";  // string? | Process name whose memory limit should be waived. (optional) 
            var memLimitRequest = new MemLimitRequest?(); // MemLimitRequest? |  (optional) 

            try
            {
                // Waive memory limit
                MemLimitResult result = apiInstance.DevicesMemLimitOff(udid, process, memLimitRequest);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesMemLimitOff: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesMemLimitOffWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Waive memory limit
    ApiResponse<MemLimitResult> response = apiInstance.DevicesMemLimitOffWithHttpInfo(udid, process, memLimitRequest);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesMemLimitOffWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **process** | **string?** | Process name whose memory limit should be waived. | [optional]  |
| **memLimitRequest** | [**MemLimitRequest?**](MemLimitRequest?.md) |  | [optional]  |

### Return type

[**MemLimitResult**](MemLimitResult.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesmountimage"></a>
# **DevicesMountImage**
> GenericResponse DevicesMountImage (string udid, bool? auto = null, string? basedir = null, Object? body = null)

Mount a developer image

Mount a Developer Disk Image.  Either let the server auto-resolve and download the correct image (`auto=true`, optionally with `basedir`), or stream the image bytes as the raw request body (up to 2 GiB).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesMountImageExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var auto = true;  // bool? | Auto-resolve and download the matching DDI for the device. (optional) 
            var basedir = "basedir_example";  // string? | Base directory the server uses to cache/lookup DDIs when `auto=true`. (optional) 
            var body = null;  // Object? | Raw Developer Disk Image bytes (used when not auto-resolving). Content up to 2 GiB. (optional) 

            try
            {
                // Mount a developer image
                GenericResponse result = apiInstance.DevicesMountImage(udid, auto, basedir, body);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesMountImage: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesMountImageWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Mount a developer image
    ApiResponse<GenericResponse> response = apiInstance.DevicesMountImageWithHttpInfo(udid, auto, basedir, body);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesMountImageWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **auto** | **bool?** | Auto-resolve and download the matching DDI for the device. | [optional]  |
| **basedir** | **string?** | Base directory the server uses to cache/lookup DDIs when &#x60;auto&#x3D;true&#x60;. | [optional]  |
| **body** | **Object?** | Raw Developer Disk Image bytes (used when not auto-resolving). Content up to 2 GiB. | [optional]  |

### Return type

[**GenericResponse**](GenericResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/octet-stream
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicespair"></a>
# **DevicesPair**
> GenericResponse DevicesPair (string udid, bool supervised, Object p12file, string? supervisionPassword = null)

Pair device

Pair with the device.  For a supervised pairing (`supervised=true`) upload the supervision identity as `p12file` (multipart) and supply the passphrase in the `Supervision-Password` header.  Returns `423` when the device is locked and pairing cannot proceed.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesPairExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var supervised = true;  // bool | Whether this is a supervised pairing.
            var p12file = new Object(); // Object | 
            var supervisionPassword = "supervisionPassword_example";  // string? | Supervision identity passphrase (required when supervised). (optional) 

            try
            {
                // Pair device
                GenericResponse result = apiInstance.DevicesPair(udid, supervised, p12file, supervisionPassword);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesPair: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesPairWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Pair device
    ApiResponse<GenericResponse> response = apiInstance.DevicesPairWithHttpInfo(udid, supervised, p12file, supervisionPassword);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesPairWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **supervised** | **bool** | Whether this is a supervised pairing. |  |
| **p12file** | [**Object**](Object.md) |  |  |
| **supervisionPassword** | **string?** | Supervision identity passphrase (required when supervised). | [optional]  |

### Return type

[**GenericResponse**](GenericResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: multipart/form-data
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **423** | 423 — device is locked; pairing cannot proceed. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicespullfile"></a>
# **DevicesPullFile**
> Object DevicesPullFile (string udid, FileDomain domain, string remote, string? identifier = null)

Pull file

Download a file from the device, streamed as the response body (CLI: `ios file pull`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesPullFileExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var domain = new FileDomain(); // FileDomain | File service domain: `app`, `app-group`, `crash` or `temp`.
            var remote = "remote_example";  // string | Remote file path on the device.
            var identifier = "identifier_example";  // string? | Bundle/group id for the `app`/`app-group` domains. (optional) 

            try
            {
                // Pull file
                Object result = apiInstance.DevicesPullFile(udid, domain, remote, identifier);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesPullFile: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesPullFileWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Pull file
    ApiResponse<Object> response = apiInstance.DevicesPullFileWithHttpInfo(udid, domain, remote, identifier);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesPullFileWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **domain** | [**FileDomain**](FileDomain.md) | File service domain: &#x60;app&#x60;, &#x60;app-group&#x60;, &#x60;crash&#x60; or &#x60;temp&#x60;. |  |
| **remote** | **string** | Remote file path on the device. |  |
| **identifier** | **string?** | Bundle/group id for the &#x60;app&#x60;/&#x60;app-group&#x60; domains. | [optional]  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/octet-stream, application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicespushfile"></a>
# **DevicesPushFile**
> FilePushResult DevicesPushFile (string udid, FileDomain domain, string remote, Object body, string? identifier = null)

Push file

Upload the request body to a device path (CLI: `ios file push`). A `Content-Length` header is required.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesPushFileExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var domain = new FileDomain(); // FileDomain | File service domain: `app`, `app-group`, `crash` or `temp`.
            var remote = "remote_example";  // string | Destination path on the device.
            var body = null;  // Object | Raw file bytes to upload.
            var identifier = "identifier_example";  // string? | Bundle/group id for the `app`/`app-group` domains. (optional) 

            try
            {
                // Push file
                FilePushResult result = apiInstance.DevicesPushFile(udid, domain, remote, body, identifier);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesPushFile: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesPushFileWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Push file
    ApiResponse<FilePushResult> response = apiInstance.DevicesPushFileWithHttpInfo(udid, domain, remote, body, identifier);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesPushFileWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **domain** | [**FileDomain**](FileDomain.md) | File service domain: &#x60;app&#x60;, &#x60;app-group&#x60;, &#x60;crash&#x60; or &#x60;temp&#x60;. |  |
| **remote** | **string** | Destination path on the device. |  |
| **body** | **Object** | Raw file bytes to upload. |  |
| **identifier** | **string?** | Bundle/group id for the &#x60;app&#x60;/&#x60;app-group&#x60; domains. | [optional]  |

### Return type

[**FilePushResult**](FilePushResult.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/octet-stream
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesreboot"></a>
# **DevicesReboot**
> GenericResponse DevicesReboot (string udid)

Reboot device

Reboot the device (CLI: `ios reboot`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesRebootExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Reboot device
                GenericResponse result = apiInstance.DevicesReboot(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesReboot: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesRebootWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Reboot device
    ApiResponse<GenericResponse> response = apiInstance.DevicesRebootWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesRebootWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

[**GenericResponse**](GenericResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesremovecrashes"></a>
# **DevicesRemoveCrashes**
> GenericResponse DevicesRemoveCrashes (string udid, string cwd, string pattern)

Delete crash reports

Delete crash reports (CLI: `ios crash rm`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesRemoveCrashesExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var cwd = "cwd_example";  // string | Working directory on the device.
            var pattern = "pattern_example";  // string | Glob pattern of reports to delete.

            try
            {
                // Delete crash reports
                GenericResponse result = apiInstance.DevicesRemoveCrashes(udid, cwd, pattern);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesRemoveCrashes: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesRemoveCrashesWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Delete crash reports
    ApiResponse<GenericResponse> response = apiInstance.DevicesRemoveCrashesWithHttpInfo(udid, cwd, pattern);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesRemoveCrashesWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **cwd** | **string** | Working directory on the device. |  |
| **pattern** | **string** | Glob pattern of reports to delete. |  |

### Return type

[**GenericResponse**](GenericResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesremovehttpproxy"></a>
# **DevicesRemoveHttpProxy**
> GenericResponse DevicesRemoveHttpProxy (string udid)

Remove HTTP proxy

Clear the global HTTP proxy (CLI: `ios httpproxy remove`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesRemoveHttpProxyExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Remove HTTP proxy
                GenericResponse result = apiInstance.DevicesRemoveHttpProxy(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesRemoveHttpProxy: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesRemoveHttpProxyWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Remove HTTP proxy
    ApiResponse<GenericResponse> response = apiInstance.DevicesRemoveHttpProxyWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesRemoveHttpProxyWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

[**GenericResponse**](GenericResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesremoveprofile"></a>
# **DevicesRemoveProfile**
> GenericResponse DevicesRemoveProfile (string udid, string name)

Remove profile

Remove a configuration profile by identifier (CLI: `ios profile remove`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesRemoveProfileExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var name = "name_example";  // string | The profile identifier to remove.

            try
            {
                // Remove profile
                GenericResponse result = apiInstance.DevicesRemoveProfile(udid, name);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesRemoveProfile: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesRemoveProfileWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Remove profile
    ApiResponse<GenericResponse> response = apiInstance.DevicesRemoveProfileWithHttpInfo(udid, name);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesRemoveProfileWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **name** | **string** | The profile identifier to remove. |  |

### Return type

[**GenericResponse**](GenericResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesremovewifi"></a>
# **DevicesRemoveWifi**
> GenericResponse DevicesRemoveWifi (string udid, string ssid)

Remove wifi

Remove a provisioned wifi network (CLI: `ios wifi - -remove`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesRemoveWifiExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var ssid = "ssid_example";  // string | SSID of the network to remove.

            try
            {
                // Remove wifi
                GenericResponse result = apiInstance.DevicesRemoveWifi(udid, ssid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesRemoveWifi: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesRemoveWifiWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Remove wifi
    ApiResponse<GenericResponse> response = apiInstance.DevicesRemoveWifiWithHttpInfo(udid, ssid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesRemoveWifiWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **ssid** | **string** | SSID of the network to remove. |  |

### Return type

[**GenericResponse**](GenericResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesresetaccessibility"></a>
# **DevicesResetAccessibility**
> GenericResponse DevicesResetAccessibility (string udid)

Reset accessibility

Reset accessibility settings on the device.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesResetAccessibilityExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Reset accessibility
                GenericResponse result = apiInstance.DevicesResetAccessibility(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesResetAccessibility: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesResetAccessibilityWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Reset accessibility
    ApiResponse<GenericResponse> response = apiInstance.DevicesResetAccessibilityWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesResetAccessibilityWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

[**GenericResponse**](GenericResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesresetlocation"></a>
# **DevicesResetLocation**
> GenericResponse DevicesResetLocation (string udid)

Reset simulated location

Reset the simulated location back to the device's real GPS location.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesResetLocationExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Reset simulated location
                GenericResponse result = apiInstance.DevicesResetLocation(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesResetLocation: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesResetLocationWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Reset simulated location
    ApiResponse<GenericResponse> response = apiInstance.DevicesResetLocationWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesResetLocationWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

[**GenericResponse**](GenericResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesscreenshot"></a>
# **DevicesScreenshot**
> Object DevicesScreenshot (string udid)

Capture screenshot

Capture a screenshot. Returns raw PNG bytes (`image/png`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesScreenshotExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Capture screenshot
                Object result = apiInstance.DevicesScreenshot(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesScreenshot: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesScreenshotWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Capture screenshot
    ApiResponse<Object> response = apiInstance.DevicesScreenshotWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesScreenshotWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: image/png, application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicessetassistivetouch"></a>
# **DevicesSetAssistiveTouch**
> AssistiveTouchState DevicesSetAssistiveTouch (string udid, EnabledRequest enabledRequest)

Set AssistiveTouch

Enable/disable AssistiveTouch (CLI: `ios assistivetouch enable|disable`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesSetAssistiveTouchExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var enabledRequest = new EnabledRequest(); // EnabledRequest | 

            try
            {
                // Set AssistiveTouch
                AssistiveTouchState result = apiInstance.DevicesSetAssistiveTouch(udid, enabledRequest);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesSetAssistiveTouch: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesSetAssistiveTouchWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Set AssistiveTouch
    ApiResponse<AssistiveTouchState> response = apiInstance.DevicesSetAssistiveTouchWithHttpInfo(udid, enabledRequest);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesSetAssistiveTouchWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **enabledRequest** | [**EnabledRequest**](EnabledRequest.md) |  |  |

### Return type

[**AssistiveTouchState**](AssistiveTouchState.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicessetdevmode"></a>
# **DevicesSetDevMode**
> GenericResponse DevicesSetDevMode (string udid, DevModeRequest devModeRequest)

Set developer mode

Enable or reveal developer mode (CLI: `ios devmode enable|reveal`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesSetDevModeExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var devModeRequest = new DevModeRequest(); // DevModeRequest | 

            try
            {
                // Set developer mode
                GenericResponse result = apiInstance.DevicesSetDevMode(udid, devModeRequest);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesSetDevMode: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesSetDevModeWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Set developer mode
    ApiResponse<GenericResponse> response = apiInstance.DevicesSetDevModeWithHttpInfo(udid, devModeRequest);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesSetDevModeWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **devModeRequest** | [**DevModeRequest**](DevModeRequest.md) |  |  |

### Return type

[**GenericResponse**](GenericResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicessethttpproxy"></a>
# **DevicesSetHttpProxy**
> GenericResponse DevicesSetHttpProxy (string udid, string host, string port, Object p12, string? user = null, string? pass = null, string? password = null)

Set HTTP proxy (supervised)

Configure a global HTTP proxy (CLI: `ios httpproxy`). Supervised: send multipart form-data with `host`, `port`, a `p12` supervisor identity and optional `user`/`pass`/`password` fields.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesSetHttpProxyExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var host = "host_example";  // string | Proxy host.
            var port = "port_example";  // string | Proxy port.
            var p12 = new Object(); // Object | 
            var user = "user_example";  // string? | Proxy username. (optional) 
            var pass = "pass_example";  // string? | Proxy password. (optional) 
            var password = "password_example";  // string? | Passphrase for the `.p12` identity. (optional) 

            try
            {
                // Set HTTP proxy (supervised)
                GenericResponse result = apiInstance.DevicesSetHttpProxy(udid, host, port, p12, user, pass, password);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesSetHttpProxy: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesSetHttpProxyWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Set HTTP proxy (supervised)
    ApiResponse<GenericResponse> response = apiInstance.DevicesSetHttpProxyWithHttpInfo(udid, host, port, p12, user, pass, password);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesSetHttpProxyWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **host** | **string** | Proxy host. |  |
| **port** | **string** | Proxy port. |  |
| **p12** | [**Object**](Object.md) |  |  |
| **user** | **string?** | Proxy username. | [optional]  |
| **pass** | **string?** | Proxy password. | [optional]  |
| **password** | **string?** | Passphrase for the &#x60;.p12&#x60; identity. | [optional]  |

### Return type

[**GenericResponse**](GenericResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: multipart/form-data
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesseticonlayout"></a>
# **DevicesSetIconLayout**
> GenericResponse DevicesSetIconLayout (string udid, Object body)

Set icon layout

Restore a SpringBoard icon layout (CLI: `ios set-icon-layout`). Body is the layout JSON as returned by GET.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesSetIconLayoutExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var body = null;  // Object | 

            try
            {
                // Set icon layout
                GenericResponse result = apiInstance.DevicesSetIconLayout(udid, body);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesSetIconLayout: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesSetIconLayoutWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Set icon layout
    ApiResponse<GenericResponse> response = apiInstance.DevicesSetIconLayoutWithHttpInfo(udid, body);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesSetIconLayoutWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **body** | **Object** |  |  |

### Return type

[**GenericResponse**](GenericResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicessetlanguage"></a>
# **DevicesSetLanguage**
> LanguageConfiguration DevicesSetLanguage (string udid, SetLanguageRequest setLanguageRequest)

Set language

Set the device language and/or locale (CLI: `ios lang - -setlang - -setlocale`). Returns the resulting configuration.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesSetLanguageExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var setLanguageRequest = new SetLanguageRequest(); // SetLanguageRequest | 

            try
            {
                // Set language
                LanguageConfiguration result = apiInstance.DevicesSetLanguage(udid, setLanguageRequest);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesSetLanguage: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesSetLanguageWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Set language
    ApiResponse<LanguageConfiguration> response = apiInstance.DevicesSetLanguageWithHttpInfo(udid, setLanguageRequest);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesSetLanguageWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **setLanguageRequest** | [**SetLanguageRequest**](SetLanguageRequest.md) |  |  |

### Return type

[**LanguageConfiguration**](LanguageConfiguration.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicessetlocation"></a>
# **DevicesSetLocation**
> GenericResponse DevicesSetLocation (string udid, string latitude, string longitude)

Set simulated location

Simulate a GPS location on the device.  NOTE: the longitude parameter was historically misspelled `longtitude` on the wire. This spec fixes it to `longitude`; the go-ios server accepts `longitude` (and may keep `longtitude` as a deprecated alias).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesSetLocationExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var latitude = "latitude_example";  // string | Latitude in decimal degrees.
            var longitude = "longitude_example";  // string | Longitude in decimal degrees.

            try
            {
                // Set simulated location
                GenericResponse result = apiInstance.DevicesSetLocation(udid, latitude, longitude);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesSetLocation: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesSetLocationWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Set simulated location
    ApiResponse<GenericResponse> response = apiInstance.DevicesSetLocationWithHttpInfo(udid, latitude, longitude);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesSetLocationWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **latitude** | **string** | Latitude in decimal degrees. |  |
| **longitude** | **string** | Longitude in decimal degrees. |  |

### Return type

[**GenericResponse**](GenericResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicessetpasteboard"></a>
# **DevicesSetPasteboard**
> GenericResponse DevicesSetPasteboard (string udid, string body)

Set pasteboard

Set the pasteboard text from the raw request body (CLI: `ios pasteboard set`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesSetPasteboardExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var body = "body_example";  // string | 

            try
            {
                // Set pasteboard
                GenericResponse result = apiInstance.DevicesSetPasteboard(udid, body);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesSetPasteboard: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesSetPasteboardWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Set pasteboard
    ApiResponse<GenericResponse> response = apiInstance.DevicesSetPasteboardWithHttpInfo(udid, body);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesSetPasteboardWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **body** | **string** |  |  |

### Return type

[**GenericResponse**](GenericResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: text/plain
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicessettimeformat"></a>
# **DevicesSetTimeFormat**
> TimeFormatState DevicesSetTimeFormat (string udid, TimeFormatRequest timeFormatRequest)

Set time format

Set 24-hour / 12-hour clock (CLI: `ios timeformat 24h|12h`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesSetTimeFormatExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var timeFormatRequest = new TimeFormatRequest(); // TimeFormatRequest | 

            try
            {
                // Set time format
                TimeFormatState result = apiInstance.DevicesSetTimeFormat(udid, timeFormatRequest);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesSetTimeFormat: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesSetTimeFormatWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Set time format
    ApiResponse<TimeFormatState> response = apiInstance.DevicesSetTimeFormatWithHttpInfo(udid, timeFormatRequest);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesSetTimeFormatWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **timeFormatRequest** | [**TimeFormatRequest**](TimeFormatRequest.md) |  |  |

### Return type

[**TimeFormatState**](TimeFormatState.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicessetwallpaper"></a>
# **DevicesSetWallpaper**
> GenericResponse DevicesSetWallpaper (string udid, Object image, Object p12, string? password = null, string? screen = null)

Set wallpaper (supervised)

Set the wallpaper (CLI: `ios set-wallpaper`). Supervised: upload the image and a `.p12` supervisor identity as multipart form-data.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesSetWallpaperExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var image = new Object(); // Object | 
            var p12 = new Object(); // Object | 
            var password = "password_example";  // string? | Passphrase for the `.p12` identity. (optional) 
            var screen = "screen_example";  // string? | Target screen (`home`, `lock`, `both`). (optional) 

            try
            {
                // Set wallpaper (supervised)
                GenericResponse result = apiInstance.DevicesSetWallpaper(udid, image, p12, password, screen);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesSetWallpaper: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesSetWallpaperWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Set wallpaper (supervised)
    ApiResponse<GenericResponse> response = apiInstance.DevicesSetWallpaperWithHttpInfo(udid, image, p12, password, screen);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesSetWallpaperWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **image** | [**Object**](Object.md) |  |  |
| **p12** | [**Object**](Object.md) |  |  |
| **password** | **string?** | Passphrase for the &#x60;.p12&#x60; identity. | [optional]  |
| **screen** | **string?** | Target screen (&#x60;home&#x60;, &#x60;lock&#x60;, &#x60;both&#x60;). | [optional]  |

### Return type

[**GenericResponse**](GenericResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: multipart/form-data
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicessetwifi"></a>
# **DevicesSetWifi**
> GenericResponse DevicesSetWifi (string udid, WifiRequest wifiRequest)

Provision wifi

Provision a wifi network (CLI: `ios wifi`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesSetWifiExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var wifiRequest = new WifiRequest(); // WifiRequest | 

            try
            {
                // Provision wifi
                GenericResponse result = apiInstance.DevicesSetWifi(udid, wifiRequest);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesSetWifi: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesSetWifiWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Provision wifi
    ApiResponse<GenericResponse> response = apiInstance.DevicesSetWifiWithHttpInfo(udid, wifiRequest);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesSetWifiWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **wifiRequest** | [**WifiRequest**](WifiRequest.md) |  |  |

### Return type

[**GenericResponse**](GenericResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesshutdown"></a>
# **DevicesShutdown**
> GenericResponse DevicesShutdown (string udid)

Shut down device

Shut down the device (CLI: `ios shutdown`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesShutdownExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Shut down device
                GenericResponse result = apiInstance.DevicesShutdown(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesShutdown: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesShutdownWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Shut down device
    ApiResponse<GenericResponse> response = apiInstance.DevicesShutdownWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesShutdownWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

[**GenericResponse**](GenericResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesstartforward"></a>
# **DevicesStartForward**
> Job DevicesStartForward (string udid, ForwardRequest forwardRequest)

Start port forward (job)

Start a TCP port forward host→device as an async job (CLI: `ios forward`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesStartForwardExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var forwardRequest = new ForwardRequest(); // ForwardRequest | 

            try
            {
                // Start port forward (job)
                Job result = apiInstance.DevicesStartForward(udid, forwardRequest);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesStartForward: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesStartForwardWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Start port forward (job)
    ApiResponse<Job> response = apiInstance.DevicesStartForwardWithHttpInfo(udid, forwardRequest);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesStartForwardWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **forwardRequest** | [**ForwardRequest**](ForwardRequest.md) |  |  |

### Return type

[**Job**](Job.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **202** | The request has been accepted for processing, but processing has not yet completed. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesstartruntest"></a>
# **DevicesStartRunTest**
> Job DevicesStartRunTest (string udid, RunTestRequest runTestRequest)

Start test run (job)

Start an XCUITest/unit-test run as an async job (CLI: `ios runtest`). Returns `202` with the created job.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesStartRunTestExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var runTestRequest = new RunTestRequest(); // RunTestRequest | 

            try
            {
                // Start test run (job)
                Job result = apiInstance.DevicesStartRunTest(udid, runTestRequest);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesStartRunTest: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesStartRunTestWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Start test run (job)
    ApiResponse<Job> response = apiInstance.DevicesStartRunTestWithHttpInfo(udid, runTestRequest);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesStartRunTestWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **runTestRequest** | [**RunTestRequest**](RunTestRequest.md) |  |  |

### Return type

[**Job**](Job.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **202** | The request has been accepted for processing, but processing has not yet completed. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesstartrunwda"></a>
# **DevicesStartRunWda**
> Job DevicesStartRunWda (string udid, RunTestRequest? runTestRequest = null)

Start WDA runner (job)

Start the WebDriverAgent runner as an async job (CLI: `ios runwda`). Body fields are optional and default to the standard WDA bundle id and config.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesStartRunWdaExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var runTestRequest = new RunTestRequest?(); // RunTestRequest? |  (optional) 

            try
            {
                // Start WDA runner (job)
                Job result = apiInstance.DevicesStartRunWda(udid, runTestRequest);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesStartRunWda: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesStartRunWdaWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Start WDA runner (job)
    ApiResponse<Job> response = apiInstance.DevicesStartRunWdaWithHttpInfo(udid, runTestRequest);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesStartRunWdaWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **runTestRequest** | [**RunTestRequest?**](RunTestRequest?.md) |  | [optional]  |

### Return type

[**Job**](Job.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **202** | The request has been accepted for processing, but processing has not yet completed. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesstopjob"></a>
# **DevicesStopJob**
> GenericResponse DevicesStopJob (string udid, string id)

Stop or delete job

Stop a running job, or purge an already-terminal one from the registry (CLI: Ctrl-C on the equivalent command).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesStopJobExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var id = "id_example";  // string | The job id.

            try
            {
                // Stop or delete job
                GenericResponse result = apiInstance.DevicesStopJob(udid, id);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesStopJob: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesStopJobWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Stop or delete job
    ApiResponse<GenericResponse> response = apiInstance.DevicesStopJobWithHttpInfo(udid, id);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesStopJobWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **id** | **string** | The job id. |  |

### Return type

[**GenericResponse**](GenericResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — the requested resource (e.g. a job) was not found for this device. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesstreamjoblogs"></a>
# **DevicesStreamJobLogs**
> string DevicesStreamJobLogs (string udid, string id)

Stream job logs (SSE)

Stream a job's log output as Server-Sent Events: the buffered history first, then live lines until the job ends or the client disconnects.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesStreamJobLogsExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var id = "id_example";  // string | The job id.

            try
            {
                // Stream job logs (SSE)
                string result = apiInstance.DevicesStreamJobLogs(udid, id);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesStreamJobLogs: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesStreamJobLogsWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Stream job logs (SSE)
    ApiResponse<string> response = apiInstance.DevicesStreamJobLogsWithHttpInfo(udid, id);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesStreamJobLogsWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **id** | **string** | The job id. |  |

### Return type

**string**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: text/event-stream, application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — the requested resource (e.g. a job) was not found for this device. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesstreamlisten"></a>
# **DevicesStreamListen**
> string DevicesStreamListen (string udid)

Stream device attach/detach (SSE)

Stream device attach/detach events as Server-Sent Events.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesStreamListenExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Stream device attach/detach (SSE)
                string result = apiInstance.DevicesStreamListen(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesStreamListen: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesStreamListenWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Stream device attach/detach (SSE)
    ApiResponse<string> response = apiInstance.DevicesStreamListenWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesStreamListenWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

**string**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: text/event-stream, application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesstreamnotifications"></a>
# **DevicesStreamNotifications**
> string DevicesStreamNotifications (string udid)

Stream app-state notifications (SSE)

Stream application state-change notifications as Server-Sent Events.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesStreamNotificationsExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Stream app-state notifications (SSE)
                string result = apiInstance.DevicesStreamNotifications(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesStreamNotifications: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesStreamNotificationsWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Stream app-state notifications (SSE)
    ApiResponse<string> response = apiInstance.DevicesStreamNotificationsWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesStreamNotificationsWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

**string**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: text/event-stream, application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesstreamostrace"></a>
# **DevicesStreamOsTrace**
> string DevicesStreamOsTrace (string udid, int? pid = null, string? level = null, string? subsystem = null, string? match = null, string? exclude = null)

Stream os_log trace (SSE)

Stream structured os_log trace entries as Server-Sent Events. All filters are optional and combine with AND semantics.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesStreamOsTraceExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var pid = 56;  // int? | Only include entries from this process id. (optional) 
            var level = "level_example";  // string? | Minimum log level to include (e.g. `info`, `debug`, `error`). (optional) 
            var subsystem = "subsystem_example";  // string? | Only include entries from this subsystem. (optional) 
            var match = "match_example";  // string? | Only include entries whose message matches this substring/pattern. (optional) 
            var exclude = "exclude_example";  // string? | Exclude entries whose message matches this substring/pattern. (optional) 

            try
            {
                // Stream os_log trace (SSE)
                string result = apiInstance.DevicesStreamOsTrace(udid, pid, level, subsystem, match, exclude);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesStreamOsTrace: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesStreamOsTraceWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Stream os_log trace (SSE)
    ApiResponse<string> response = apiInstance.DevicesStreamOsTraceWithHttpInfo(udid, pid, level, subsystem, match, exclude);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesStreamOsTraceWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **pid** | **int?** | Only include entries from this process id. | [optional]  |
| **level** | **string?** | Minimum log level to include (e.g. &#x60;info&#x60;, &#x60;debug&#x60;, &#x60;error&#x60;). | [optional]  |
| **subsystem** | **string?** | Only include entries from this subsystem. | [optional]  |
| **match** | **string?** | Only include entries whose message matches this substring/pattern. | [optional]  |
| **exclude** | **string?** | Exclude entries whose message matches this substring/pattern. | [optional]  |

### Return type

**string**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: text/event-stream, application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesstreamsyslog"></a>
# **DevicesStreamSyslog**
> string DevicesStreamSyslog (string udid)

Stream syslog (SSE)

Stream device syslog lines as Server-Sent Events.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesStreamSyslogExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Stream syslog (SSE)
                string result = apiInstance.DevicesStreamSyslog(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesStreamSyslog: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesStreamSyslogWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Stream syslog (SSE)
    ApiResponse<string> response = apiInstance.DevicesStreamSyslogWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesStreamSyslogWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

**string**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: text/event-stream, application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesstreamsysmontap"></a>
# **DevicesStreamSysmontap**
> string DevicesStreamSysmontap (string udid)

Stream CPU usage (SSE)

Stream CPU-usage samples as Server-Sent Events (CLI: `ios sysmontap`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesStreamSysmontapExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Stream CPU usage (SSE)
                string result = apiInstance.DevicesStreamSysmontap(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesStreamSysmontap: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesStreamSysmontapWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Stream CPU usage (SSE)
    ApiResponse<string> response = apiInstance.DevicesStreamSysmontapWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesStreamSysmontapWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

**string**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: text/event-stream, application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesuninstallapp"></a>
# **DevicesUninstallApp**
> GenericResponse DevicesUninstallApp (string udid, string bundleID)

Uninstall app

Uninstall an application by bundle id.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesUninstallAppExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var bundleID = "bundleID_example";  // string | Bundle id of the app to uninstall.

            try
            {
                // Uninstall app
                GenericResponse result = apiInstance.DevicesUninstallApp(udid, bundleID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesUninstallApp: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesUninstallAppWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Uninstall app
    ApiResponse<GenericResponse> response = apiInstance.DevicesUninstallAppWithHttpInfo(udid, bundleID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesUninstallAppWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **bundleID** | **string** | Bundle id of the app to uninstall. |  |

### Return type

[**GenericResponse**](GenericResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="devicesunmountimage"></a>
# **DevicesUnmountImage**
> GenericResponse DevicesUnmountImage (string udid)

Unmount developer image

Unmount the developer disk image (CLI: `ios image unmount`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DevicesUnmountImageExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Unmount developer image
                GenericResponse result = apiInstance.DevicesUnmountImage(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DevicesUnmountImage: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DevicesUnmountImageWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Unmount developer image
    ApiResponse<GenericResponse> response = apiInstance.DevicesUnmountImageWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DevicesUnmountImageWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

[**GenericResponse**](GenericResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="diagnosticsnetgetbatteryregistry"></a>
# **DiagnosticsNetGetBatteryRegistry**
> BatteryRegistry DiagnosticsNetGetBatteryRegistry (string udid)

Get battery IORegistry

Get the battery IORegistry stats (Temperature, Voltage, CurrentCapacity, ...) via the diagnostics relay (CLI: `ios diagnostics ioregistry`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DiagnosticsNetGetBatteryRegistryExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Get battery IORegistry
                BatteryRegistry result = apiInstance.DiagnosticsNetGetBatteryRegistry(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DiagnosticsNetGetBatteryRegistry: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DiagnosticsNetGetBatteryRegistryWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get battery IORegistry
    ApiResponse<BatteryRegistry> response = apiInstance.DiagnosticsNetGetBatteryRegistryWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DiagnosticsNetGetBatteryRegistryWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

[**BatteryRegistry**](BatteryRegistry.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="diagnosticsnetgetdeviceip"></a>
# **DiagnosticsNetGetDeviceIp**
> NetworkInfo DiagnosticsNetGetDeviceIp (string udid)

Get device IP / network info

Resolve the device's network addresses (MAC/IPv4/IPv6) by sniffing pcapd (CLI: `ios ip`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DiagnosticsNetGetDeviceIpExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Get device IP / network info
                NetworkInfo result = apiInstance.DiagnosticsNetGetDeviceIp(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DiagnosticsNetGetDeviceIp: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DiagnosticsNetGetDeviceIpWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get device IP / network info
    ApiResponse<NetworkInfo> response = apiInstance.DiagnosticsNetGetDeviceIpWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DiagnosticsNetGetDeviceIpWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

[**NetworkInfo**](NetworkInfo.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="diagnosticsnetgetdiskspace"></a>
# **DiagnosticsNetGetDiskSpace**
> DiskSpaceInfo DiagnosticsNetGetDiskSpace (string udid)

Get disk space info

Get filesystem info for the device (total/free/used bytes, block size) via AFC (CLI: `ios diskspace`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DiagnosticsNetGetDiskSpaceExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Get disk space info
                DiskSpaceInfo result = apiInstance.DiagnosticsNetGetDiskSpace(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DiagnosticsNetGetDiskSpace: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DiagnosticsNetGetDiskSpaceWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get disk space info
    ApiResponse<DiskSpaceInfo> response = apiInstance.DiagnosticsNetGetDiskSpaceWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DiagnosticsNetGetDiskSpaceWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

[**DiskSpaceInfo**](DiskSpaceInfo.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="diagnosticsnetgetrsdservices"></a>
# **DiagnosticsNetGetRsdServices**
> Object DiagnosticsNetGetRsdServices (string udid)

Get RSD service list

Get the device's RSD (Remote Service Discovery) service list (CLI: `ios rsd ls`). Requires a running tunnel (iOS 17+); devices without RSD return `400`.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class DiagnosticsNetGetRsdServicesExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Get RSD service list
                Object result = apiInstance.DiagnosticsNetGetRsdServices(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.DiagnosticsNetGetRsdServices: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the DiagnosticsNetGetRsdServicesWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get RSD service list
    ApiResponse<Object> response = apiInstance.DiagnosticsNetGetRsdServicesWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.DiagnosticsNetGetRsdServicesWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="fsyncfsyncls"></a>
# **FsyncFsyncLs**
> FsyncListing FsyncFsyncLs (string udid, string? bundleID = null, string? path = null)

List a directory over AFC

List a device directory over AFC (CLI: `ios fsync ls`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class FsyncFsyncLsExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var bundleID = "bundleID_example";  // string? | App bundle id to scope to its container (else the media dir). (optional) 
            var path = "path_example";  // string? | Device-side path (rejects `..` elements). (optional) 

            try
            {
                // List a directory over AFC
                FsyncListing result = apiInstance.FsyncFsyncLs(udid, bundleID, path);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.FsyncFsyncLs: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the FsyncFsyncLsWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // List a directory over AFC
    ApiResponse<FsyncListing> response = apiInstance.FsyncFsyncLsWithHttpInfo(udid, bundleID, path);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.FsyncFsyncLsWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **bundleID** | **string?** | App bundle id to scope to its container (else the media dir). | [optional]  |
| **path** | **string?** | Device-side path (rejects &#x60;..&#x60; elements). | [optional]  |

### Return type

[**FsyncListing**](FsyncListing.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="fsyncfsyncmkdir"></a>
# **FsyncFsyncMkdir**
> FsyncMessage FsyncFsyncMkdir (string udid, string path, string? bundleID = null)

Create a directory over AFC

Create a directory over AFC (CLI: `ios fsync mkdir`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class FsyncFsyncMkdirExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var path = "path_example";  // string | Directory path to create (required).
            var bundleID = "bundleID_example";  // string? | App bundle id to scope to its container (else the media dir). (optional) 

            try
            {
                // Create a directory over AFC
                FsyncMessage result = apiInstance.FsyncFsyncMkdir(udid, path, bundleID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.FsyncFsyncMkdir: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the FsyncFsyncMkdirWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Create a directory over AFC
    ApiResponse<FsyncMessage> response = apiInstance.FsyncFsyncMkdirWithHttpInfo(udid, path, bundleID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.FsyncFsyncMkdirWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **path** | **string** | Directory path to create (required). |  |
| **bundleID** | **string?** | App bundle id to scope to its container (else the media dir). | [optional]  |

### Return type

[**FsyncMessage**](FsyncMessage.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="fsyncfsyncpull"></a>
# **FsyncFsyncPull**
> Object FsyncFsyncPull (string udid, string path, string? bundleID = null)

Download a file over AFC

Download a file from the device over AFC (CLI: `ios fsync pull`). Returns the raw file bytes. `path` is required.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class FsyncFsyncPullExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var path = "path_example";  // string | Remote file path on the device (required).
            var bundleID = "bundleID_example";  // string? | App bundle id to scope to its container (else the media dir). (optional) 

            try
            {
                // Download a file over AFC
                Object result = apiInstance.FsyncFsyncPull(udid, path, bundleID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.FsyncFsyncPull: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the FsyncFsyncPullWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Download a file over AFC
    ApiResponse<Object> response = apiInstance.FsyncFsyncPullWithHttpInfo(udid, path, bundleID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.FsyncFsyncPullWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **path** | **string** | Remote file path on the device (required). |  |
| **bundleID** | **string?** | App bundle id to scope to its container (else the media dir). | [optional]  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/octet-stream, application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="fsyncfsyncpush"></a>
# **FsyncFsyncPush**
> FsyncPushResult FsyncFsyncPush (string udid, string path, Object body, string? bundleID = null)

Upload a file over AFC

Upload a file to the device over AFC (CLI: `ios fsync push`). Accepts either raw bytes (application/octet-stream) or a multipart form with a `file` field. `path` is required. Bounded server-side; oversized uploads get `413`.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class FsyncFsyncPushExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var path = "path_example";  // string | Destination path on the device (required).
            var body = null;  // Object | Raw file bytes to upload (application/octet-stream).
            var bundleID = "bundleID_example";  // string? | App bundle id to scope to its container (else the media dir). (optional) 

            try
            {
                // Upload a file over AFC
                FsyncPushResult result = apiInstance.FsyncFsyncPush(udid, path, body, bundleID);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.FsyncFsyncPush: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the FsyncFsyncPushWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Upload a file over AFC
    ApiResponse<FsyncPushResult> response = apiInstance.FsyncFsyncPushWithHttpInfo(udid, path, body, bundleID);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.FsyncFsyncPushWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **path** | **string** | Destination path on the device (required). |  |
| **body** | **Object** | Raw file bytes to upload (application/octet-stream). |  |
| **bundleID** | **string?** | App bundle id to scope to its container (else the media dir). | [optional]  |

### Return type

[**FsyncPushResult**](FsyncPushResult.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/octet-stream
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **413** | 413 — the uploaded body exceeded the server&#39;s size cap. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="fsyncfsyncrm"></a>
# **FsyncFsyncRm**
> FsyncMessage FsyncFsyncRm (string udid, string path, string? bundleID = null, bool? recursive = null)

Remove a file or directory over AFC

Remove a file or directory over AFC (CLI: `ios fsync rm`). Pass `recursive=true` to delete a non-empty directory.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class FsyncFsyncRmExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var path = "path_example";  // string | Path to remove (required).
            var bundleID = "bundleID_example";  // string? | App bundle id to scope to its container (else the media dir). (optional) 
            var recursive = true;  // bool? | Remove directory contents recursively. (optional) 

            try
            {
                // Remove a file or directory over AFC
                FsyncMessage result = apiInstance.FsyncFsyncRm(udid, path, bundleID, recursive);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.FsyncFsyncRm: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the FsyncFsyncRmWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Remove a file or directory over AFC
    ApiResponse<FsyncMessage> response = apiInstance.FsyncFsyncRmWithHttpInfo(udid, path, bundleID, recursive);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.FsyncFsyncRmWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **path** | **string** | Path to remove (required). |  |
| **bundleID** | **string?** | App bundle id to scope to its container (else the media dir). | [optional]  |
| **recursive** | **bool?** | Remove directory contents recursively. | [optional]  |

### Return type

[**FsyncMessage**](FsyncMessage.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="fsyncfsynctree"></a>
# **FsyncFsyncTree**
> FsyncTreeListing FsyncFsyncTree (string udid, string? bundleID = null, string? path = null)

Recursively list a directory over AFC

Recursively list a device directory over AFC (CLI: `ios fsync tree`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class FsyncFsyncTreeExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var bundleID = "bundleID_example";  // string? | App bundle id to scope to its container (else the media dir). (optional) 
            var path = "path_example";  // string? | Device-side path (rejects `..` elements). (optional) 

            try
            {
                // Recursively list a directory over AFC
                FsyncTreeListing result = apiInstance.FsyncFsyncTree(udid, bundleID, path);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.FsyncFsyncTree: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the FsyncFsyncTreeWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Recursively list a directory over AFC
    ApiResponse<FsyncTreeListing> response = apiInstance.FsyncFsyncTreeWithHttpInfo(udid, bundleID, path);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.FsyncFsyncTreeWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **bundleID** | **string?** | App bundle id to scope to its container (else the media dir). | [optional]  |
| **path** | **string?** | Device-side path (rejects &#x60;..&#x60; elements). | [optional]  |

### Return type

[**FsyncTreeListing**](FsyncTreeListing.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="fsyncgetcloudconfig"></a>
# **FsyncGetCloudConfig**
> Object FsyncGetCloudConfig (string udid)

Get device cloud configuration

Get the device cloud configuration (supervision status, skip-setup options, organization info).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class FsyncGetCloudConfigExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Get device cloud configuration
                Object result = apiInstance.FsyncGetCloudConfig(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.FsyncGetCloudConfig: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the FsyncGetCloudConfigWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get device cloud configuration
    ApiResponse<Object> response = apiInstance.FsyncGetCloudConfigWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.FsyncGetCloudConfigWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="getprepareskipoptions"></a>
# **GetPrepareSkipOptions**
> PrepareSkipOptions GetPrepareSkipOptions ()

List setup skip options

List all setup-pane skip options usable when preparing a device (CLI: `ios prepare printskip`). Static, device-free list.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class GetPrepareSkipOptionsExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);

            try
            {
                // List setup skip options
                PrepareSkipOptions result = apiInstance.GetPrepareSkipOptions();
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.GetPrepareSkipOptions: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the GetPrepareSkipOptionsWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // List setup skip options
    ApiResponse<PrepareSkipOptions> response = apiInstance.GetPrepareSkipOptionsWithHttpInfo();
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.GetPrepareSkipOptionsWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters
This endpoint does not need any parameter.
### Return type

[**PrepareSkipOptions**](PrepareSkipOptions.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="listdevices"></a>
# **ListDevices**
> DeviceList ListDevices ()

List devices

List all attached / reachable devices.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class ListDevicesExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);

            try
            {
                // List devices
                DeviceList result = apiInstance.ListDevices();
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.ListDevices: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ListDevicesWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // List devices
    ApiResponse<DeviceList> response = apiInstance.ListDevicesWithHttpInfo();
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.ListDevicesWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters
This endpoint does not need any parameter.
### Return type

[**DeviceList**](DeviceList.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="listtunnels"></a>
# **ListTunnels**
> List&lt;Tunnel&gt; ListTunnels ()

List tunnels

List running device tunnels (CLI: `ios tunnel ls`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class ListTunnelsExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);

            try
            {
                // List tunnels
                List<Tunnel> result = apiInstance.ListTunnels();
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.ListTunnels: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ListTunnelsWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // List tunnels
    ApiResponse<List<Tunnel>> response = apiInstance.ListTunnelsWithHttpInfo();
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.ListTunnelsWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters
This endpoint does not need any parameter.
### Return type

[**List&lt;Tunnel&gt;**](Tunnel.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **502** | 502 — the tunnel agent could not be reached or returned an error. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="preparecreatecert"></a>
# **PrepareCreateCert**
> SupervisionCert PrepareCreateCert ()

Generate a supervision certificate

Generate a self-signed supervision identity (CLI: `ios prepare create-cert`) and return the DER (base64) and PEM for both the certificate and private key. Host-scoped (device-free).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class PrepareCreateCertExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);

            try
            {
                // Generate a supervision certificate
                SupervisionCert result = apiInstance.PrepareCreateCert();
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.PrepareCreateCert: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the PrepareCreateCertWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Generate a supervision certificate
    ApiResponse<SupervisionCert> response = apiInstance.PrepareCreateCertWithHttpInfo();
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.PrepareCreateCertWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters
This endpoint does not need any parameter.
### Return type

[**SupervisionCert**](SupervisionCert.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="preparepreparedevice"></a>
# **PreparePrepareDevice**
> PrepareResult PreparePrepareDevice (string udid, Object? cert = null, string? p12password = null, List<string>? skip = null, string? orgname = null, string? locale = null, string? lang = null)

Prepare (and optionally supervise) a device

Run the device preparation/provisioning flow (CLI: `ios prepare`). Send multipart/form-data. To supervise the device include a `cert` file (DER/PEM/P12 supervision identity) and optional `p12password`; without a cert the device is prepared without supervision.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class PreparePrepareDeviceExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var cert = new Object?(); // Object? |  (optional) 
            var p12password = "p12password_example";  // string? | P12 password (when `cert` is a P12). (optional) 
            var skip = new List<string>?(); // List<string>? | Setup panes to skip (see /prepare/skip-options). Repeatable. (optional) 
            var orgname = "orgname_example";  // string? | Supervision organization name. (optional) 
            var locale = "locale_example";  // string? | Device locale (default en_US). (optional) 
            var lang = "lang_example";  // string? | Device language (default en). (optional) 

            try
            {
                // Prepare (and optionally supervise) a device
                PrepareResult result = apiInstance.PreparePrepareDevice(udid, cert, p12password, skip, orgname, locale, lang);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.PreparePrepareDevice: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the PreparePrepareDeviceWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Prepare (and optionally supervise) a device
    ApiResponse<PrepareResult> response = apiInstance.PreparePrepareDeviceWithHttpInfo(udid, cert, p12password, skip, orgname, locale, lang);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.PreparePrepareDeviceWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **cert** | [**Object?**](Object?.md) |  | [optional]  |
| **p12password** | **string?** | P12 password (when &#x60;cert&#x60; is a P12). | [optional]  |
| **skip** | [**List&lt;string&gt;?**](string.md) | Setup panes to skip (see /prepare/skip-options). Repeatable. | [optional]  |
| **orgname** | **string?** | Supervision organization name. | [optional]  |
| **locale** | **string?** | Device locale (default en_US). | [optional]  |
| **lang** | **string?** | Device language (default en). | [optional]  |

### Return type

[**PrepareResult**](PrepareResult.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: multipart/form-data
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="refreshtunnel"></a>
# **RefreshTunnel**
> Tunnel RefreshTunnel (string udid)

Refresh tunnel

Restart the tunnel for a device and wait for it to come up (CLI: `ios tunnel refresh`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class RefreshTunnelExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Refresh tunnel
                Tunnel result = apiInstance.RefreshTunnel(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.RefreshTunnel: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the RefreshTunnelWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Refresh tunnel
    ApiResponse<Tunnel> response = apiInstance.RefreshTunnelWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.RefreshTunnelWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

[**Tunnel**](Tunnel.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **502** | 502 — the tunnel agent could not be reached or returned an error. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="shutdowntunnelagent"></a>
# **ShutdownTunnelAgent**
> AgentShutdown ShutdownTunnelAgent ()

Shut down tunnel agent

Shut down the tunnel agent (CLI: `ios tunnel stopagent`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class ShutdownTunnelAgentExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);

            try
            {
                // Shut down tunnel agent
                AgentShutdown result = apiInstance.ShutdownTunnelAgent();
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.ShutdownTunnelAgent: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the ShutdownTunnelAgentWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Shut down tunnel agent
    ApiResponse<AgentShutdown> response = apiInstance.ShutdownTunnelAgentWithHttpInfo();
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.ShutdownTunnelAgentWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters
This endpoint does not need any parameter.
### Return type

[**AgentShutdown**](AgentShutdown.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **502** | 502 — the tunnel agent could not be reached or returned an error. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="signapp"></a>
# **SignApp**
> Object SignApp (Object ipa, Object p12file, Object profile, string? p12password = null, string? bundleid = null)

Resign an app/IPA

Resign an uploaded app/IPA with an uploaded P12 identity and provisioning profile, returning the signed IPA. Synchronous. Host-scoped.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class SignAppExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var ipa = new Object(); // Object | 
            var p12file = new Object(); // Object | 
            var profile = new Object(); // Object | 
            var p12password = "p12password_example";  // string? | P12 password. (optional) 
            var bundleid = "bundleid_example";  // string? | Override bundle id. (optional) 

            try
            {
                // Resign an app/IPA
                Object result = apiInstance.SignApp(ipa, p12file, profile, p12password, bundleid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.SignApp: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the SignAppWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Resign an app/IPA
    ApiResponse<Object> response = apiInstance.SignAppWithHttpInfo(ipa, p12file, profile, p12password, bundleid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.SignAppWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **ipa** | [**Object**](Object.md) |  |  |
| **p12file** | [**Object**](Object.md) |  |  |
| **profile** | [**Object**](Object.md) |  |  |
| **p12password** | **string?** | P12 password. | [optional]  |
| **bundleid** | **string?** | Override bundle id. | [optional]  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: multipart/form-data
 - **Accept**: application/octet-stream, application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="signcertificate"></a>
# **SignCertificate**
> Object SignCertificate (Object ascPrivateKey, string ascKeyId, string ascIssuerId, string? revokeExisting = null, string? p12password = null)

Create a signing certificate

Create one App Store Connect signing certificate and return its P12 (certificate + private key) as a downloadable `application/x-pkcs12` file. The P12 password is echoed in the `X-P12-Password` response header and the certificate resource id in `X-Certificate-Id`. Host-scoped (device-free).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class SignCertificateExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var ascPrivateKey = new Object(); // Object | 
            var ascKeyId = "ascKeyId_example";  // string | App Store Connect key id.
            var ascIssuerId = "ascIssuerId_example";  // string | App Store Connect issuer id.
            var revokeExisting = "revokeExisting_example";  // string? | Revoke existing iOS Development certificates first. (optional) 
            var p12password = "p12password_example";  // string? | Password to protect the generated P12. (optional) 

            try
            {
                // Create a signing certificate
                Object result = apiInstance.SignCertificate(ascPrivateKey, ascKeyId, ascIssuerId, revokeExisting, p12password);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.SignCertificate: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the SignCertificateWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Create a signing certificate
    ApiResponse<Object> response = apiInstance.SignCertificateWithHttpInfo(ascPrivateKey, ascKeyId, ascIssuerId, revokeExisting, p12password);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.SignCertificateWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **ascPrivateKey** | [**Object**](Object.md) |  |  |
| **ascKeyId** | **string** | App Store Connect key id. |  |
| **ascIssuerId** | **string** | App Store Connect issuer id. |  |
| **revokeExisting** | **string?** | Revoke existing iOS Development certificates first. | [optional]  |
| **p12password** | **string?** | Password to protect the generated P12. | [optional]  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: multipart/form-data
 - **Accept**: application/x-pkcs12, application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |
| **502** | 502 — the tunnel agent could not be reached or returned an error. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="signprovision"></a>
# **SignProvision**
> ProvisioningResult SignProvision (Object ascPrivateKey, string ascKeyId, string ascIssuerId, string bundleid, string udid, string? bundlename = null, string? profilename = null, string? devicename = null, string? certificateId = null, string? revokeExisting = null, string? p12password = null)

Create a provisioning profile + P12

Create a bundle id, development certificate and provisioning profile via App Store Connect and return both artifacts base64-encoded in a JSON envelope. The target device udid is supplied as a form field. Host-scoped.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class SignProvisionExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var ascPrivateKey = new Object(); // Object | 
            var ascKeyId = "ascKeyId_example";  // string | App Store Connect key id.
            var ascIssuerId = "ascIssuerId_example";  // string | App Store Connect issuer id.
            var bundleid = "bundleid_example";  // string | App bundle identifier.
            var udid = "udid_example";  // string | Target device udid to register against the profile.
            var bundlename = "bundlename_example";  // string? | Bundle display name. (optional) 
            var profilename = "profilename_example";  // string? | Provisioning profile name. (optional) 
            var devicename = "devicename_example";  // string? | Device display name. (optional) 
            var certificateId = "certificateId_example";  // string? | Reuse an existing certificate (no new P12 is generated). (optional) 
            var revokeExisting = "revokeExisting_example";  // string? | Revoke existing certificates first. (optional) 
            var p12password = "p12password_example";  // string? | Password to protect the generated P12. (optional) 

            try
            {
                // Create a provisioning profile + P12
                ProvisioningResult result = apiInstance.SignProvision(ascPrivateKey, ascKeyId, ascIssuerId, bundleid, udid, bundlename, profilename, devicename, certificateId, revokeExisting, p12password);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.SignProvision: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the SignProvisionWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Create a provisioning profile + P12
    ApiResponse<ProvisioningResult> response = apiInstance.SignProvisionWithHttpInfo(ascPrivateKey, ascKeyId, ascIssuerId, bundleid, udid, bundlename, profilename, devicename, certificateId, revokeExisting, p12password);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.SignProvisionWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **ascPrivateKey** | [**Object**](Object.md) |  |  |
| **ascKeyId** | **string** | App Store Connect key id. |  |
| **ascIssuerId** | **string** | App Store Connect issuer id. |  |
| **bundleid** | **string** | App bundle identifier. |  |
| **udid** | **string** | Target device udid to register against the profile. |  |
| **bundlename** | **string?** | Bundle display name. | [optional]  |
| **profilename** | **string?** | Provisioning profile name. | [optional]  |
| **devicename** | **string?** | Device display name. | [optional]  |
| **certificateId** | **string?** | Reuse an existing certificate (no new P12 is generated). | [optional]  |
| **revokeExisting** | **string?** | Revoke existing certificates first. | [optional]  |
| **p12password** | **string?** | Password to protect the generated P12. | [optional]  |

### Return type

[**ProvisioningResult**](ProvisioningResult.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: multipart/form-data
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |
| **502** | 502 — the tunnel agent could not be reached or returned an error. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="stoptunnel"></a>
# **StopTunnel**
> TunnelStopped StopTunnel (string udid)

Stop tunnel

Stop the tunnel for a device (CLI: `ios tunnel stop - -udid`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class StopTunnelExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // Stop tunnel
                TunnelStopped result = apiInstance.StopTunnel(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.StopTunnel: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the StopTunnelWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Stop tunnel
    ApiResponse<TunnelStopped> response = apiInstance.StopTunnelWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.StopTunnelWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

[**TunnelStopped**](TunnelStopped.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **502** | 502 — the tunnel agent could not be reached or returned an error. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="streamspcap"></a>
# **StreamsPcap**
> Object StreamsPcap (string udid, int? timeout = null)

Stream a live pcap capture (binary)

Stream a live packet capture from the device as a libpcap byte stream (pipeable into wireshark/tshark). Runs until `timeout` (seconds) elapses, the default timeout is reached, or the client disconnects.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class StreamsPcapExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var timeout = 56;  // int? | Capture duration in seconds (default 60, max 3600). (optional) 

            try
            {
                // Stream a live pcap capture (binary)
                Object result = apiInstance.StreamsPcap(udid, timeout);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.StreamsPcap: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the StreamsPcapWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Stream a live pcap capture (binary)
    ApiResponse<Object> response = apiInstance.StreamsPcapWithHttpInfo(udid, timeout);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.StreamsPcapWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **timeout** | **int?** | Capture duration in seconds (default 60, max 3600). | [optional]  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/vnd.tcpdump.pcap, application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="streamsscreenshotstream"></a>
# **StreamsScreenshotStream**
> Object StreamsScreenshotStream (string udid, int? quality = null)

Stream screenshots as MJPEG (binary)

Serve an MJPEG (multipart/x-mixed-replace) stream of device screenshots captured via the instruments screenshot service. Streams until the client disconnects or the source fails.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class StreamsScreenshotStreamExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var quality = 56;  // int? | Optional JPEG quality (1–100, default 80). (optional) 

            try
            {
                // Stream screenshots as MJPEG (binary)
                Object result = apiInstance.StreamsScreenshotStream(udid, quality);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.StreamsScreenshotStream: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the StreamsScreenshotStreamWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Stream screenshots as MJPEG (binary)
    ApiResponse<Object> response = apiInstance.StreamsScreenshotStreamWithHttpInfo(udid, quality);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.StreamsScreenshotStreamWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **quality** | **int?** | Optional JPEG quality (1–100, default 80). | [optional]  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: image/jpeg, application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="streamsuistream"></a>
# **StreamsUiStream**
> Object StreamsUiStream (string udid, string? backend = null, string? wdaUrl = null, int? timeout = null, string? codec = null, string? fps = null, string? quality = null, string? scale = null, string? bitrate = null)

Stream UI video (binary)

Open a live UI video stream against a forwarded WDA/DeviceKit backend and pipe it straight through to the client. Default codec is MJPEG (multipart/x-mixed-replace); `codec=h264` returns an H.264 elementary stream (requires the devicekit backend). Streams until the client disconnects or the backend ends.  Requires a running, forwarded WDA/DeviceKit backend (see the UI routes).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class StreamsUiStreamExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var backend = "backend_example";  // string? | Backend to target: `wda` (default) or `devicekit`. (optional) 
            var wdaUrl = "wdaUrl_example";  // string? | Forwarded backend base URL (defaults per backend). (optional) 
            var timeout = 56;  // int? | Per-request HTTP timeout in seconds (default 60). (optional) 
            var codec = "codec_example";  // string? | Video codec: `mjpeg` (default) or `h264` (devicekit backend only). (optional) 
            var fps = "fps_example";  // string? | Target frames per second (backend-dependent). (optional) 
            var quality = "quality_example";  // string? | JPEG quality for the mjpeg codec. (optional) 
            var scale = "scale_example";  // string? | Scale factor (backend-dependent). (optional) 
            var bitrate = "bitrate_example";  // string? | Target bitrate for the h264 codec. (optional) 

            try
            {
                // Stream UI video (binary)
                Object result = apiInstance.StreamsUiStream(udid, backend, wdaUrl, timeout, codec, fps, quality, scale, bitrate);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.StreamsUiStream: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the StreamsUiStreamWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Stream UI video (binary)
    ApiResponse<Object> response = apiInstance.StreamsUiStreamWithHttpInfo(udid, backend, wdaUrl, timeout, codec, fps, quality, scale, bitrate);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.StreamsUiStreamWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **backend** | **string?** | Backend to target: &#x60;wda&#x60; (default) or &#x60;devicekit&#x60;. | [optional]  |
| **wdaUrl** | **string?** | Forwarded backend base URL (defaults per backend). | [optional]  |
| **timeout** | **int?** | Per-request HTTP timeout in seconds (default 60). | [optional]  |
| **codec** | **string?** | Video codec: &#x60;mjpeg&#x60; (default) or &#x60;h264&#x60; (devicekit backend only). | [optional]  |
| **fps** | **string?** | Target frames per second (backend-dependent). | [optional]  |
| **quality** | **string?** | JPEG quality for the mjpeg codec. | [optional]  |
| **scale** | **string?** | Scale factor (backend-dependent). | [optional]  |
| **bitrate** | **string?** | Target bitrate for the h264 codec. | [optional]  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/octet-stream, application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |
| **501** | 501 — the selected UI-automation backend does not support this operation. |  -  |
| **502** | 502 — the tunnel agent could not be reached or returned an error. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="uiuiapi"></a>
# **UIUiApi**
> Object UIUiApi (string udid, UIAPIRequest uIAPIRequest, string? backend = null, string? wdaUrl = null, int? timeout = null)

Raw backend passthrough

Raw passthrough to the backend. For WDA supply `method`/`path`/`body`; for DeviceKit supply `rpcMethod`/`rpcParams`. The backend response is forwarded verbatim.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class UIUiApiExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var uIAPIRequest = new UIAPIRequest(); // UIAPIRequest | 
            var backend = "backend_example";  // string? | Backend to target: `wda` (default) or `devicekit`. (optional) 
            var wdaUrl = "wdaUrl_example";  // string? | Forwarded backend base URL (defaults per backend). (optional) 
            var timeout = 56;  // int? | Per-request HTTP timeout in seconds (default 60). (optional) 

            try
            {
                // Raw backend passthrough
                Object result = apiInstance.UIUiApi(udid, uIAPIRequest, backend, wdaUrl, timeout);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.UIUiApi: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UIUiApiWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Raw backend passthrough
    ApiResponse<Object> response = apiInstance.UIUiApiWithHttpInfo(udid, uIAPIRequest, backend, wdaUrl, timeout);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.UIUiApiWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **uIAPIRequest** | [**UIAPIRequest**](UIAPIRequest.md) |  |  |
| **backend** | **string?** | Backend to target: &#x60;wda&#x60; (default) or &#x60;devicekit&#x60;. | [optional]  |
| **wdaUrl** | **string?** | Forwarded backend base URL (defaults per backend). | [optional]  |
| **timeout** | **int?** | Per-request HTTP timeout in seconds (default 60). | [optional]  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |
| **501** | 501 — the selected UI-automation backend does not support this operation. |  -  |
| **502** | 502 — the tunnel agent could not be reached or returned an error. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="uiuiappforeground"></a>
# **UIUiAppForeground**
> Object UIUiAppForeground (string udid, string? backend = null, string? wdaUrl = null, int? timeout = null)

Foreground app (UI backend)

Bring the backgrounded app to the foreground. Only the devicekit backend supports this; WDA returns `501`. The request body is ignored.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class UIUiAppForegroundExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var backend = "backend_example";  // string? | Backend to target: `wda` (default) or `devicekit`. (optional) 
            var wdaUrl = "wdaUrl_example";  // string? | Forwarded backend base URL (defaults per backend). (optional) 
            var timeout = 56;  // int? | Per-request HTTP timeout in seconds (default 60). (optional) 

            try
            {
                // Foreground app (UI backend)
                Object result = apiInstance.UIUiAppForeground(udid, backend, wdaUrl, timeout);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.UIUiAppForeground: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UIUiAppForegroundWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Foreground app (UI backend)
    ApiResponse<Object> response = apiInstance.UIUiAppForegroundWithHttpInfo(udid, backend, wdaUrl, timeout);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.UIUiAppForegroundWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **backend** | **string?** | Backend to target: &#x60;wda&#x60; (default) or &#x60;devicekit&#x60;. | [optional]  |
| **wdaUrl** | **string?** | Forwarded backend base URL (defaults per backend). | [optional]  |
| **timeout** | **int?** | Per-request HTTP timeout in seconds (default 60). | [optional]  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |
| **501** | 501 — the selected UI-automation backend does not support this operation. |  -  |
| **502** | 502 — the tunnel agent could not be reached or returned an error. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="uiuiapplaunch"></a>
# **UIUiAppLaunch**
> Object UIUiAppLaunch (string udid, UIAppRequest uIAppRequest, string? backend = null, string? wdaUrl = null, int? timeout = null)

Launch app (UI backend)

Launch the app identified by `bundleId`.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class UIUiAppLaunchExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var uIAppRequest = new UIAppRequest(); // UIAppRequest | 
            var backend = "backend_example";  // string? | Backend to target: `wda` (default) or `devicekit`. (optional) 
            var wdaUrl = "wdaUrl_example";  // string? | Forwarded backend base URL (defaults per backend). (optional) 
            var timeout = 56;  // int? | Per-request HTTP timeout in seconds (default 60). (optional) 

            try
            {
                // Launch app (UI backend)
                Object result = apiInstance.UIUiAppLaunch(udid, uIAppRequest, backend, wdaUrl, timeout);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.UIUiAppLaunch: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UIUiAppLaunchWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Launch app (UI backend)
    ApiResponse<Object> response = apiInstance.UIUiAppLaunchWithHttpInfo(udid, uIAppRequest, backend, wdaUrl, timeout);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.UIUiAppLaunchWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **uIAppRequest** | [**UIAppRequest**](UIAppRequest.md) |  |  |
| **backend** | **string?** | Backend to target: &#x60;wda&#x60; (default) or &#x60;devicekit&#x60;. | [optional]  |
| **wdaUrl** | **string?** | Forwarded backend base URL (defaults per backend). | [optional]  |
| **timeout** | **int?** | Per-request HTTP timeout in seconds (default 60). | [optional]  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |
| **501** | 501 — the selected UI-automation backend does not support this operation. |  -  |
| **502** | 502 — the tunnel agent could not be reached or returned an error. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="uiuiappterminate"></a>
# **UIUiAppTerminate**
> Object UIUiAppTerminate (string udid, UIAppRequest uIAppRequest, string? backend = null, string? wdaUrl = null, int? timeout = null)

Terminate app (UI backend)

Terminate the app identified by `bundleId`.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class UIUiAppTerminateExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var uIAppRequest = new UIAppRequest(); // UIAppRequest | 
            var backend = "backend_example";  // string? | Backend to target: `wda` (default) or `devicekit`. (optional) 
            var wdaUrl = "wdaUrl_example";  // string? | Forwarded backend base URL (defaults per backend). (optional) 
            var timeout = 56;  // int? | Per-request HTTP timeout in seconds (default 60). (optional) 

            try
            {
                // Terminate app (UI backend)
                Object result = apiInstance.UIUiAppTerminate(udid, uIAppRequest, backend, wdaUrl, timeout);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.UIUiAppTerminate: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UIUiAppTerminateWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Terminate app (UI backend)
    ApiResponse<Object> response = apiInstance.UIUiAppTerminateWithHttpInfo(udid, uIAppRequest, backend, wdaUrl, timeout);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.UIUiAppTerminateWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **uIAppRequest** | [**UIAppRequest**](UIAppRequest.md) |  |  |
| **backend** | **string?** | Backend to target: &#x60;wda&#x60; (default) or &#x60;devicekit&#x60;. | [optional]  |
| **wdaUrl** | **string?** | Forwarded backend base URL (defaults per backend). | [optional]  |
| **timeout** | **int?** | Per-request HTTP timeout in seconds (default 60). | [optional]  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |
| **501** | 501 — the selected UI-automation backend does not support this operation. |  -  |
| **502** | 502 — the tunnel agent could not be reached or returned an error. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="uiuibutton"></a>
# **UIUiButton**
> Object UIUiButton (string udid, UIButtonRequest uIButtonRequest, string? backend = null, string? wdaUrl = null, int? timeout = null)

Press hardware button

Press a hardware button by name (WDA supports only `home`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class UIUiButtonExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var uIButtonRequest = new UIButtonRequest(); // UIButtonRequest | 
            var backend = "backend_example";  // string? | Backend to target: `wda` (default) or `devicekit`. (optional) 
            var wdaUrl = "wdaUrl_example";  // string? | Forwarded backend base URL (defaults per backend). (optional) 
            var timeout = 56;  // int? | Per-request HTTP timeout in seconds (default 60). (optional) 

            try
            {
                // Press hardware button
                Object result = apiInstance.UIUiButton(udid, uIButtonRequest, backend, wdaUrl, timeout);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.UIUiButton: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UIUiButtonWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Press hardware button
    ApiResponse<Object> response = apiInstance.UIUiButtonWithHttpInfo(udid, uIButtonRequest, backend, wdaUrl, timeout);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.UIUiButtonWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **uIButtonRequest** | [**UIButtonRequest**](UIButtonRequest.md) |  |  |
| **backend** | **string?** | Backend to target: &#x60;wda&#x60; (default) or &#x60;devicekit&#x60;. | [optional]  |
| **wdaUrl** | **string?** | Forwarded backend base URL (defaults per backend). | [optional]  |
| **timeout** | **int?** | Per-request HTTP timeout in seconds (default 60). | [optional]  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |
| **501** | 501 — the selected UI-automation backend does not support this operation. |  -  |
| **502** | 502 — the tunnel agent could not be reached or returned an error. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="uiuigetorientation"></a>
# **UIUiGetOrientation**
> Object UIUiGetOrientation (string udid, string? backend = null, string? wdaUrl = null, int? timeout = null)

Get orientation

Get the current device orientation payload.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class UIUiGetOrientationExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var backend = "backend_example";  // string? | Backend to target: `wda` (default) or `devicekit`. (optional) 
            var wdaUrl = "wdaUrl_example";  // string? | Forwarded backend base URL (defaults per backend). (optional) 
            var timeout = 56;  // int? | Per-request HTTP timeout in seconds (default 60). (optional) 

            try
            {
                // Get orientation
                Object result = apiInstance.UIUiGetOrientation(udid, backend, wdaUrl, timeout);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.UIUiGetOrientation: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UIUiGetOrientationWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Get orientation
    ApiResponse<Object> response = apiInstance.UIUiGetOrientationWithHttpInfo(udid, backend, wdaUrl, timeout);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.UIUiGetOrientationWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **backend** | **string?** | Backend to target: &#x60;wda&#x60; (default) or &#x60;devicekit&#x60;. | [optional]  |
| **wdaUrl** | **string?** | Forwarded backend base URL (defaults per backend). | [optional]  |
| **timeout** | **int?** | Per-request HTTP timeout in seconds (default 60). | [optional]  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |
| **501** | 501 — the selected UI-automation backend does not support this operation. |  -  |
| **502** | 502 — the tunnel agent could not be reached or returned an error. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="uiuilongpress"></a>
# **UIUiLongPress**
> Object UIUiLongPress (string udid, UILongPressRequest uILongPressRequest, string? backend = null, string? wdaUrl = null, int? timeout = null)

Long press

Press and hold at (x,y).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class UIUiLongPressExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var uILongPressRequest = new UILongPressRequest(); // UILongPressRequest | 
            var backend = "backend_example";  // string? | Backend to target: `wda` (default) or `devicekit`. (optional) 
            var wdaUrl = "wdaUrl_example";  // string? | Forwarded backend base URL (defaults per backend). (optional) 
            var timeout = 56;  // int? | Per-request HTTP timeout in seconds (default 60). (optional) 

            try
            {
                // Long press
                Object result = apiInstance.UIUiLongPress(udid, uILongPressRequest, backend, wdaUrl, timeout);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.UIUiLongPress: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UIUiLongPressWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Long press
    ApiResponse<Object> response = apiInstance.UIUiLongPressWithHttpInfo(udid, uILongPressRequest, backend, wdaUrl, timeout);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.UIUiLongPressWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **uILongPressRequest** | [**UILongPressRequest**](UILongPressRequest.md) |  |  |
| **backend** | **string?** | Backend to target: &#x60;wda&#x60; (default) or &#x60;devicekit&#x60;. | [optional]  |
| **wdaUrl** | **string?** | Forwarded backend base URL (defaults per backend). | [optional]  |
| **timeout** | **int?** | Per-request HTTP timeout in seconds (default 60). | [optional]  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |
| **501** | 501 — the selected UI-automation backend does not support this operation. |  -  |
| **502** | 502 — the tunnel agent could not be reached or returned an error. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="uiuiscreenshot"></a>
# **UIUiScreenshot**
> Object UIUiScreenshot (string udid, string? backend = null, string? wdaUrl = null, int? timeout = null)

UI screenshot (PNG)

Capture the screen and return raw PNG bytes.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class UIUiScreenshotExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var backend = "backend_example";  // string? | Backend to target: `wda` (default) or `devicekit`. (optional) 
            var wdaUrl = "wdaUrl_example";  // string? | Forwarded backend base URL (defaults per backend). (optional) 
            var timeout = 56;  // int? | Per-request HTTP timeout in seconds (default 60). (optional) 

            try
            {
                // UI screenshot (PNG)
                Object result = apiInstance.UIUiScreenshot(udid, backend, wdaUrl, timeout);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.UIUiScreenshot: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UIUiScreenshotWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // UI screenshot (PNG)
    ApiResponse<Object> response = apiInstance.UIUiScreenshotWithHttpInfo(udid, backend, wdaUrl, timeout);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.UIUiScreenshotWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **backend** | **string?** | Backend to target: &#x60;wda&#x60; (default) or &#x60;devicekit&#x60;. | [optional]  |
| **wdaUrl** | **string?** | Forwarded backend base URL (defaults per backend). | [optional]  |
| **timeout** | **int?** | Per-request HTTP timeout in seconds (default 60). | [optional]  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: image/png, application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |
| **501** | 501 — the selected UI-automation backend does not support this operation. |  -  |
| **502** | 502 — the tunnel agent could not be reached or returned an error. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="uiuisetorientation"></a>
# **UIUiSetOrientation**
> Object UIUiSetOrientation (string udid, UIOrientationRequest uIOrientationRequest, string? backend = null, string? wdaUrl = null, int? timeout = null)

Set orientation

Set the device orientation.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class UIUiSetOrientationExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var uIOrientationRequest = new UIOrientationRequest(); // UIOrientationRequest | 
            var backend = "backend_example";  // string? | Backend to target: `wda` (default) or `devicekit`. (optional) 
            var wdaUrl = "wdaUrl_example";  // string? | Forwarded backend base URL (defaults per backend). (optional) 
            var timeout = 56;  // int? | Per-request HTTP timeout in seconds (default 60). (optional) 

            try
            {
                // Set orientation
                Object result = apiInstance.UIUiSetOrientation(udid, uIOrientationRequest, backend, wdaUrl, timeout);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.UIUiSetOrientation: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UIUiSetOrientationWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Set orientation
    ApiResponse<Object> response = apiInstance.UIUiSetOrientationWithHttpInfo(udid, uIOrientationRequest, backend, wdaUrl, timeout);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.UIUiSetOrientationWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **uIOrientationRequest** | [**UIOrientationRequest**](UIOrientationRequest.md) |  |  |
| **backend** | **string?** | Backend to target: &#x60;wda&#x60; (default) or &#x60;devicekit&#x60;. | [optional]  |
| **wdaUrl** | **string?** | Forwarded backend base URL (defaults per backend). | [optional]  |
| **timeout** | **int?** | Per-request HTTP timeout in seconds (default 60). | [optional]  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |
| **501** | 501 — the selected UI-automation backend does not support this operation. |  -  |
| **502** | 502 — the tunnel agent could not be reached or returned an error. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="uiuisource"></a>
# **UIUiSource**
> Object UIUiSource (string udid, string? backend = null, string? wdaUrl = null, int? timeout = null)

UI source hierarchy

Return the current view hierarchy (XML for WDA; backend Content-Type preserved).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class UIUiSourceExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var backend = "backend_example";  // string? | Backend to target: `wda` (default) or `devicekit`. (optional) 
            var wdaUrl = "wdaUrl_example";  // string? | Forwarded backend base URL (defaults per backend). (optional) 
            var timeout = 56;  // int? | Per-request HTTP timeout in seconds (default 60). (optional) 

            try
            {
                // UI source hierarchy
                Object result = apiInstance.UIUiSource(udid, backend, wdaUrl, timeout);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.UIUiSource: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UIUiSourceWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // UI source hierarchy
    ApiResponse<Object> response = apiInstance.UIUiSourceWithHttpInfo(udid, backend, wdaUrl, timeout);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.UIUiSourceWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **backend** | **string?** | Backend to target: &#x60;wda&#x60; (default) or &#x60;devicekit&#x60;. | [optional]  |
| **wdaUrl** | **string?** | Forwarded backend base URL (defaults per backend). | [optional]  |
| **timeout** | **int?** | Per-request HTTP timeout in seconds (default 60). | [optional]  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/xml, application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |
| **501** | 501 — the selected UI-automation backend does not support this operation. |  -  |
| **502** | 502 — the tunnel agent could not be reached or returned an error. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="uiuistatus"></a>
# **UIUiStatus**
> Object UIUiStatus (string udid, string? backend = null, string? wdaUrl = null, int? timeout = null)

UI backend status

Return the backend status/health payload (WDA /status or DeviceKit /health).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class UIUiStatusExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var backend = "backend_example";  // string? | Backend to target: `wda` (default) or `devicekit`. (optional) 
            var wdaUrl = "wdaUrl_example";  // string? | Forwarded backend base URL (defaults per backend). (optional) 
            var timeout = 56;  // int? | Per-request HTTP timeout in seconds (default 60). (optional) 

            try
            {
                // UI backend status
                Object result = apiInstance.UIUiStatus(udid, backend, wdaUrl, timeout);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.UIUiStatus: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UIUiStatusWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // UI backend status
    ApiResponse<Object> response = apiInstance.UIUiStatusWithHttpInfo(udid, backend, wdaUrl, timeout);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.UIUiStatusWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **backend** | **string?** | Backend to target: &#x60;wda&#x60; (default) or &#x60;devicekit&#x60;. | [optional]  |
| **wdaUrl** | **string?** | Forwarded backend base URL (defaults per backend). | [optional]  |
| **timeout** | **int?** | Per-request HTTP timeout in seconds (default 60). | [optional]  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |
| **501** | 501 — the selected UI-automation backend does not support this operation. |  -  |
| **502** | 502 — the tunnel agent could not be reached or returned an error. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="uiuiswipe"></a>
# **UIUiSwipe**
> Object UIUiSwipe (string udid, UISwipeRequest uISwipeRequest, string? backend = null, string? wdaUrl = null, int? timeout = null)

Swipe

Drag from (x1,y1) to (x2,y2).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class UIUiSwipeExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var uISwipeRequest = new UISwipeRequest(); // UISwipeRequest | 
            var backend = "backend_example";  // string? | Backend to target: `wda` (default) or `devicekit`. (optional) 
            var wdaUrl = "wdaUrl_example";  // string? | Forwarded backend base URL (defaults per backend). (optional) 
            var timeout = 56;  // int? | Per-request HTTP timeout in seconds (default 60). (optional) 

            try
            {
                // Swipe
                Object result = apiInstance.UIUiSwipe(udid, uISwipeRequest, backend, wdaUrl, timeout);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.UIUiSwipe: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UIUiSwipeWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Swipe
    ApiResponse<Object> response = apiInstance.UIUiSwipeWithHttpInfo(udid, uISwipeRequest, backend, wdaUrl, timeout);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.UIUiSwipeWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **uISwipeRequest** | [**UISwipeRequest**](UISwipeRequest.md) |  |  |
| **backend** | **string?** | Backend to target: &#x60;wda&#x60; (default) or &#x60;devicekit&#x60;. | [optional]  |
| **wdaUrl** | **string?** | Forwarded backend base URL (defaults per backend). | [optional]  |
| **timeout** | **int?** | Per-request HTTP timeout in seconds (default 60). | [optional]  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |
| **501** | 501 — the selected UI-automation backend does not support this operation. |  -  |
| **502** | 502 — the tunnel agent could not be reached or returned an error. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="uiuitap"></a>
# **UIUiTap**
> Object UIUiTap (string udid, UITapRequest uITapRequest, string? backend = null, string? wdaUrl = null, int? timeout = null)

Tap

Tap at absolute coordinates.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class UIUiTapExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var uITapRequest = new UITapRequest(); // UITapRequest | 
            var backend = "backend_example";  // string? | Backend to target: `wda` (default) or `devicekit`. (optional) 
            var wdaUrl = "wdaUrl_example";  // string? | Forwarded backend base URL (defaults per backend). (optional) 
            var timeout = 56;  // int? | Per-request HTTP timeout in seconds (default 60). (optional) 

            try
            {
                // Tap
                Object result = apiInstance.UIUiTap(udid, uITapRequest, backend, wdaUrl, timeout);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.UIUiTap: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UIUiTapWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Tap
    ApiResponse<Object> response = apiInstance.UIUiTapWithHttpInfo(udid, uITapRequest, backend, wdaUrl, timeout);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.UIUiTapWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **uITapRequest** | [**UITapRequest**](UITapRequest.md) |  |  |
| **backend** | **string?** | Backend to target: &#x60;wda&#x60; (default) or &#x60;devicekit&#x60;. | [optional]  |
| **wdaUrl** | **string?** | Forwarded backend base URL (defaults per backend). | [optional]  |
| **timeout** | **int?** | Per-request HTTP timeout in seconds (default 60). | [optional]  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |
| **501** | 501 — the selected UI-automation backend does not support this operation. |  -  |
| **502** | 502 — the tunnel agent could not be reached or returned an error. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="uiuitype"></a>
# **UIUiType**
> Object UIUiType (string udid, UITypeRequest uITypeRequest, string? backend = null, string? wdaUrl = null, int? timeout = null)

Type text

Send text as keyboard input.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class UIUiTypeExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var uITypeRequest = new UITypeRequest(); // UITypeRequest | 
            var backend = "backend_example";  // string? | Backend to target: `wda` (default) or `devicekit`. (optional) 
            var wdaUrl = "wdaUrl_example";  // string? | Forwarded backend base URL (defaults per backend). (optional) 
            var timeout = 56;  // int? | Per-request HTTP timeout in seconds (default 60). (optional) 

            try
            {
                // Type text
                Object result = apiInstance.UIUiType(udid, uITypeRequest, backend, wdaUrl, timeout);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.UIUiType: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UIUiTypeWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Type text
    ApiResponse<Object> response = apiInstance.UIUiTypeWithHttpInfo(udid, uITypeRequest, backend, wdaUrl, timeout);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.UIUiTypeWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **uITypeRequest** | [**UITypeRequest**](UITypeRequest.md) |  |  |
| **backend** | **string?** | Backend to target: &#x60;wda&#x60; (default) or &#x60;devicekit&#x60;. | [optional]  |
| **wdaUrl** | **string?** | Forwarded backend base URL (defaults per backend). | [optional]  |
| **timeout** | **int?** | Per-request HTTP timeout in seconds (default 60). | [optional]  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |
| **501** | 501 — the selected UI-automation backend does not support this operation. |  -  |
| **502** | 502 — the tunnel agent could not be reached or returned an error. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="uiuiwindowsize"></a>
# **UIUiWindowSize**
> Object UIUiWindowSize (string udid, string? backend = null, string? wdaUrl = null, int? timeout = null)

UI window size

Return the device window/screen size payload (typically {width,height}).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class UIUiWindowSizeExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var backend = "backend_example";  // string? | Backend to target: `wda` (default) or `devicekit`. (optional) 
            var wdaUrl = "wdaUrl_example";  // string? | Forwarded backend base URL (defaults per backend). (optional) 
            var timeout = 56;  // int? | Per-request HTTP timeout in seconds (default 60). (optional) 

            try
            {
                // UI window size
                Object result = apiInstance.UIUiWindowSize(udid, backend, wdaUrl, timeout);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.UIUiWindowSize: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the UIUiWindowSizeWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // UI window size
    ApiResponse<Object> response = apiInstance.UIUiWindowSizeWithHttpInfo(udid, backend, wdaUrl, timeout);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.UIUiWindowSizeWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **backend** | **string?** | Backend to target: &#x60;wda&#x60; (default) or &#x60;devicekit&#x60;. | [optional]  |
| **wdaUrl** | **string?** | Forwarded backend base URL (defaults per backend). | [optional]  |
| **timeout** | **int?** | Per-request HTTP timeout in seconds (default 60). | [optional]  |

### Return type

**Object**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |
| **501** | 501 — the selected UI-automation backend does not support this operation. |  -  |
| **502** | 502 — the tunnel agent could not be reached or returned an error. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="webinspectorwebinspectoreval"></a>
# **WebInspectorWebInspectorEval**
> WebInspectorEvalResult WebInspectorWebInspectorEval (string udid, WebInspectorEvalRequest webInspectorEvalRequest)

Evaluate JavaScript in a page

Evaluate JavaScript in an inspectable page and return the result (CLI: `ios webinspector eval`). `404` when no matching page exists.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class WebInspectorWebInspectorEvalExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var webInspectorEvalRequest = new WebInspectorEvalRequest(); // WebInspectorEvalRequest | 

            try
            {
                // Evaluate JavaScript in a page
                WebInspectorEvalResult result = apiInstance.WebInspectorWebInspectorEval(udid, webInspectorEvalRequest);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.WebInspectorWebInspectorEval: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the WebInspectorWebInspectorEvalWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Evaluate JavaScript in a page
    ApiResponse<WebInspectorEvalResult> response = apiInstance.WebInspectorWebInspectorEvalWithHttpInfo(udid, webInspectorEvalRequest);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.WebInspectorWebInspectorEvalWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **webInspectorEvalRequest** | [**WebInspectorEvalRequest**](WebInspectorEvalRequest.md) |  |  |

### Return type

[**WebInspectorEvalResult**](WebInspectorEvalResult.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — the requested resource (e.g. a job) was not found for this device. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **424** | 424 — a device-side prerequisite is missing. Used by the WebInspector routes when Web Inspector / Remote Automation is not enabled on the device. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="webinspectorwebinspectorlaunch"></a>
# **WebInspectorWebInspectorLaunch**
> WebInspectorLaunchResult WebInspectorWebInspectorLaunch (string udid, string? url = null, WebInspectorLaunchRequest? webInspectorLaunchRequest = null)

Open a URL in a new inspectable page

Open a URL in a new inspectable page via a remote automation session (CLI: `ios webinspector launch <url>`). `url` may be a query param or in the body; `bundleId` defaults to Safari.

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class WebInspectorWebInspectorLaunchExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 
            var url = "url_example";  // string? | URL to open (alternative to the request body). (optional) 
            var webInspectorLaunchRequest = new WebInspectorLaunchRequest?(); // WebInspectorLaunchRequest? |  (optional) 

            try
            {
                // Open a URL in a new inspectable page
                WebInspectorLaunchResult result = apiInstance.WebInspectorWebInspectorLaunch(udid, url, webInspectorLaunchRequest);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.WebInspectorWebInspectorLaunch: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the WebInspectorWebInspectorLaunchWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // Open a URL in a new inspectable page
    ApiResponse<WebInspectorLaunchResult> response = apiInstance.WebInspectorWebInspectorLaunchWithHttpInfo(udid, url, webInspectorLaunchRequest);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.WebInspectorWebInspectorLaunchWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |
| **url** | **string?** | URL to open (alternative to the request body). | [optional]  |
| **webInspectorLaunchRequest** | [**WebInspectorLaunchRequest?**](WebInspectorLaunchRequest?.md) |  | [optional]  |

### Return type

[**WebInspectorLaunchResult**](WebInspectorLaunchResult.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **400** | 400 — malformed request (missing required query/body, bad payload). |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **424** | 424 — a device-side prerequisite is missing. Used by the WebInspector routes when Web Inspector / Remote Automation is not enabled on the device. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

<a id="webinspectorwebinspectorpages"></a>
# **WebInspectorWebInspectorPages**
> List&lt;Object&gt; WebInspectorWebInspectorPages (string udid)

List inspectable pages

List inspectable pages reported by the device (CLI: `ios webinspector list`).

### Example
```csharp
using System.Collections.Generic;
using System.Diagnostics;
using System.Net.Http;
using GoIos.Sdk.Generated.Api;
using GoIos.Sdk.Generated.Client;
using GoIos.Sdk.Generated.Model;

namespace Example
{
    public class WebInspectorWebInspectorPagesExample
    {
        public static void Main()
        {
            Configuration config = new Configuration();
            config.BasePath = "http://localhost:60105";
            // Configure Bearer token for authorization: BearerAuth
            config.AccessToken = "YOUR_BEARER_TOKEN";

            // create instances of HttpClient, HttpClientHandler to be reused later with different Api classes
            HttpClient httpClient = new HttpClient();
            HttpClientHandler httpClientHandler = new HttpClientHandler();
            var apiInstance = new DefaultApi(httpClient, config, httpClientHandler);
            var udid = "udid_example";  // string | 

            try
            {
                // List inspectable pages
                List<Object> result = apiInstance.WebInspectorWebInspectorPages(udid);
                Debug.WriteLine(result);
            }
            catch (ApiException  e)
            {
                Debug.Print("Exception when calling DefaultApi.WebInspectorWebInspectorPages: " + e.Message);
                Debug.Print("Status Code: " + e.ErrorCode);
                Debug.Print(e.StackTrace);
            }
        }
    }
}
```

#### Using the WebInspectorWebInspectorPagesWithHttpInfo variant
This returns an ApiResponse object which contains the response data, status code and headers.

```csharp
try
{
    // List inspectable pages
    ApiResponse<List<Object>> response = apiInstance.WebInspectorWebInspectorPagesWithHttpInfo(udid);
    Debug.Write("Status Code: " + response.StatusCode);
    Debug.Write("Response Headers: " + response.Headers);
    Debug.Write("Response Body: " + response.Data);
}
catch (ApiException e)
{
    Debug.Print("Exception when calling DefaultApi.WebInspectorWebInspectorPagesWithHttpInfo: " + e.Message);
    Debug.Print("Status Code: " + e.ErrorCode);
    Debug.Print(e.StackTrace);
}
```

### Parameters

| Name | Type | Description | Notes |
|------|------|-------------|-------|
| **udid** | **string** |  |  |

### Return type

**List<Object>**

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | The request has succeeded. |  -  |
| **401** | 401 — missing/invalid bearer token (when auth is enabled). |  -  |
| **404** | 404 — device (udid) not found. |  -  |
| **422** | 422 — empty/invalid udid. |  -  |
| **424** | 424 — a device-side prerequisite is missing. Used by the WebInspector routes when Web Inspector / Remote Automation is not enabled on the device. |  -  |
| **500** | 500 — internal error while talking to the device. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

