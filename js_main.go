package main

import (
	"github.com/danielpaulus/go-ios/ios"
)

func main() {
	printDeviceList()
}

func printDeviceList() {
	println("run list")
	deviceList, err := ios.ListDevices()
	println(err)
	println(deviceList)
}
