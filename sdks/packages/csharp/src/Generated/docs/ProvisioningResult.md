# GoIos.Sdk.Generated.Model.ProvisioningResult
`POST /sign/provision` — provisioning assets envelope. The mobileprovision (and optionally the P12) are base64-encoded so one JSON response can carry both binary artifacts. Host-scoped (device-free).

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**BundleId** | **string** | The app bundle identifier registered with App Store Connect. | 
**CertificateId** | **string** | The signing certificate resource id. | 
**MobileprovisionBase64** | **string** | The &#x60;.mobileprovision&#x60;, base64-encoded. | 
**P12Base64** | **string** | The generated &#x60;.p12&#x60;, base64-encoded (absent when reusing a certificate). | [optional] 
**P12Password** | **string** | The password protecting &#x60;p12Base64&#x60;, echoed back (client-supplied). | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

