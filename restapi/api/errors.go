package api

import (
	"errors"

	"github.com/gin-gonic/gin"
)

// Sentinel errors for common request-validation failures.
var (
	errMissingKey        = errors.New("missing required query param: key")
	errEraseNotConfirmed = errors.New("erase is destructive; pass ?confirm=true to proceed")
	errUnknownAction     = errors.New("unknown action")
	errMissingProcess    = errors.New("missing required 'process' (query param or JSON body)")
)

// RespondError writes a consistent JSON error envelope ({"error": "..."}) and
// aborts the request with the given status. Handlers should use this instead of
// ad-hoc error responses so clients get a uniform error shape.
func RespondError(c *gin.Context, status int, err error) {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	c.AbortWithStatusJSON(status, gin.H{"error": msg})
}
