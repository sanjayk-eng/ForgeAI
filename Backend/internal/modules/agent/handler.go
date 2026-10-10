package agent

import (
	"net/http"

	"ai-agent/internal/middleware"
	"ai-agent/internal/modules/terminal/domain"
	basehandler "ai-agent/internal/shared/handler"
	apierrors "ai-agent/internal/shared/errors"
	"ai-agent/pkg/validate"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	basehandler.BaseHandler
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (handler *Handler) Status(c *gin.Context) {
	userID, ok := middleware.ExtractUserID(c)
	if !ok {
		return
	}
	
	status, err := handler.service.Status(c.Request.Context(), userID, c.Param("sandbox_id"))
	if err != nil {
		handler.writeError(c, err)
		return
	}
	apierrors.Success(c, http.StatusOK, "agent status fetched", status)
}

func (handler *Handler) RunTask(c *gin.Context) {
	userID, ok := middleware.ExtractUserID(c)
	if !ok {
		return
	}
	
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 256*1024)
	var request TaskRequest
	if err := validate.BindAndValidate(c, &request); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error(), nil)
		return
	}
	
	result, err := handler.service.RunTask(c.Request.Context(), userID, c.Param("sandbox_id"), request.Prompt)
	if err != nil {
		handler.writeError(c, err)
		return
	}
	apierrors.Success(c, http.StatusOK, "agent task completed", result)
}

func (handler *Handler) writeError(c *gin.Context, err error) {
	errorMappings := map[error]basehandler.ErrorMapping{
		ErrNotConfigured:               {http.StatusServiceUnavailable, "AI_NOT_CONFIGURED", "AI model is not configured on the server"},
		ErrInvalidPrompt:               {http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error()},
		domain.ErrSandboxNotFound:      {http.StatusNotFound, apierrors.ErrCodeNotFound, "sandbox not found"},
		domain.ErrSandboxAccessDenied:  {http.StatusForbidden, apierrors.ErrCodeForbidden, "workspace access denied"},
		ErrModelRequest:                {http.StatusBadGateway, "AI_PROVIDER_ERROR", "AI provider could not complete this task"},
		ErrModelResponse:               {http.StatusBadGateway, "AI_PROVIDER_ERROR", "AI provider could not complete this task"},
	}
	
	handler.WriteServiceError(c, err, errorMappings)
}
