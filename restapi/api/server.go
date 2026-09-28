package api

import (
	"context"
	"crypto/tls"
	"flag"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danielpaulus/go-ios/ios/golog"
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
	addr := fs.String("addr", "127.0.0.1:0", "address to listen on (host:port); the default binds an ephemeral loopback port and publishes it to the discovery file. Pin/expose with e.g. :8080 or 0.0.0.0:9000")
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
		// No Addr: we bind the listener ourselves (below) via net.Listen so the
		// real OS-assigned port is knowable, then srv.Serve(ln). Addr is unused by
		// Serve and would misleadingly read 127.0.0.1:0 for ephemeral binds.
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
		// No ReadTimeout/WriteTimeout on purpose: the /syslog, /listen, /ostrace
		// and /notifications endpoints stream for the lifetime of the connection,
		// and a WriteTimeout would sever them. ReadHeaderTimeout + IdleTimeout
		// still bound slow-header and idle-keepalive abuse.
	}

	useTLS := cfg.tlsCert != "" && cfg.tlsKey != ""

	// Bind explicitly so the OS-assigned port (when --addr uses :0) is knowable
	// before we start serving: we need the real host:port to log it and to
	// publish the discovery file for SDK auto-discovery.
	ln, err := net.Listen("tcp", cfg.addr)
	if err != nil {
		golog.Error("failed to bind go-ios REST API", "module", logModule, "addr", cfg.addr, "error", err.Error())
		return
	}
	if useTLS {
		cert, certErr := tls.LoadX509KeyPair(cfg.tlsCert, cfg.tlsKey)
		if certErr != nil {
			golog.Error("failed to load TLS key pair", "module", logModule, "cert", cfg.tlsCert, "error", certErr.Error())
			_ = ln.Close()
			return
		}
		ln = tls.NewListener(ln, &tls.Config{Certificates: []tls.Certificate{cert}})
	}

	// The bound TCP address carries the real, possibly OS-assigned port (even for
	// the TLS listener, which wraps the TCP listener above).
	tcpAddr := ln.Addr().(*net.TCPAddr)
	host := tcpAddr.IP.String()
	port := tcpAddr.Port

	// Publish the discovery file so SDKs can auto-find this daemon regardless of
	// which (ephemeral or pinned) port it landed on.
	info := newDiscoveryInfo(host, port, useTLS)
	discoveryPath, derr := writeDiscoveryFile(info)
	if derr != nil {
		// Non-fatal: the daemon still serves; SDKs just can't auto-discover it.
		golog.Warn("failed to write REST API discovery file", "module", logModule, "error", derr.Error())
	} else {
		golog.Info("wrote REST API discovery file", "module", logModule, "path", discoveryPath, "baseUrl", info.BaseUrl)
	}

	golog.Info("go-ios REST API listening", "module", logModule, "baseUrl", info.BaseUrl, "host", host, "port", port, "tls", useTLS, "pid", os.Getpid())

	// Graceful shutdown on SIGINT/SIGTERM so in-flight requests can drain, and
	// remove the discovery file so stale entries don't linger after exit.
	shutdownDone := make(chan struct{})
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs
		golog.Info("shutting down go-ios REST API", "module", logModule)
		if rmErr := removeDiscoveryFile(); rmErr != nil {
			golog.Warn("failed to remove REST API discovery file", "module", logModule, "error", rmErr.Error())
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			golog.Error("graceful shutdown failed", "module", logModule, "error", err.Error())
		}
		close(shutdownDone)
	}()

	if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
		golog.Error("go-ios REST API serve error", "module", logModule, "error", err.Error())
		// Best-effort cleanup: on a serve failure the shutdown goroutine may never
		// fire, so drop the discovery file here too.
		_ = removeDiscoveryFile()
		return
	}
	<-shutdownDone
}
