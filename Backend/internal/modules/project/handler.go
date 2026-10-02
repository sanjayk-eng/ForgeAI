package project

import (
	"errors"
	"net/http"

	"ai-agent/internal/middleware"
	"ai-agent/internal/modules/project/core"
	apierrors "ai-agent/internal/shared/errors"
	"ai-agent/internal/shared/pagination"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service core.Service
}

func NewHandler(service core.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(c *gin.Context) {
	var input core.CreateProjectRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "invalid input", nil)
		return
	}

	project, err := h.service.Create(c.Request.Context(), c.Param("organization_id"), userID(c), input)
	if err != nil {
		h.writeError(c, err)
		return
	}

	apierrors.Success(c, http.StatusCreated, "project created", project)
}

func (h *Handler) List(c *gin.Context) {
	projects, err := h.service.ListByOrganization(c.Request.Context(), c.Param("organization_id"), pagination.FromContext(c))
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
	project, err := h.service.FindByOrganizationSlug(c.Request.Context(), c.Param("organization_id"), c.Param("slug"))
	if err != nil {
		h.writeError(c, err)
		return
	}

	apierrors.Success(c, http.StatusOK, "project fetched", project)
}

func (h *Handler) Update(c *gin.Context) {
	var input core.UpdateProjectRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "invalid input", nil)
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

func (h *Handler) writeError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	code, message := apierrors.CodeOf(err)

	switch {
	case errors.Is(err, core.ErrInvalidProjectInput):
		status, code, message = http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error()
	case errors.Is(err, core.ErrProjectNotFound):
		status, code, message = http.StatusNotFound, apierrors.ErrCodeNotFound, err.Error()
	case errors.Is(err, core.ErrOrganizationAccessRequired):
		status, code, message = http.StatusForbidden, apierrors.ErrCodeForbidden, "organization access required"
	}

	apierrors.Error(c, status, code, message, nil)
}

func userID(c *gin.Context) string {
	value, _ := middleware.UserID(c)
	return value
}
