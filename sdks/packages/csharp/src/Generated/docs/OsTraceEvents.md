# GoIos.Sdk.Generated.Model.OsTraceEvents

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Pid** | **int** | Process id that emitted the entry. | [optional] 
**ProcessName** | **string** | Emitting process/executable name. | [optional] 
**Level** | **string** | Log level, e.g. &#x60;default&#x60;, &#x60;info&#x60;, &#x60;debug&#x60;, &#x60;error&#x60;, &#x60;fault&#x60;. | [optional] 
**Subsystem** | **string** | Subsystem string (e.g. &#x60;com.apple.network&#x60;). | [optional] 
**Category** | **string** | Category within the subsystem. | [optional] 
**Message** | **string** | The formatted log message. | 
**Timestamp** | **long** | Unix epoch milliseconds when the entry was emitted, if known. | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

