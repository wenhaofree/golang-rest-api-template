package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func Logger(logger *zap.Logger, collection *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// End timer
		duration := time.Since(start)

		// Log the request details to structured logger
		logger.Info("Request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
		)

		// 只有当MongoDB collection不为nil时才记录到MongoDB
		if collection != nil {
			logEntry := bson.M{
				"timestamp":  time.Now(),
				"method":     c.Request.Method,
				"path":       c.Request.URL.Path,
				"status":     c.Writer.Status(),
				"duration":   duration.Milliseconds(), // 转换为毫秒便于查询
				"ip":         c.ClientIP(),
				"user-agent": c.Request.UserAgent(),
				"errors":     c.Errors.ByType(gin.ErrorTypePrivate).String(),
			}

			// Log to MongoDB (异步处理，避免阻塞请求)
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				
				if _, err := collection.InsertOne(ctx, logEntry); err != nil {
					logger.Error("Failed to log to MongoDB", zap.Error(err))
				}
			}()
		}
	}
}
