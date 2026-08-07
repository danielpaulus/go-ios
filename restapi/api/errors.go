package api

import (
	"github.com/gin-gonic/gin"
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
