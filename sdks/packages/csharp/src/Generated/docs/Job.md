# GoIos.Sdk.Generated.Model.Job
A long-running operation started via the REST API (test run, WDA runner, port forward). Mirrors the server's `jobView`.

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | Opaque job id, e.g. &#x60;runtest-3&#x60;. | 
**Kind** | **string** | Job kind: &#x60;runtest&#x60;, &#x60;runwda&#x60; or &#x60;forward&#x60;. | 
**Udid** | **string** | The device udid the job runs on. | 
**Status** | [**JobStatus**](JobStatus.md) |  | 
**StartedAt** | **DateTimeOffset** | When the job started (ISO-8601). | 
**FinishedAt** | **DateTimeOffset** | When the job reached a terminal state (absent while running). | [optional] 
**Error** | **string** | Error message when &#x60;status&#x60; is &#x60;failed&#x60;. | [optional] 
**Result** | **Object** |  | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

