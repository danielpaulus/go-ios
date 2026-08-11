# GoIos.Sdk.Generated.Model.UIAPIRequest
`POST /device/{udid}/ui/api` request — raw passthrough to the backend (`uidriver.APIRequest`). For WDA supply `method`/`path`/`body`; for DeviceKit supply `rpcMethod`/`rpcParams`.

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Method** | **string** | HTTP method for a WDA passthrough (defaults to GET). | [optional] 
**Path** | **string** | HTTP path for a WDA passthrough (required for the wda backend). | [optional] 
**Body** | **string** | Raw HTTP request body for a WDA passthrough (base64 bytes on the wire). | [optional] 
**RpcMethod** | **string** | JSON-RPC method name for a DeviceKit passthrough. | [optional] 
**RpcParams** | **Object** |  | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

