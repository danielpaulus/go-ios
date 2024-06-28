package cmd

import (
	"fmt"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/mobileactivation"
	"github.com/danielpaulus/go-ios/ios/tunnel"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var activateCmd = &cobra.Command{
	Use:   "activate",
	Short: "Activate a device",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("activate called")

		udid, _ := arguments.String("--udid")
		address, addressErr := arguments.String("--address")
		rsdPort, rsdErr := arguments.Int("--rsd-port")

		device, err := ios.GetDevice(udid)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Fatalf(msg)
		}
		// device address and rsd port are only available after the tunnel started

		exitIfError("Device not found: "+udid, err)
		if addressErr == nil && rsdErr == nil {
			device = deviceWithRsdProvider(device, udid, address, rsdPort)
		} else {
			info, err := tunnel.TunnelInfoForDevice(device.Properties.SerialNumber, tunnelInfoPort)
			if err == nil {
				device = deviceWithRsdProvider(device, udid, info.Address, info.RsdPort)
			} else {
				log.WithField("udid", device.Properties.SerialNumber).Warn("failed to get tunnel info")
			}
		}

		exitIfError("failed activation", mobileactivation.Activate(device))
		return
	},
}

func init() {
	rootCmd.AddCommand(activateCmd)

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// activateCmd.PersistentFlags().String("foo", "", "A help for foo")

	activateCmd.Flags().String("udid", "", "udid of the device to activate")
	activateCmd.Flags().String("address", "", `
Address of the device on the interface.
This parameter is optional and can be set if a tunnel created by MacOS needs to be used.
`)
	activateCmd.Flags().Int("rsd-port", 0, "Port of remote service discovery on the device through the tunnel")
}
