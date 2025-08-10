package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestTimeout sets a timeout on the request context so downstream calls (DB/Redis)
// that use c.Request.Context() will be canceled when the timeout elapses.
// If timeoutMs <= 0, no timeout is applied.
func RequestTimeout(timeoutMs int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if timeoutMs <= 0 {
			c.Next()
			return
		}
		d := time.Duration(timeoutMs) * time.Millisecond
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()
		r := c.Request.Clone(ctx)
		c.Request = r
		c.Next()
	}
}
