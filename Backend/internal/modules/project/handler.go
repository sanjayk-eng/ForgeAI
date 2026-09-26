package project

import (
	"errors"
	"net/http"

	"ai-agent/internal/middleware"
	"ai-agent/internal/modules/project/core"
	"ai-agent/internal/modules/project/github"
	projectrepo "ai-agent/internal/modules/project/repository"
	apierrors "ai-agent/internal/shared/errors"
	"ai-agent/internal/shared/pagination"
	"ai-agent/pkg/validate"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(c *gin.Context) {
	var input CreateProjectRequest
	if !bindAndValidate(c, &input) {
		return
	}

	project, err := h.service.Create(c.Request.Context(), c.Param("workspace_id"), userID(c), input)
	if err != nil {
		h.writeError(c, err)
		return
	}

	apierrors.Success(c, http.StatusCreated, "project created", project)
}

func (h *Handler) List(c *gin.Context) {
	projects, err := h.service.ListByWorkspace(c.Request.Context(), c.Param("workspace_id"), pagination.FromContext(c))
	if err != nil {
		h.writeError(c, err)
		return
	}

	apierrors.Success(c, http.StatusOK, "projects fetched", projects)
}

func (h *Handler) Get(c *gin.Context) {
	project, err := h.service.FindByID(c.Request.Context(), c.Param("project_id"))
	if err != nil {
		h.writeError(c, err)
		return
	}

	apierrors.Success(c, http.StatusOK, "project fetched", project)
}

func (h *Handler) GetBySlug(c *gin.Context) {
	project, err := h.service.FindByWorkspaceSlug(c.Request.Context(), c.Param("workspace_id"), c.Param("slug"))
	if err != nil {
		h.writeError(c, err)
		return
	}

	apierrors.Success(c, http.StatusOK, "project fetched", project)
}

func (h *Handler) Update(c *gin.Context) {
	var input UpdateProjectRequest
	if !bindAndValidate(c, &input) {
		return
	}

	project, err := h.service.Update(c.Request.Context(), c.Param("project_id"), userID(c), input)
	if err != nil {
		h.writeError(c, err)
		return
	}

	apierrors.Success(c, http.StatusOK, "project updated", project)
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.service.Delete(c.Request.Context(), c.Param("project_id"), userID(c)); err != nil {
		h.writeError(c, err)
		return
	}

	apierrors.Success(c, http.StatusOK, "project deleted", nil)
}

func (h *Handler) ConnectRepository(c *gin.Context) {
	var input ConnectRepositoryRequest
	if !bindAndValidate(c, &input) {
		return
	}

	repository, err := h.service.ConnectRepository(c.Request.Context(), c.Param("project_id"), input)
	if err != nil {
		h.writeError(c, err)
		return
	}

	apierrors.Success(c, http.StatusCreated, "repository connected", repository)
}

func (h *Handler) UpdateRepositoryBranch(c *gin.Context) {
	var input UpdateRepositoryBranchRequest
	if !bindAndValidate(c, &input) {
		return
	}

	repository, err := h.service.UpdateRepositoryBranch(c.Request.Context(), c.Param("project_id"), input.DefaultBranch)
	if err != nil {
		h.writeError(c, err)
		return
	}

	apierrors.Success(c, http.StatusOK, "repository branch updated", repository)
}

func (h *Handler) Sync(c *gin.Context) {
	project, err := h.service.SyncProject(c.Request.Context(), c.Param("project_id"))
	if err != nil {
		h.writeError(c, err)
		return
	}

	apierrors.Success(c, http.StatusOK, "repository sync completed", project)
}

func (h *Handler) SyncWorkspaceProjects(c *gin.Context) {
	result, err := h.service.SyncWorkspaceProjects(c.Request.Context(), c.Param("workspace_id"))
	if err != nil {
		h.writeError(c, err)
		return
	}

	apierrors.Success(c, http.StatusOK, "workspace project sync started", result)
}

func (h *Handler) ResolveRepository(c *gin.Context) {
	var input ResolveRepositoryRequest
	if !bindAndValidate(c, &input) {
		return
	}

	repository, err := h.service.ResolveRepository(c.Request.Context(), input.RepositoryURL)
	if err != nil {
		h.writeError(c, err)
		return
	}

	apierrors.Success(c, http.StatusOK, "repository resolved", repository)
}

func (h *Handler) ListGitHubRepositories(c *gin.Context) {
	repositories, err := h.service.ListGitHubRepositories(c.Request.Context(), c.Param("workspace_id"), userID(c), c.Query("owner"))
	if err != nil {
		h.writeError(c, err)
		return
	}

	apierrors.Success(c, http.StatusOK, "GitHub repositories fetched", repositories)
}

func (h *Handler) ImportGitHubRepositories(c *gin.Context) {
	var input ImportGitHubRepositoriesRequest
	if !bindAndValidate(c, &input) {
		return
	}

	result, err := h.service.ImportGitHubRepositories(c.Request.Context(), c.Param("workspace_id"), userID(c), input)
	if err != nil {
		h.writeError(c, err)
		return
	}

	apierrors.Success(c, http.StatusCreated, "GitHub repositories imported", result)
}

func (h *Handler) writeError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	code, message := apierrors.CodeOf(err)

	switch {
	case errors.Is(err, core.ErrInvalidProjectInput):
		status, code, message = http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error()
	case errors.Is(err, core.ErrProjectNotFound):
		status, code, message = http.StatusNotFound, apierrors.ErrCodeNotFound, err.Error()
	case errors.Is(err, core.ErrWorkspaceOwnerRequired):
		status, code, message = http.StatusForbidden, apierrors.ErrCodeForbidden, err.Error()
	case errors.Is(err, projectrepo.ErrInvalidRepositoryInput):
		status, code, message = http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error()
	case errors.Is(err, projectrepo.ErrRepositoryNotFound):
		status, code, message = http.StatusNotFound, apierrors.ErrCodeNotFound, err.Error()
	case errors.Is(err, projectrepo.ErrRepositoryConflict):
		status, code, message = http.StatusConflict, apierrors.ErrCodeConflict, err.Error()
	case errors.Is(err, github.ErrGitHubAccountUnavailable), errors.Is(err, github.ErrGitHubCatalogUnavailable):
		status, code, message = http.StatusBadRequest, apierrors.ErrCodeBadRequest, err.Error()
	case errors.Is(err, github.ErrWorkspaceOwnerRequired):
		status, code, message = http.StatusForbidden, apierrors.ErrCodeForbidden, err.Error()
	}

	apierrors.Error(c, status, code, message, nil)
}

func bindAndValidate(c *gin.Context, input any) bool {
	if err := validate.BindAndValidate(c, input); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error(), nil)
		return false
	}
	return true
}

func userID(c *gin.Context) string {
	value, _ := middleware.UserID(c)
	return value
}
