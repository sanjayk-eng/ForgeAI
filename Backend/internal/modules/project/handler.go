package project

import (
	"errors"
	"net/http"

	"ai-agent/internal/middleware"
	apierrors "ai-agent/internal/shared/errors"
	"ai-agent/pkg/validate"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (handler *Handler) Create(c *gin.Context) {
	var input CreateProjectRequest
	if !bindAndValidate(c, &input) {
		return
	}
	project, err := handler.service.Create(c.Request.Context(), c.Param("workspace_id"), userID(c), input)
	if err != nil {
		handler.writeError(c, err)
		return
	}
	apierrors.Success(c, http.StatusCreated, "project created", project)
}

func (handler *Handler) List(c *gin.Context) {
	projects, err := handler.service.ListByWorkspace(c.Request.Context(), c.Param("workspace_id"))
	if err != nil {
		handler.writeError(c, err)
		return
	}
	apierrors.Success(c, http.StatusOK, "projects fetched", projects)
}

func (handler *Handler) Get(c *gin.Context) {
	project, err := handler.service.FindByID(c.Request.Context(), c.Param("project_id"))
	if err != nil {
		handler.writeError(c, err)
		return
	}
	apierrors.Success(c, http.StatusOK, "project fetched", project)
}

func (handler *Handler) GetBySlug(c *gin.Context) {
	project, err := handler.service.FindByWorkspaceSlug(c.Request.Context(), c.Param("workspace_id"), c.Param("slug"))
	if err != nil {
		handler.writeError(c, err)
		return
	}
	apierrors.Success(c, http.StatusOK, "project fetched", project)
}

func (handler *Handler) Update(c *gin.Context) {
	var input UpdateProjectRequest
	if !bindAndValidate(c, &input) {
		return
	}
	project, err := handler.service.Update(c.Request.Context(), c.Param("project_id"), input)
	if err != nil {
		handler.writeError(c, err)
		return
	}
	apierrors.Success(c, http.StatusOK, "project updated", project)
}

func (handler *Handler) ConnectRepository(c *gin.Context) {
	var input ConnectRepositoryRequest
	if !bindAndValidate(c, &input) {
		return
	}
	repository, err := handler.service.ConnectRepository(c.Request.Context(), c.Param("project_id"), input)
	if err != nil {
		handler.writeError(c, err)
		return
	}
	apierrors.Success(c, http.StatusCreated, "repository connected", repository)
}

func (handler *Handler) Sync(c *gin.Context) {
	project, err := handler.service.SyncProject(c.Request.Context(), c.Param("project_id"))
	if err != nil {
		handler.writeError(c, err)
		return
	}
	apierrors.Success(c, http.StatusOK, "repository sync completed", project)
}

func (handler *Handler) ResolveRepository(c *gin.Context) {
	var input ResolveRepositoryRequest
	if !bindAndValidate(c, &input) {
		return
	}
	repository, err := handler.service.ResolveRepository(c.Request.Context(), input.RepositoryURL)
	if err != nil {
		handler.writeError(c, err)
		return
	}
	apierrors.Success(c, http.StatusOK, "repository resolved", repository)
}

func (handler *Handler) writeError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	code, message := apierrors.CodeOf(err)
	switch {
	case errors.Is(err, ErrInvalidProjectInput):
		status, code, message = http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error()
	case errors.Is(err, ErrProjectNotFound):
		status, code, message = http.StatusNotFound, apierrors.ErrCodeNotFound, err.Error()
	case errors.Is(err, ErrRepositoryConflict):
		status, code, message = http.StatusConflict, apierrors.ErrCodeConflict, err.Error()
	case errors.Is(err, ErrSyncUnavailable):
		status, code, message = http.StatusConflict, apierrors.ErrCodeConflict, err.Error()
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
