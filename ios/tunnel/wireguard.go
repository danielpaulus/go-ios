package tunnel

import (
	"fmt"

	"golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/windows/tunnel"
)

func Tmain() {
	// Create a TUN device
	tunDevice, err := tun.CreateTUN("tun0", 1500)
	if err != nil {
		fmt.Println("Error creating TUN device:", err)
		return
	}

	// Create a new WireGuard device
	device, err := tunnel.CreateTUNWithRequestedGUID("wg0", 1500, nil)
	if err != nil {
		fmt.Println("Error creating WireGuard device:", err)
		return
	}

	// Use the TUN device and WireGuard device as needed...
	fmt.Println("TUN device and WireGuard device created successfully")
}
