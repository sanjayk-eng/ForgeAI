package agent

import (
	"errors"
	"net/http"

	"ai-agent/internal/middleware"
	"ai-agent/internal/modules/terminal/domain"
	apierrors "ai-agent/internal/shared/errors"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (handler *Handler) Status(c *gin.Context) {
	status, err := handler.service.Status(c.Request.Context(), requestUserID(c), c.Param("sandbox_id"))
	if err != nil {
		handler.writeError(c, err)
		return
	}
	apierrors.Success(c, http.StatusOK, "agent status fetched", status)
}

func (handler *Handler) RunTask(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 256*1024)
	var request TaskRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "prompt is required", nil)
		return
	}
	result, err := handler.service.RunTask(c.Request.Context(), requestUserID(c), c.Param("sandbox_id"), request.Prompt)
	if err != nil {
		handler.writeError(c, err)
		return
	}
	apierrors.Success(c, http.StatusOK, "agent task completed", result)
}

func (handler *Handler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotConfigured):
		apierrors.Error(c, http.StatusServiceUnavailable, "AI_NOT_CONFIGURED", "AI model is not configured on the server", nil)
	case errors.Is(err, ErrInvalidPrompt):
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error(), nil)
	case errors.Is(err, domain.ErrSandboxNotFound):
		apierrors.Error(c, http.StatusNotFound, apierrors.ErrCodeNotFound, "sandbox not found", nil)
	case errors.Is(err, domain.ErrSandboxAccessDenied):
		apierrors.Error(c, http.StatusForbidden, apierrors.ErrCodeForbidden, "workspace access denied", nil)
	case errors.Is(err, ErrModelRequest), errors.Is(err, ErrModelResponse):
		apierrors.Error(c, http.StatusBadGateway, "AI_PROVIDER_ERROR", "AI provider could not complete this task", nil)
	default:
		apierrors.Error(c, http.StatusInternalServerError, apierrors.ErrCodeInternalServer, "agent task failed", nil)
	}
}

func requestUserID(c *gin.Context) string {
	value, _ := middleware.UserID(c)
	return value
}
