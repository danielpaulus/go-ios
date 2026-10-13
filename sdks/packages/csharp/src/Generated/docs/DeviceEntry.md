# GoIos.Sdk.Generated.Model.DeviceEntry
A single device as returned by `GET /list`.

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DeviceID** | **int** |  | 
**MessageType** | **string** |  | [optional] 
**Properties** | [**DeviceProperties**](DeviceProperties.md) |  | 
**Address** | **string** | Network address for a device reached over the network / tunnel. | [optional] 
**UserspaceTUN** | **bool** | True if reachable via the userspace TUN tunnel. | [optional] 
**UserspaceTUNHost** | **string** |  | [optional] 
**UserspaceTUNPort** | **int** |  | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

