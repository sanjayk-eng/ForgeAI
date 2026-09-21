package errors

import (
	"time"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

type ErrorInfo struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

const (
	ErrCodeBadRequest     = "BAD_REQUEST"
	ErrCodeUnauthorized   = "UNAUTHORIZED"
	ErrCodeForbidden      = "FORBIDDEN"
	ErrCodeNotFound       = "NOT_FOUND"
	ErrCodeValidation     = "VALIDATION_ERROR"
	ErrCodeInternalServer = "INTERNAL_SERVER_ERROR"
)

func Success(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, APIResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	})
}

func Error(c *gin.Context, status int, code, message string, details map[string]string) {
	c.JSON(status, APIResponse{
		Success:   false,
		Error:     &ErrorInfo{Code: code, Message: message, Details: details},
		Timestamp: time.Now(),
	})
}
