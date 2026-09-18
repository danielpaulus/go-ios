# GoIos.Sdk.Generated.Model.ListenEvents

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Event** | **string** | Event kind. &#x60;attached&#x60; when a device connects, &#x60;detached&#x60; when it disconnects, &#x60;paired&#x60; when a pairing record appears. | 
**DeviceID** | **int** | usbmuxd device id. | [optional] 
**Udid** | **string** | The device udid (serial number), when known. | [optional] 
**Properties** | [**DeviceProperties**](DeviceProperties.md) | Full device properties, present on &#x60;attached&#x60;. | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

