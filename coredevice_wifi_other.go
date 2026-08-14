//go:build !darwin

package main

import (
	"context"

	"github.com/danielpaulus/go-ios/ios/discovery"
)

func discoverCoreDeviceWifiPairing(ctx context.Context) []discovery.Device {
	return nil
}
