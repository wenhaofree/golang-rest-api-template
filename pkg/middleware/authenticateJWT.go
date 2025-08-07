package middleware

import (
	"golang-rest-api-template/pkg/auth"
	"golang-rest-api-template/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
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

		// 使用StandardClaims来匹配token生成时的结构
		claims := &jwt.StandardClaims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
			return auth.JwtKey, nil
		})

		if err != nil {
			response.Unauthorized(c, "Invalid token")
			c.Abort()
			return
		}

		if !token.Valid {
			response.Unauthorized(c, "Invalid token")
			c.Abort()
			return
		}

		// 从Issuer字段获取email（与token生成时保持一致）
		if claims.Issuer == "" {
			response.Unauthorized(c, "Invalid token: missing user information")
			c.Abort()
			return
		}

		c.Set("username", claims.Issuer)
		c.Next()
	}
}
