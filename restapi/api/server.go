package api

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"net/http"
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

/*
func limitNumClientsUDID(f http.HandlerFunc) http.HandlerFunc {
	maxClients := 1
	semaMap := map[string]chan struct{}{}
	mux := sync.Mutex{}
	return func(w http.ResponseWriter, r *http.Request) {
		udid := strings.TrimSpace(r.URL.Query().Get("udid"))
		if udid == "" {
			serverError("missing udid", http.StatusBadRequest, w)
			return
		}
		mux.Lock()
		var sema chan struct{}
		sema, ok := semaMap[udid]
		if !ok {
			sema = make(chan struct{}, maxClients)
			semaMap[udid] = sema
		}
		mux.Unlock()
		sema <- struct{}{}
		defer func() { <-sema }()
		f(w, r)
	}
}
*/
func Main() {
	router := gin.Default()

	// Add event-streaming headers
	router.Use(HeadersMiddleware())

	v1 := router.Group("/api/v1", gin.BasicAuth(gin.Accounts{
		"admin": "admin123", // username : admin, password : admin123
	}))

	v1.GET("/info", Info)
	v1.GET("/shot", Screenshot)
	v1.GET("/listen", Listen)
	v1.GET("/list", List)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	err := router.Run(":8080")
	if err != nil {
		log.Error(err)
	}
}

func HeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")
		c.Next()
	}
}
