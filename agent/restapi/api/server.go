package api

import (
	"io"
	"os"

	"github.com/danielpaulus/go-ios/agent/devicestatemgmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const IOSDEVICEBRIDGE = "ios_device_Brige_handle"
const ANDROIDDEVICEBRIDGE = "android_device_Brige_handle"
const DEVICE_LIST = "devicelist_go_devicepool"

func Main(list *devicestatemgmt.DeviceList) {
	router := gin.Default()
	router.Use(func(context *gin.Context) {
		context.Set(DEVICE_LIST, list)
		context.Next()
	})
	log := logrus.New()
	myfile, _ := os.Create("go-ios.log")
	gin.DefaultWriter = io.MultiWriter(myfile, os.Stdout)
	router.Use(MyLogger(log), gin.Recovery())

	v1 := router.Group("/api/v1")
	registerRoutes(v1)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	err := router.Run(":" + os.Getenv("HTTP_PORT"))
	if err != nil {
		log.Error(err)
	}
}
