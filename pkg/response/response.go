package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

// ErrorEnvelope 统一错误响应结构
// 示例: {"requestId":"...","code":"APP-400","message":"Bad request","details":{...}}
type ErrorEnvelope struct {
	RequestID string      `json:"requestId"`
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	Details   interface{} `json:"details,omitempty"`
}

const (
	SUCCESS = 0
	ERROR   = 1
)

// 成功响应保持不变，避免破坏已有客户端
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    SUCCESS,
		Data:    data,
		Message: "success",
	})
}

func SuccessWithMessage(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    SUCCESS,
		Data:    data,
		Message: message,
	})
}

// 兼容旧错误函数（不含 requestId），建议统一使用 ErrorJSON
func Error(c *gin.Context, httpStatus int, message string) {
	c.JSON(httpStatus, Response{
		Code:    ERROR,
		Data:    nil,
		Message: message,
	})
}

func ErrorWithData(c *gin.Context, httpStatus int, data interface{}, message string) {
	c.JSON(httpStatus, Response{
		Code:    ERROR,
		Data:    data,
		Message: message,
	})
}

// ErrorJSON 输出统一错误结构（包含 requestId, code, message, details）
func ErrorJSON(c *gin.Context, httpStatus int, appCode string, message string, details interface{}) {
	reqID := c.GetString("request_id")
	if reqID == "" {
		reqID = c.GetHeader("X-Request-ID")
	}
	env := ErrorEnvelope{RequestID: reqID, Code: appCode, Message: message, Details: details}
	c.JSON(httpStatus, env)
}

func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message)
}

func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}
