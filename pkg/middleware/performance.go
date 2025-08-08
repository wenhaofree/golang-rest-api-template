package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PerformanceMonitor 性能监控中间件
func PerformanceMonitor(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// 处理请求
		c.Next()

		// 计算处理时间
		duration := time.Since(start)
		status := c.Writer.Status()

		// 记录性能指标
		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("duration", duration),
			zap.String("ip", c.ClientIP()),
		}
		if reqID, ok := c.Get("request_id"); ok {
			fields = append(fields, zap.String("request_id", reqID.(string)))
		}
		logger.Info("API Performance", fields...)

		// 如果请求处理时间超过阈值，记录警告
		if duration > 1*time.Second {
			logger.Warn("Slow API Request", fields...)
		}

		// 设置响应头，便于客户端监控
		c.Header("X-Response-Time", duration.String())
	}
}

// RequestSizeLimit 请求大小限制中间件
func RequestSizeLimit(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			c.JSON(413, gin.H{
				"code":    1,
				"message": "Request entity too large",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
