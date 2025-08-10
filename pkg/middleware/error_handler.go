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

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		// 优先处理 AppError，输出标准错误包裹（含 requestId）
		if ae, ok := apperrors.IsAppError(err); ok {
			// 将 domain code 包裹为分类码（APP-xxx），便于前端与监控统一
			classCode := apperrors.ClassifyCode(ae.HTTPStatus)
			response.ErrorJSON(c, ae.HTTPStatus, classCode, ae.Message, gin.H{"code": ae.Code})
			return
		}
		// 兜底未知错误
		response.ErrorJSON(c, http.StatusInternalServerError, apperrors.ClassifyCode(http.StatusInternalServerError), apperrors.ErrInternal.Message, nil)
		_ = start
	}
}
