package main

import (
	"github.com/danielpaulus/go-ios/restapi/api"
	_ "github.com/danielpaulus/go-ios/restapi/docs"
	log "github.com/sirupsen/logrus"
	"os"
)

// @title           Swagger Example API
// @version         1.0
// @description     This is a sample server celler server.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth
func main() {
	log.WithFields(log.Fields{"args": os.Args, "version": api.GetVersion()}).Infof("starting go-iOS-API")

	const address = "0.0.0.0:16800"
	api.Main()
}
