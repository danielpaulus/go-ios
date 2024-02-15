// Package tunnel implements creating a TUN interface for iOS 17 devices. These tunnels are used to connect to services
// running on the device. On older iOS versions this was done over the 'usbmuxd'-socket. Services that are available via
// usbmuxd can also be used over the tunnel (they are listed in remote service discovery).
//
// Starting a tunnel is a two-step process:
//
// - Pair Device/Verify Device Pairing
//
// - Setting up a TUN interface
//
// # Device Pairing
//
// The step of device pairing means either that a new pairing is set up, or credentials from an already existing
// pairing are verified.
// For device pairing we connect to the RemoteXPC service 'com.apple.internal.dt.coredevice.untrusted.tunnelservice'
// directly on the ethernet interface exposed by the device.
// Process:
// - host	->	device	: 'setupManualPairing' request
//
// - device	->	host	: device public key and salt to initialize SRP
//
// - host	->	device	: host public key and SRP client proof
//
// - device	->	host	: SRP host proof
//
// - host	-> 	device	: host info (signed with the key from the host pair record) encrypted with the session key obtained through SRP
//
// - device	->	host	: device info encrypted with the same key as above
//
// Verifying a device pairing works by signing the host identifier (which is part of the pair record on the host)
// with a key derived through ECDH. The public keys used there are the key of the pair record on the host and the device
// provides a public key when the host asks to verify the pairing.
//
// # Tunnel Interface
//
// For the TUN interface we continue to use the same connection that we opened for the device pairing.
// The host asks the device to create a listener, and provides a public key, as well as the requested tunnel type.
// There are two types of tunnels that can be created, QUIC and TCP. In this package only QUIC is supported.
// The device will respond with a public key and a port number on which the tunnel endpoint listens on the device.
// The host opens a QUIC connection using a self-signed certificate with the public key that was sent to the device earlier.
// The first messages on this QUIC connection exchange the parameters for this tunnel. The host sends a 'clientHandshakeRequest'
// containing the MTU used for the TUN interface, and the device responds with the information about what IP addresses
// the host and the device have to use for this tunnel, as well as the port on which remote service discovery (RSD) is
// reachable on the device
//
// After that all services listed in RSD are available via this TUN interface
package tunnel
