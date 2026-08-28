# GoIos.Sdk.Generated.Model.RunTestRequest
`POST /device/{udid}/jobs/runtest` (and `runwda`) request.

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**BundleId** | **string** | Bundle id of the app under test. | [optional] 
**TestRunnerBundleId** | **string** | Bundle id of the test runner. Defaults to &#x60;bundleId&#x60; if omitted. | [optional] 
**XctestConfig** | **string** | Name of the &#x60;.xctestconfiguration&#x60;. | [optional] 
**Env** | **Object** | Extra environment variables for the test runner. | [optional] 
**Args** | **List&lt;string&gt;** | Extra process arguments for the test runner. | [optional] 
**TestsToRun** | **List&lt;string&gt;** | Only run these tests (&#x60;Class/method&#x60; identifiers). | [optional] 
**TestsToSkip** | **List&lt;string&gt;** | Skip these tests. | [optional] 
**Xctest** | **bool** | Run as a plain XCTest (vs XCUITest). | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

