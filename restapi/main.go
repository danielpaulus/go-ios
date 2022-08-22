package main

import (
	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/restapi/api"
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {
	ios.Listen()
	log.WithFields(log.Fields{"args": os.Args, "version": api.GetVersion()}).Infof("starting go-iOS-API")

	const address = "0.0.0.0:16800"
	api.Main()
}
