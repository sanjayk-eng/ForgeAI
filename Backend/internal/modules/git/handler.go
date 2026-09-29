package git

import (
	"context"
	"net/http"
	"strings"

	"ai-agent/internal/middleware"
	apierrors "ai-agent/internal/shared/errors"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
	access  SandboxAccessValidator
}

type SandboxAccessValidator interface {
	ValidateSandboxAccess(ctx context.Context, userID, sandboxID string) error
}

func NewHandler(service *Service, access SandboxAccessValidator) *Handler {
	return &Handler{service: service, access: access}
}

func (h *Handler) Status(c *gin.Context) {
	sandboxID := c.Param("sandbox_id")
	if strings.TrimSpace(sandboxID) == "" {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "sandbox_id is required", nil)
		return
	}
	if !h.authorize(c, sandboxID) {
		return
	}
	status, err := h.service.Status(c.Request.Context(), sandboxID)
	if err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error(), nil)
		return
	}
	apierrors.Success(c, http.StatusOK, "git status fetched", gin.H{"status": status})
}

func (h *Handler) Diff(c *gin.Context) {
	sandboxID := c.Param("sandbox_id")
	if strings.TrimSpace(sandboxID) == "" {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "sandbox_id is required", nil)
		return
	}
	if !h.authorize(c, sandboxID) {
		return
	}
	diff, err := h.service.Diff(c.Request.Context(), sandboxID)
	if err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error(), nil)
		return
	}
	apierrors.Success(c, http.StatusOK, "git diff fetched", diff)
}

func (h *Handler) Commit(c *gin.Context) {
	sandboxID := c.Param("sandbox_id")
	if strings.TrimSpace(sandboxID) == "" {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "sandbox_id is required", nil)
		return
	}
	if !h.authorize(c, sandboxID) {
		return
	}
	var payload CommitRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "invalid commit payload", nil)
		return
	}
	result, err := h.service.Commit(c.Request.Context(), sandboxID, payload)
	if err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error(), nil)
		return
	}
	apierrors.Success(c, http.StatusOK, "commit created", result)
}

func (h *Handler) Push(c *gin.Context) {
	sandboxID := c.Param("sandbox_id")
	if strings.TrimSpace(sandboxID) == "" {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "sandbox_id is required", nil)
		return
	}
	if !h.authorize(c, sandboxID) {
		return
	}
	var payload PushRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "invalid push payload", nil)
		return
	}
	result, err := h.service.Push(c.Request.Context(), sandboxID, payload)
	if err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error(), nil)
		return
	}
	apierrors.Success(c, http.StatusOK, "push completed", result)
}

func (h *Handler) authorize(c *gin.Context, sandboxID string) bool {
	if h.access == nil {
		apierrors.Error(c, http.StatusInternalServerError, apierrors.ErrCodeInternalServer, "sandbox access validator is not configured", nil)
		return false
	}
	if err := h.access.ValidateSandboxAccess(c.Request.Context(), requestUserID(c), sandboxID); err != nil {
		apierrors.Error(c, http.StatusForbidden, apierrors.ErrCodeForbidden, "workspace access denied", nil)
		return false
	}
	return true
}

func requestUserID(c *gin.Context) string {
	value, _ := middleware.UserID(c)
	return value
}
