package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterHealthRoutes registers unauthenticated liveness and readiness probes.
// These are intentionally outside the /api/v1 auth group so orchestrators
// (Kubernetes, load balancers, systemd) can health-check without a token.
func RegisterHealthRoutes(router *gin.Engine) {
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/readyz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
}
