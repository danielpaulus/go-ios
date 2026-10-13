# GoIos.Sdk.Generated.Model.GenericResponse
The dominant response envelope used across the API. Success responses set `message`; error responses set `error`. Streaming/middleware paths that emit `gin.H{\"error\"|\"message\"}` are compatible with this shape.

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Message** | **string** | Human-readable success or status message. | [optional] 
**Error** | **string** | Human-readable error message. Present on failures. | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

