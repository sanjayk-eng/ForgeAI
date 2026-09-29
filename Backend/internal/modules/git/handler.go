package git

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"ai-agent/internal/middleware"
	"ai-agent/internal/modules/auth"
	"ai-agent/internal/modules/terminal/domain"
	apierrors "ai-agent/internal/shared/errors"
	"ai-agent/internal/shared/realtime"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service       *Service
	access        SandboxAccessValidator
	githubAccount GitHubAccountStore
	events        realtime.Publisher
}

type SandboxAccessValidator interface {
	ValidateSandboxAccess(ctx context.Context, userID, sandboxID string) error
	Get(ctx context.Context, sandboxID string) (domain.Sandbox, error)
}

type GitHubAccountStore interface {
	FindGitHubAccessToken(ctx context.Context, userID string) (string, error)
	FindUserByID(ctx context.Context, userID string) (auth.UserProfile, error)
}

func NewHandler(service *Service, access SandboxAccessValidator, githubAccounts GitHubAccountStore, publishers ...realtime.Publisher) *Handler {
	handler := &Handler{service: service, access: access, githubAccount: githubAccounts}
	if len(publishers) > 0 {
		handler.events = publishers[0]
	}
	return handler
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
	diff, err := h.service.Diff(c.Request.Context(), sandboxID, c.Query("path"))
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
	if h.githubAccount == nil {
		apierrors.Error(c, http.StatusInternalServerError, apierrors.ErrCodeInternalServer, "user account store is not configured", nil)
		return
	}
	profile, err := h.githubAccount.FindUserByID(c.Request.Context(), requestUserID(c))
	if err != nil {
		apierrors.Error(c, http.StatusInternalServerError, apierrors.ErrCodeInternalServer, "could not load commit identity", nil)
		return
	}
	if strings.TrimSpace(profile.Name) == "" || strings.TrimSpace(profile.Email) == "" {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "user name and email are required to commit", nil)
		return
	}
	var payload CommitRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "invalid commit payload", nil)
		return
	}
	result, err := h.service.Commit(c.Request.Context(), sandboxID, payload, profile.Name, profile.Email)
	if err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error(), nil)
		return
	}
	h.publishGitStatus(c.Request.Context(), sandboxID)
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
	if h.githubAccount == nil {
		apierrors.Error(c, http.StatusInternalServerError, apierrors.ErrCodeInternalServer, "GitHub token store is not configured", nil)
		return
	}
	accessToken, err := h.githubAccount.FindGitHubAccessToken(c.Request.Context(), requestUserID(c))
	if err != nil {
		if errors.Is(err, auth.ErrGitHubAccountNotConnected) {
			apierrors.Error(c, http.StatusForbidden, apierrors.ErrCodeForbidden, "connect a GitHub account before pushing", nil)
			return
		}
		apierrors.Error(c, http.StatusInternalServerError, apierrors.ErrCodeInternalServer, "could not load GitHub credentials", nil)
		return
	}
	if strings.TrimSpace(accessToken) == "" {
		apierrors.Error(c, http.StatusForbidden, apierrors.ErrCodeForbidden, "connect a GitHub account before pushing", nil)
		return
	}
	result, err := h.service.Push(c.Request.Context(), sandboxID, payload, accessToken)
	if err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error(), nil)
		return
	}
	h.publishGitStatus(c.Request.Context(), sandboxID)
	apierrors.Success(c, http.StatusOK, "push completed", result)
}

func (h *Handler) publishGitStatus(ctx context.Context, sandboxID string) {
	if h.events == nil || h.access == nil {
		return
	}
	sandbox, err := h.access.Get(ctx, sandboxID)
	if err != nil {
		return
	}
	h.events.Publish(realtime.Event{
		Version: 1, Event: "git.status.changed", WorkspaceID: sandbox.WorkspaceID,
		ProjectID: sandbox.ProjectID, SandboxID: sandbox.ID,
	})
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
