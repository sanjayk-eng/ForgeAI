package github

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"ai-agent/internal/middleware"
	"ai-agent/internal/modules/organization/core"
	"ai-agent/internal/modules/organization/member"
	"ai-agent/internal/shared/logger"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service       Service
	organizations core.Service
	members       member.Service
	frontendURL   string
	logger        logger.Logger
}

func NewHandler(service Service, organizations core.Service, members member.Service, frontendURL string, log logger.Logger) *Handler {
	return &Handler{service: service, organizations: organizations, members: members, frontendURL: frontendURL, logger: log}
}

func (h *Handler) List(c *gin.Context) {
	organizationID := c.Param("organization_id")
	userID, ok := middleware.UserID(c)
	if !ok || h.organizations.ValidateAccess(c.Request.Context(), userID, organizationID) != nil {
		writeError(c, http.StatusForbidden, "FORBIDDEN", "organization access is required")
		return
	}
	items, err := h.service.List(c.Request.Context(), organizationID)
	if err != nil {
		h.logError(c, "list integrations", err)
		writeError(c, http.StatusInternalServerError, "INTEGRATION_ERROR", "could not load organization integrations")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *Handler) CreateExportRepository(c *gin.Context) {
	organizationID := c.Param("organization_id")
	projectID := c.Param("project_id")
	userID, ok := middleware.UserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication is required")
		return
	}
	if err := h.organizations.ValidateAccess(c.Request.Context(), userID, organizationID); err != nil {
		writeError(c, http.StatusForbidden, "FORBIDDEN", "organization access is required")
		return
	}
	if err := h.members.CanManage(c.Request.Context(), userID, organizationID); err != nil {
		writeError(c, http.StatusForbidden, "FORBIDDEN", "organization owner or admin permission is required to create GitHub repositories")
		return
	}
	var request CreateExportRepositoryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_REQUEST", "repository visibility is required")
		return
	}
	repository, err := h.service.CreateExportRepository(c.Request.Context(), organizationID, projectID, userID, *request.Private)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": repository})
}

func (h *Handler) GetExportRepository(c *gin.Context) {
	organizationID := c.Param("organization_id")
	userID, ok := middleware.UserID(c)
	if !ok || h.organizations.ValidateAccess(c.Request.Context(), userID, organizationID) != nil {
		writeError(c, http.StatusForbidden, "FORBIDDEN", "organization access is required")
		return
	}
	repository, err := h.service.GetExportRepository(c.Request.Context(), organizationID, c.Param("project_id"))
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": repository})
}

func (h *Handler) Install(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication is required")
		return
	}
	organizationID := c.Query("organization_id")
	if err := h.organizations.ValidateAccess(c.Request.Context(), userID, organizationID); err != nil {
		writeError(c, http.StatusForbidden, "FORBIDDEN", "organization access is required")
		return
	}
	if err := h.members.CanManage(c.Request.Context(), userID, organizationID); err != nil {
		writeError(c, http.StatusForbidden, "FORBIDDEN", "organization owner or admin permission is required")
		return
	}
	installURL, err := h.service.StartInstallation(c.Request.Context(), organizationID, userID)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"url": installURL}})
}

func (h *Handler) Callback(c *gin.Context) {
	integration, err := h.service.CompleteInstallation(c.Request.Context(), c.Query("state"), c.Query("installation_id"))
	if err != nil {
		h.logError(c, "complete installation callback", err)
		c.Redirect(http.StatusSeeOther, h.frontendURL+"/?github_integration=error")
		return
	}
	redirectURL, err := url.Parse(h.frontendURL + "/organizations/" + url.PathEscape(integration.OrganizationID) + "/settings/integrations")
	if err != nil {
		h.logError(c, "build installation redirect", err)
		c.Redirect(http.StatusSeeOther, h.frontendURL+"/?github_integration=error")
		return
	}
	query := redirectURL.Query()
	query.Set("github_integration", "connected")
	redirectURL.RawQuery = query.Encode()
	c.Redirect(http.StatusSeeOther, redirectURL.String())
}

