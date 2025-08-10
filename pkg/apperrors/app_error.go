package apperrors

import (
	"errors"
	"fmt"
	"net/http"
)

type AppError struct {
	Code       string // domain-specific code, e.g. USER_NOT_FOUND
	Message    string // client-facing message
	HTTPStatus int    // http status to return
	Cause      error  // internal cause, not exposed
}

func (e *AppError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.Cause }

// Builders
func New(code, message string, httpStatus int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: httpStatus}
}

func Wrap(code, message string, httpStatus int, cause error) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: httpStatus, Cause: cause}
}

// Common helpers
var (
	ErrBadRequest   = New("BAD_REQUEST", "Bad request", http.StatusBadRequest)
	ErrUnauthorized = New("UNAUTHORIZED", "Unauthorized", http.StatusUnauthorized)
	ErrForbidden    = New("FORBIDDEN", "Forbidden", http.StatusForbidden)
	ErrNotFound     = New("NOT_FOUND", "Not found", http.StatusNotFound)
	ErrConflict     = New("CONFLICT", "Conflict", http.StatusConflict)
	ErrInternal     = New("INTERNAL", "Internal server error", http.StatusInternalServerError)
)

func IsAppError(err error) (*AppError, bool) {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}

// ClassifyCode 返回标准错误分类码，如 APP-400/APP-401/APP-500
func ClassifyCode(httpStatus int) string {
	switch httpStatus {
	case http.StatusBadRequest:
		return "APP-400"
	case http.StatusUnauthorized:
		return "APP-401"
	case http.StatusForbidden:
		return "APP-403"
	case http.StatusNotFound:
		return "APP-404"
	case http.StatusConflict:
		return "APP-409"
	case http.StatusTooManyRequests:
		return "APP-429"
	default:
		if httpStatus >= 500 {
			return "APP-500"
		}
		return fmt.Sprintf("APP-%d", httpStatus)
	}
}
