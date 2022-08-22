package api

import (
	"encoding/json"
	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/screenshotr"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
	"sync"
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

func Main() {
	router := gin.Default()

	// Add event-streaming headers
	router.Use(HeadersMiddleware())

	// Basic Authentication
	authorized := router.Group("/", gin.BasicAuth(gin.Accounts{
		"admin": "admin123", // username : admin, password : admin123
	}))

	authorized.GET("/shot", func(c *gin.Context) {

		dev, _ := ios.GetDevice("")
		conn, err := screenshotr.New(dev)
		log.Error(err)
		b, _ := conn.TakeScreenshot()

		c.Header("Content-Type", "image/png")
		c.Data(http.StatusOK, "application/octet-stream", b)
	})

	// Authorized client can stream the event
	authorized.GET("/stream", func(c *gin.Context) {
		// We are streaming current time to clients in the interval 10 seconds
		log.Info("connect")
		a, _, _ := ios.Listen()
		c.Stream(func(w io.Writer) bool {
			l, _ := a()
			// Stream message to client from message channel
			w.Write([]byte(MustMarshal(l)))
			return true
		})

	})

	//Parse Static files
	router.StaticFile("/", "./public/index.html")

	router.Run(":8085")
}

func MustMarshal(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
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
