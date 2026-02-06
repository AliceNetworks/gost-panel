package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorCode 错误代码
type ErrorCode string

const (
	ErrCodeBadRequest     ErrorCode = "BAD_REQUEST"
	ErrCodeUnauthorized   ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden      ErrorCode = "FORBIDDEN"
	ErrCodeNotFound       ErrorCode = "NOT_FOUND"
	ErrCodeConflict       ErrorCode = "CONFLICT"
	ErrCodeValidation     ErrorCode = "VALIDATION_ERROR"
	ErrCodeInternal       ErrorCode = "INTERNAL_ERROR"
	ErrCodeQuotaExceeded  ErrorCode = "QUOTA_EXCEEDED"
	ErrCodeServiceUnavail ErrorCode = "SERVICE_UNAVAILABLE"
)

// APIError 统一错误响应结构
type APIError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details any       `json:"details,omitempty"`
}

// APIResponse 统一成功响应结构
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// respondError 返回错误响应
func respondError(c *gin.Context, status int, code ErrorCode, message string, details ...any) {
	err := APIError{
		Code:    code,
		Message: message,
	}
	if len(details) > 0 {
		err.Details = details[0]
	}
	c.JSON(status, gin.H{"error": err})
}

// respondSuccess 返回成功响应
func respondSuccess(c *gin.Context, data any) {
	c.JSON(http.StatusOK, data)
}

// respondCreated 返回创建成功响应
func respondCreated(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, data)
}

// respondMessage 返回简单消息响应
func respondMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: message,
	})
}

// 常用错误响应快捷方法

// badRequest 400 错误
func badRequest(c *gin.Context, message string) {
	respondError(c, http.StatusBadRequest, ErrCodeBadRequest, message)
}

// validationError 400 验证错误
func validationError(c *gin.Context, message string, details any) {
	respondError(c, http.StatusBadRequest, ErrCodeValidation, message, details)
}

// unauthorized 401 未授权
func unauthorized(c *gin.Context, message string) {
	respondError(c, http.StatusUnauthorized, ErrCodeUnauthorized, message)
}

// forbidden 403 禁止访问
func forbidden(c *gin.Context, message string) {
	respondError(c, http.StatusForbidden, ErrCodeForbidden, message)
}

// notFound 404 未找到
func notFound(c *gin.Context, resource string) {
	respondError(c, http.StatusNotFound, ErrCodeNotFound, resource+" not found")
}

// conflict 409 冲突
func conflict(c *gin.Context, message string) {
	respondError(c, http.StatusConflict, ErrCodeConflict, message)
}

// internalError 500 内部错误
func internalError(c *gin.Context, message string) {
	respondError(c, http.StatusInternalServerError, ErrCodeInternal, message)
}

// serviceUnavailable 503 服务不可用
func serviceUnavailable(c *gin.Context, message string) {
	respondError(c, http.StatusServiceUnavailable, ErrCodeServiceUnavail, message)
}
