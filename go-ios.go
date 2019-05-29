package main

import (
	"usbmuxd/usbmux"
)

func main() {
	printDeviceList()
}

func printDeviceList() {
	deviceList := usbmux.ListDevices()

	deviceList.Print()

}
