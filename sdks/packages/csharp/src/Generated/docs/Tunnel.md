# GoIos.Sdk.Generated.Model.Tunnel
A running device tunnel as reported by the tunnel agent (`GET /tunnels`, `POST /tunnels/{udid}/refresh`). Mirrors `tunnel.Tunnel`.

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Udid** | **string** | The device udid this tunnel serves. | 
**Address** | **string** | Tunnel address (IPv6) reachable for RemoteXPC/RSD. | 
**RsdPort** | **int** | RemoteServiceDiscovery port on the tunnel. | 
**UserspaceTUN** | **bool** | Whether this tunnel is a userspace TUN. | [optional] 
**UserspaceTUNPort** | **int** | Userspace TUN port, when &#x60;UserspaceTUN&#x60; is true. | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

