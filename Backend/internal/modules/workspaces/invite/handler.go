package invite

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

func (handler *Handler) CreateInvite(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		apierrors.Error(c, http.StatusUnauthorized, apierrors.ErrCodeUnauthorized, "authenticated user is required", nil)
		return
	}

	workspaceID := c.Param("workspace_id")
	var input CreateInviteRequest
	if err := validate.BindAndValidate(c, &input); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error(), nil)
		return
	}

	invite, err := handler.service.CreateInvite(c.Request.Context(), workspaceID, userID, input.Email, input.Role)
	if err != nil {
		apierrors.Error(c, http.StatusInternalServerError, apierrors.ErrCodeInternalServer, err.Error(), nil)
		return
	}
	apierrors.Success(c, http.StatusCreated, "workspace invite created", invite)
}

func (handler *Handler) ListInvites(c *gin.Context) {
	workspaceID := c.Param("workspace_id")
	invites, err := handler.service.ListInvites(c.Request.Context(), workspaceID)
	if err != nil {
		apierrors.Error(c, http.StatusInternalServerError, apierrors.ErrCodeInternalServer, err.Error(), nil)
		return
	}
	apierrors.Success(c, http.StatusOK, "workspace invites fetched", invites)
}

func (handler *Handler) UpdateInviteStatus(c *gin.Context) {
	workspaceID := c.Param("workspace_id")
	inviteID := c.Param("invite_id")
	userID, ok := middleware.UserID(c)
	if !ok {
		apierrors.Error(c, http.StatusUnauthorized, apierrors.ErrCodeUnauthorized, "authenticated user is required", nil)
		return
	}

	var input UpdateInviteStatusRequest
	if err := validate.BindAndValidate(c, &input); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error(), nil)
		return
	}

	invite, err := handler.service.UpdateInviteStatus(c.Request.Context(), workspaceID, inviteID, userID, input.Status)
	if err != nil {
		apierrors.Error(c, http.StatusInternalServerError, apierrors.ErrCodeInternalServer, err.Error(), nil)
		return
	}
	apierrors.Success(c, http.StatusOK, "workspace invite status updated", invite)
}
