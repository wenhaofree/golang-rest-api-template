package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang-rest-api-template/pkg/auth"
	"golang-rest-api-template/pkg/response"
	"strings"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		const BearerSchema = "Bearer "
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Unauthorized(c, "Missing Authorization Header")
			c.Abort()
			return
		}

		if !strings.HasPrefix(header, BearerSchema) {
			response.Unauthorized(c, "Invalid Authorization Header")
			c.Abort()
			return
		}

		tokenStr := header[len(BearerSchema):]
		claims := &auth.Claims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return auth.JwtKey, nil
		})

		println("error:", err.Error())
		println("tokenStr:", tokenStr)

		if err != nil {
			//response.Unauthorized(c, "Invalid token")
			response.Unauthorized(c, err.Error())
			c.Abort()
			return
		}

		if !token.Valid {
			response.Unauthorized(c, "Invalid token")
			c.Abort()
			return
		}

		c.Set("username", claims.Username)
		c.Next()
	}
}
