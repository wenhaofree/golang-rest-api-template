package middleware

import (
	"net/http"
	"time"

	"golang-rest-api-template/pkg/apperrors"
	"golang-rest-api-template/pkg/response"

	"github.com/gin-gonic/gin"
)

// ErrorHandler is a global middleware that converts returned errors into a unified JSON
// and prevents leaking internal details in production.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		// If any handlers called c.Error(err), unify here
		if len(c.Errors) == 0 { return }

		// take last error
		err := c.Errors.Last().Err
		if ae, ok := apperrors.IsAppError(err); ok {
			response.ErrorWithData(c, ae.HTTPStatus, gin.H{"code": ae.Code}, ae.Message)
			return
		}
		// Fallback unexpected error
		response.ErrorWithData(c, http.StatusInternalServerError, gin.H{"code": apperrors.ErrInternal.Code}, apperrors.ErrInternal.Message)
		_ = start // placeholder for optional logging latency
	}
}

