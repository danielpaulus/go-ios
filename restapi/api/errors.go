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

	errMissingDomain        = errors.New("missing required query param: domain (app|app-group|crash|temp)")
	errUnknownDomain        = errors.New("unknown domain; expected app|app-group|crash|temp")
	errMissingRemote        = errors.New("missing required query param: remote")
	errUnknownContentLength = errors.New("a Content-Length header is required for upload")
	errMissingCrashArgs     = errors.New("both 'cwd' and 'pattern' query params are required")
	errMissingProfile       = errors.New("missing profile payload (raw body or multipart 'profile' field)")
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
