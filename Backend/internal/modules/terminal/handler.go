package terminal

import (
	"errors"
	"net/http"

	"ai-agent/internal/middleware"
	"ai-agent/internal/modules/terminal/domain"
	apierrors "ai-agent/internal/shared/errors"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	module *Module
}

func NewHandler(module *Module) *Handler {
	return &Handler{module: module}
}

func (h *Handler) GetSandboxByProject(c *gin.Context) {
	projectID := c.Param("project_id")
	if projectID == "" {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "project_id is required", nil)
		return
	}

	sandbox, err := h.module.Service.GetByProject(c.Request.Context(), projectID)
	if err != nil {
		if errors.Is(err, domain.ErrSandboxNotFound) && c.Query("ensure") == "true" {
			if accessErr := h.module.Service.ValidateProjectAccess(c.Request.Context(), userID(c), projectID); accessErr != nil {
				h.writeError(c, accessErr)
				return
			}
			h.module.OnProjectCreated(c.Request.Context(), projectID, userID(c))
		}
		h.writeError(c, err)
		return
	}

	apierrors.Success(c, http.StatusOK, "sandbox fetched", sandbox)
}

func (h *Handler) writeError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	code, message := apierrors.CodeOf(err)

	switch {
	case errors.Is(err, domain.ErrSandboxNotFound):
		status, code, message = http.StatusNotFound, apierrors.ErrCodeNotFound, "sandbox not found"
	case errors.Is(err, domain.ErrSandboxAccessDenied):
		status, code, message = http.StatusForbidden, apierrors.ErrCodeForbidden, "workspace access denied"
	case errors.Is(err, domain.ErrInvalidSandbox):
		status, code, message = http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error()
	}

	apierrors.Error(c, status, code, message, nil)
}

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.GET("/projects/:project_id/sandbox", handler.GetSandboxByProject)
}

func userID(c *gin.Context) string {
	value, _ := middleware.UserID(c)
	return value
}
