package api

import (
	"flag"
	"io"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Main() {
	router := gin.Default()
	log := logrus.New()
	myfile, _ := os.Create("go-ios.log")
	gin.DefaultWriter = io.MultiWriter(myfile, os.Stdout)
	router.Use(MyLogger(log), gin.Recovery())

	// Authentication configuration. The API binds :8080 with full device control,
	// so it must not run unauthenticated by accident: either a token is supplied
	// via GO_IOS_API_KEY, or auth is explicitly disabled with --disable-auth.
	token := os.Getenv("GO_IOS_API_KEY")
	disableAuth := parseDisableAuth(os.Args[1:])

	v1 := router.Group("/api/v1")

	switch {
	case disableAuth:
		log.Warn("go-ios REST API is running WITHOUT authentication")
	case token != "":
		v1.Use(BearerAuth(token))
	default:
		log.Fatal("go-ios REST API requires authentication: set GO_IOS_API_KEY=<token>, " +
			"or pass --disable-auth to run without authentication")
	}

	registerRoutes(v1)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	err := router.Run(":8080")
	if err != nil {
		log.Error(err)
	}
}

// parseDisableAuth reports whether the --disable-auth flag was passed. It uses a
// dedicated flag set so it never clashes with the global flag set and tolerates
// unknown args gracefully.
func parseDisableAuth(args []string) bool {
	fs := flag.NewFlagSet("go-ios-restapi", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	disableAuth := fs.Bool("disable-auth", false, "run the REST API without authentication")
	// Ignore parse errors (e.g. unknown flags) so extra args don't crash startup.
	_ = fs.Parse(args)
	return *disableAuth
}
