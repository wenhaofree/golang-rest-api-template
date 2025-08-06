package middleware

import (
	"golang-rest-api-template/pkg/response"
	"os"

	"github.com/gin-gonic/gin"
)

func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == os.Getenv("API_SECRET_KEY") {
			c.Next()
		} else {
			response.Unauthorized(c, "Unauthorized")
			c.Abort()
		}
	}
}
