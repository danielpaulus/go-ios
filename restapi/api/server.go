package api

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"io"
	"net/http"
	"os"
)

//limitNumClients uses a go channel to rate limit a handler.
// It is a golang buffered channel so you can put maxClients empty structs
//into the channel without blocking. The maxClients+1 invocation will block
//until another handler finishes and removes one empty struct from the channel.
func limitNumClients(f http.HandlerFunc, maxClients int) http.HandlerFunc {
	sema := make(chan struct{}, maxClients)

	return func(w http.ResponseWriter, req *http.Request) {
		sema <- struct{}{}
		defer func() { <-sema }()
		f(w, req)
	}
}

func Main() {
	router := gin.Default()
	log := logrus.New()
	myfile, _ := os.Create("go-ios.log")
	gin.DefaultWriter = io.MultiWriter(myfile, os.Stdout)
	// Add event-streaming headers
	router.Use(MyLogger(log), gin.Recovery())

	/*v1 := router.Group("/api/v1", gin.BasicAuth(gin.Accounts{
		"admin": "admin123", // username : admin, password : admin123
	}))*/
	v1 := router.Group("/api/v1")
	registerDeviceSpecificEndpoints(v1)

	initStreamingResponseRoutes(v1)
	v1.GET("/list", List)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	err := router.Run(":8080")
	if err != nil {
		log.Error(err)
	}
}
