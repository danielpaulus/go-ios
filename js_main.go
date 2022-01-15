package main

import (
	"github.com/danielpaulus/go-ios/ios"
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
	js.Global.Get("console").Call("log", deviceList)
	//	log.Info(js.Global.Get("JSON").Call("stringify", ws).String())
}
