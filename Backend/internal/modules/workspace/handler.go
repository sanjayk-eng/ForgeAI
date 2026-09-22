package workspace

import (
	"net/http"

	"ai-agent/internal/middleware"
	apierrors "ai-agent/internal/shared/errors"
	"ai-agent/pkg/validate"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (handler *Handler) CreateWorkspace(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		apierrors.Error(c, http.StatusUnauthorized, apierrors.ErrCodeUnauthorized, "authenticated user is required", nil)
		return
	}

	var input CreateWorkspaceRequest
	if err := validate.BindAndValidate(c, &input); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error(), nil)
		return
	}

	workspace, err := handler.service.CreateWorkspace(c.Request.Context(), userID, input)
	if err != nil {
		apierrors.Error(c, http.StatusInternalServerError, apierrors.ErrCodeInternalServer, err.Error(), nil)
		return
	}

	apierrors.Success(c, http.StatusCreated, "workspace created", workspace)
}

func (handler *Handler) GetWorkspace(c *gin.Context) {
	workspaceID := c.Param("workspace_id")
	workspace, err := handler.service.GetWorkspace(c.Request.Context(), workspaceID)
	if err != nil {
		apierrors.Error(c, http.StatusNotFound, apierrors.ErrCodeNotFound, err.Error(), nil)
		return
	}
	apierrors.Success(c, http.StatusOK, "workspace fetched", workspace)
}

func (handler *Handler) ListWorkspaces(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		apierrors.Error(c, http.StatusUnauthorized, apierrors.ErrCodeUnauthorized, "authenticated user is required", nil)
		return
	}

	workspaces, err := handler.service.ListWorkspaces(c.Request.Context(), userID)
	if err != nil {
		apierrors.Error(c, http.StatusInternalServerError, apierrors.ErrCodeInternalServer, err.Error(), nil)
		return
	}
	apierrors.Success(c, http.StatusOK, "workspaces fetched", workspaces)
}

func (handler *Handler) AddMember(c *gin.Context) {
	workspaceID := c.Param("workspace_id")
	_, ok := middleware.UserID(c)
	if !ok {
		apierrors.Error(c, http.StatusUnauthorized, apierrors.ErrCodeUnauthorized, "authenticated user is required", nil)
		return
	}

	var input AddMemberRequest
	if err := validate.BindAndValidate(c, &input); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error(), nil)
		return
	}

	member, err := handler.service.AddMember(c.Request.Context(), workspaceID, "", input.UserID, input.Role)
	if err != nil {
		apierrors.Error(c, http.StatusInternalServerError, apierrors.ErrCodeInternalServer, err.Error(), nil)
		return
	}
	apierrors.Success(c, http.StatusCreated, "workspace member added", member)
}

func (handler *Handler) UpdateMemberRole(c *gin.Context) {
	workspaceID := c.Param("workspace_id")
	userID := c.Param("user_id")
	var input UpdateMemberRoleRequest
	if err := validate.BindAndValidate(c, &input); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error(), nil)
		return
	}

	if err := handler.service.UpdateMemberRole(c.Request.Context(), workspaceID, userID, input.Role); err != nil {
		apierrors.Error(c, http.StatusInternalServerError, apierrors.ErrCodeInternalServer, err.Error(), nil)
		return
	}
	apierrors.Success(c, http.StatusOK, "workspace member role updated", nil)
}

func (handler *Handler) RemoveMember(c *gin.Context) {
	workspaceID := c.Param("workspace_id")
	userID := c.Param("user_id")
	if err := handler.service.RemoveMember(c.Request.Context(), workspaceID, userID); err != nil {
		apierrors.Error(c, http.StatusInternalServerError, apierrors.ErrCodeInternalServer, err.Error(), nil)
		return
	}
	apierrors.Success(c, http.StatusOK, "workspace member removed", nil)
}

func (handler *Handler) ListMembers(c *gin.Context) {
	workspaceID := c.Param("workspace_id")
	members, err := handler.service.ListMembers(c.Request.Context(), workspaceID)
	if err != nil {
		apierrors.Error(c, http.StatusInternalServerError, apierrors.ErrCodeInternalServer, err.Error(), nil)
		return
	}
	apierrors.Success(c, http.StatusOK, "workspace members fetched", members)
}
