# GoIos.Sdk.Generated.Api.DefaultApi

All URIs are relative to *http://localhost:60105*

| Method | HTTP request | Description |
|--------|--------------|-------------|
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
| [**ListDevices**](DefaultApi.md#listdevices) | **GET** /api/v1/list | List devices |
| [**ListTunnels**](DefaultApi.md#listtunnels) | **GET** /api/v1/tunnels | List tunnels |
| [**RefreshTunnel**](DefaultApi.md#refreshtunnel) | **POST** /api/v1/tunnels/{udid}/refresh | Refresh tunnel |
| [**ShutdownTunnelAgent**](DefaultApi.md#shutdowntunnelagent) | **POST** /api/v1/tunnel-agent/shutdown | Shut down tunnel agent |
| [**StopTunnel**](DefaultApi.md#stoptunnel) | **DELETE** /api/v1/tunnels/{udid} | Stop tunnel |

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
> Object DevicesGetLockdownValues (string udid)

Get lockdown values

Get all lockdown values (CLI: `ios lockdown get`).

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

            try
            {
                // Get lockdown values
                Object result = apiInstance.DevicesGetLockdownValues(udid);
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
    ApiResponse<Object> response = apiInstance.DevicesGetLockdownValuesWithHttpInfo(udid);
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

