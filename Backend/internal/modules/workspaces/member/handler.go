package member

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

func (handler *Handler) AddMember(c *gin.Context) {
	workspaceID := c.Param("workspace_id")
	currentUser, ok := currentUserID(c)
	if !ok {
		return
	}

	var input AddMemberRequest
	if err := validate.BindAndValidate(c, &input); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error(), nil)
		return
	}

	member, err := handler.service.AddMember(c.Request.Context(), workspaceID, currentUser, input.UserID, input.Role)
	if err != nil {
		apierrors.Error(c, http.StatusInternalServerError, apierrors.ErrCodeInternalServer, err.Error(), nil)
		return
	}
	apierrors.Success(c, http.StatusCreated, "workspace member added", member)
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

func currentUserID(c *gin.Context) (string, bool) {
	userID, ok := middleware.UserID(c)
	if !ok {
		apierrors.Error(c, http.StatusUnauthorized, apierrors.ErrCodeUnauthorized, "authenticated user is required", nil)
	}
	return userID, ok
}
