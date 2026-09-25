package errors

import (
	"errors"
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
	ErrCodeConflict       = "CONFLICT"
	ErrCodeInternalServer = "INTERNAL_SERVER_ERROR"
)

type CodedError struct {
	Code    string
	Message string
	Err     error
}

func (err *CodedError) Error() string {
	return err.Message
}

func (err *CodedError) Unwrap() error {
	return err.Err
}

func NewCodedError(code, message string, cause error) error {
	return &CodedError{Code: code, Message: message, Err: cause}
}

func CodeOf(err error) (string, string) {
	var coded *CodedError
	if errors.As(err, &coded) {
		return coded.Code, coded.Message
	}
	if err != nil && err.Error() != "" {
		return ErrCodeInternalServer, err.Error()
	}
	return ErrCodeInternalServer, "internal server error"
}

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
