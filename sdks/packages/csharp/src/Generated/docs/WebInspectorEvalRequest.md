# GoIos.Sdk.Generated.Model.WebInspectorEvalRequest
`POST /device/{udid}/webinspector/eval` request body.

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Page** | **string** | Inspectable page key. When empty the first matching web/javascript page (optionally scoped by &#x60;bundleId&#x60;) is used. | [optional] 
**BundleId** | **string** | Optional bundle id to scope page selection. | [optional] 
**Script** | **string** | JavaScript source to evaluate. Required. | 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