func (h *Handler) Disconnect(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication is required")
		return
	}
	organizationID := c.Param("organization_id")
	if err := h.organizations.ValidateAccess(c.Request.Context(), userID, organizationID); err != nil {
		writeError(c, http.StatusForbidden, "FORBIDDEN", "organization access is required")
		return
	}
	if err := h.members.CanManage(c.Request.Context(), userID, organizationID); err != nil {
		writeError(c, http.StatusForbidden, "FORBIDDEN", "organization owner or admin permission is required")
		return
	}
	if err := h.service.Disconnect(c.Request.Context(), organizationID, c.Param("integration_id")); err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) Webhook(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		writeError(c, http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE", "webhook payload is too large")
		return
	}
	if err := h.service.HandleWebhook(c.Request.Context(), c.GetHeader("X-GitHub-Event"), c.GetHeader("X-Hub-Signature-256"), body); err != nil {
		h.logError(c, "process webhook", err)
		writeError(c, http.StatusUnauthorized, "INVALID_WEBHOOK", "GitHub webhook could not be verified")
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) writeServiceError(c *gin.Context, err error) {
	h.logError(c, "GitHub request", err)
	var githubErr *githubAPIError
	if errors.As(err, &githubErr) {
		status, message := githubAPIErrorResponse(githubErr)
		writeError(c, status, "GITHUB_ERROR", message)
		return
	}

	switch {
	case errors.Is(err, ErrGitHubAppNotConfigured):
		writeError(c, http.StatusServiceUnavailable, "GITHUB_APP_UNAVAILABLE", "GitHub App integration is not configured")
	case errors.Is(err, ErrStateInvalid):
		writeError(c, http.StatusBadRequest, "INVALID_STATE", "GitHub installation state is invalid or expired")
	case errors.Is(err, ErrIntegrationNotFound):
		writeError(c, http.StatusConflict, "GITHUB_NOT_CONNECTED", "connect an active GitHub App installation to this organization first")
	case errors.Is(err, ErrProjectNotFound):
		writeError(c, http.StatusNotFound, "PROJECT_NOT_FOUND", "project was not found in this organization")
	case errors.Is(err, ErrExportRepositoryNotFound):
		writeError(c, http.StatusNotFound, "EXPORT_REPOSITORY_NOT_FOUND", "project has no GitHub export repository")
	default:
		writeError(c, http.StatusBadGateway, "GITHUB_ERROR", "GitHub integration operation failed")
	}
}

func (h *Handler) logError(c *gin.Context, operation string, err error) {
	if h.logger != nil {
		h.logger.Error(c.Request.Context(), "GitHub operation failed", "operation", operation, "error", err)
	}
}

func githubAPIErrorResponse(err *githubAPIError) (int, string) {
	if err.operation == "installation token request" {
		switch err.status {
		case http.StatusUnauthorized:
			return http.StatusBadGateway, "GitHub rejected the app credentials. Check the GitHub App ID and private key."
		case http.StatusForbidden:
			return http.StatusBadGateway, "GitHub refused an installation token. Check that the app installation is active and its permissions have not changed."
		case http.StatusNotFound:
			return http.StatusBadGateway, "The GitHub App installation could not be found. Reconnect the GitHub App to this organization."
		}
	}

	if err.operation == "repository creation" {
		switch err.status {
		case http.StatusUnauthorized:
			return http.StatusBadGateway, "GitHub rejected the app credentials. Check the GitHub App ID and private key."
		case http.StatusForbidden:
			return http.StatusForbidden, "GitHub refused repository creation. Enable the GitHub App's Repository creation permission, approve the updated permissions for this installation, and check that repository creation is allowed for this account."
		case http.StatusUnprocessableEntity:
			return http.StatusConflict, "GitHub rejected this repository name or visibility. A repository with that name may already exist, or an account policy may block creation."
		}
	}

	return http.StatusBadGateway, fmt.Sprintf("GitHub %s failed with HTTP %d.", err.operation, err.status)
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}
