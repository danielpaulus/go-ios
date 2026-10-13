package api

import (
	"testing"

	"github.com/gin-gonic/gin"
)

// TestRegisterRoutesNoConflict builds the full route tree. gin panics at
// registration on a route conflict (e.g. a static segment colliding with a
// wildcard), so this fails loudly if a new endpoint clashes with an existing one.
func TestRegisterRoutesNoConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("route registration panicked (conflict?): %v", r)
		}
	}()
	router := gin.New()
	v1 := router.Group("/api/v1")
	registerRoutes(v1, 0, 0)
}
