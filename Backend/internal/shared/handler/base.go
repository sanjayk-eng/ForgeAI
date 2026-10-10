package handler

import (
	stderrors "errors"
	"net/http"

	apierrors "ai-agent/internal/shared/errors"

	"github.com/gin-gonic/gin"
)

// BaseHandler provides common handler utilities
type BaseHandler struct{}

// WriteError writes a standard error response
func (h *BaseHandler) WriteError(c *gin.Context, status int, code, message string) {
	apierrors.Error(c, status, code, message, nil)
}

// WriteSuccess writes a standard success response
func (h *BaseHandler) WriteSuccess(c *gin.Context, status int, message string, data interface{}) {
	apierrors.Success(c, status, message, data)
}

// WriteServiceError handles common service errors and writes appropriate HTTP responses
func (h *BaseHandler) WriteServiceError(c *gin.Context, err error, errorMappings map[error]ErrorMapping) {
	// Check custom error mappings first
	for targetErr, mapping := range errorMappings {
		if stderrors.Is(err, targetErr) {
			apierrors.Error(c, mapping.Status, mapping.Code, mapping.Message, nil)
			return
		}
	}

	// Check for coded errors
	code, message := apierrors.CodeOf(err)
	var status int
	switch code {
	case apierrors.ErrCodeValidation:
		status = http.StatusBadRequest
	case apierrors.ErrCodeNotFound:
		status = http.StatusNotFound
	case apierrors.ErrCodeConflict:
		status = http.StatusConflict
	case apierrors.ErrCodeUnauthorized:
		status = http.StatusUnauthorized
	case apierrors.ErrCodeForbidden:
		status = http.StatusForbidden
	default:
		status = http.StatusInternalServerError
	}

	apierrors.Error(c, status, code, message, nil)
}

// ErrorMapping defines how an error should be mapped to an HTTP response
type ErrorMapping struct {
	Status  int
	Code    string
	Message string
}
