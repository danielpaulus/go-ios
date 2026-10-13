# GoIos.Sdk.Generated.Model.WdaConfig
Configuration for launching a WebDriverAgent (XCUITest) runner session.

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**BundleId** | **string** | Bundle id of the WDA runner host app (e.g. &#x60;com.facebook.WebDriverAgentRunner.xctrunner&#x60;). | 
**TestBundleId** | **string** | Bundle id of the XCTest test bundle. | 
**XcTestConfig** | **string** | Path/name of the &#x60;.xctestconfiguration&#x60; to use. | 
**Args** | **List&lt;string&gt;** | Extra process arguments passed to the runner. | [optional] 
**Env** | **Object** | Extra environment variables passed to the runner. | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

