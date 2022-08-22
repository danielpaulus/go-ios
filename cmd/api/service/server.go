package service

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
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

func notFoundHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverError("not found", http.StatusNotFound, w)
	})
}

func methodNotAllowedHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverError("method not allowed", http.StatusMethodNotAllowed, w)
	})
}

//CreateRouter creates a new router and exposes the workspace to
//the http handlers.
func CreateRouter() *mux.Router {
	r := mux.NewRouter()
	r.MethodNotAllowedHandler = methodNotAllowedHandler()
	r.NotFoundHandler = notFoundHandler()
	r.HandleFunc("/health", limitNumClients(HealthHandler(), 1)).Methods("GET")
	r.HandleFunc("/runtest", XCTestHandler()).Methods("GET")

	return r
}

//CreateHTTPServer creates a *http.Server with routes added by the Createrouter func.
//It also configures timeouts, which is important because default timeouts are set to 0
//which can cause tcp connections being open indefinitely.
func CreateHTTPServer(address string) *http.Server {
	srv := &http.Server{
		Handler:      CreateRouter(),
		Addr:         address,
		WriteTimeout: 5 * time.Minute,
		ReadTimeout:  5 * time.Minute,
	}

	return srv
}
