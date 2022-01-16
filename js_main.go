package main

import (
	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/screenshotr"
	"github.com/gopherjs/gopherjs/js"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetOutput(consoleWriter{})
	//log.SetLevel(log.TraceLevel)
	printDeviceList()

}

type consoleWriter struct {
}

func (c consoleWriter) Write(msg []byte) (int, error) {
	js.Global.Get("console").Call("log", string(msg))
	return len(msg), nil
}

func printDeviceList() {
	js.Global.Get("console").Call("log", "run list")
	deviceList, err := ios.ListDevices()
	if err != nil {
		js.Global.Get("console").Call("log", err)
		return
	}

	for _, device := range deviceList.DeviceList {
		udid := device.Properties.SerialNumber
		allValues, err := ios.GetValues(device)
		if err != nil {
			log.Error(err)
			return
		}
		log.Info(allValues)
		log.Info(udid)
	}

	js.Global.Get("console").Call("log", deviceList)
	device := deviceList.DeviceList[0]
	log.Info("starting shotr")
	screenshotrService, err := screenshotr.New(device)
	log.Error("Starting Screenshotr failed with", err)


	imageBytes, err := screenshotrService.TakeScreenshot()
	log.Error("screenshotr failed", err)

	log.Info(imageBytes)
	//	log.Info(js.Global.Get("JSON").Call("stringify", ws).String())
}
