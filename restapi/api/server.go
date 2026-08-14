package api

import (
	"context"
	"flag"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// serverConfig holds the command-line configuration for the REST server.
type serverConfig struct {
	addr        string
	disableAuth bool
	tlsCert     string
	tlsKey      string
	rateLimit   float64
	rateBurst   int
}

// parseServerConfig parses the server flags from args. It uses a dedicated flag
// set with ContinueOnError and discarded output so unknown/extra args never
// crash startup.
func parseServerConfig(args []string) serverConfig {
	fs := flag.NewFlagSet("go-ios-restapi", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	addr := fs.String("addr", ":8080", "address to listen on (host:port)")
	disableAuth := fs.Bool("disable-auth", false, "run the REST API without authentication")
	tlsCert := fs.String("tls-cert", "", "path to a TLS certificate; enables HTTPS together with --tls-key")
	tlsKey := fs.String("tls-key", "", "path to the TLS private key for --tls-cert")
	rateLimit := fs.Float64("rate-limit", 20, "max sustained requests per second per device (0 disables)")
	rateBurst := fs.Int("rate-burst", 40, "burst size for the per-device rate limit")
	// Ignore parse errors (e.g. unknown flags) so extra args don't crash startup.
	_ = fs.Parse(args)
	return serverConfig{
		addr: *addr, disableAuth: *disableAuth, tlsCert: *tlsCert, tlsKey: *tlsKey,
		rateLimit: *rateLimit, rateBurst: *rateBurst,
	}
}

func Main() {
	router := gin.Default()
	log := logrus.New()
	myfile, _ := os.Create("go-ios.log")
	gin.DefaultWriter = io.MultiWriter(myfile, os.Stdout)
	router.Use(MyLogger(log), gin.Recovery())

	cfg := parseServerConfig(os.Args[1:])

	// Liveness/readiness probes are unauthenticated and live outside /api/v1.
	RegisterHealthRoutes(router)

	// Authentication configuration. The API exposes full device control, so it
	// must not run unauthenticated by accident: either a token is supplied via
	// GO_IOS_API_KEY, or auth is explicitly disabled with --disable-auth.
	token := os.Getenv("GO_IOS_API_KEY")
	authEnabled := token != "" && !cfg.disableAuth

	v1 := router.Group("/api/v1")
	switch {
	case cfg.disableAuth:
		log.Warn("go-ios REST API is running WITHOUT authentication")
	case token != "":
		v1.Use(BearerAuth(token))
	default:
		log.Fatal("go-ios REST API requires authentication: set GO_IOS_API_KEY=<token>, " +
			"or pass --disable-auth to run without authentication")
	}

	registerRoutes(v1, cfg.rateLimit, cfg.rateBurst)

	// Serve the swagger UI. When auth is enabled, gate it behind the token too
	// (under /api/v1) so the API schema isn't exposed unauthenticated; otherwise
	// keep it at the historical /swagger path.
	if authEnabled {
		v1.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	} else {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	srv := &http.Server{
		Addr:              cfg.addr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
		// No ReadTimeout/WriteTimeout on purpose: the /syslog, /listen, /ostrace
		// and /notifications endpoints stream for the lifetime of the connection,
		// and a WriteTimeout would sever them. ReadHeaderTimeout + IdleTimeout
		// still bound slow-header and idle-keepalive abuse.
	}

	// Graceful shutdown on SIGINT/SIGTERM so in-flight requests can drain.
	shutdownDone := make(chan struct{})
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs
		log.Info("shutting down go-ios REST API")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Errorf("graceful shutdown failed: %v", err)
		}
		close(shutdownDone)
	}()

	var err error
	if cfg.tlsCert != "" && cfg.tlsKey != "" {
		log.Infof("go-ios REST API listening on %s (TLS)", cfg.addr)
		err = srv.ListenAndServeTLS(cfg.tlsCert, cfg.tlsKey)
	} else {
		log.Infof("go-ios REST API listening on %s", cfg.addr)
		err = srv.ListenAndServe()
	}
	if err != nil && err != http.ErrServerClosed {
		log.Error(err)
		return
	}
	<-shutdownDone
}
