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

func (handler *Handler) GetInviteByToken(c *gin.Context) {
	token := c.Param("token")
	
	invite, err := handler.service.GetInviteByToken(c.Request.Context(), token)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == ErrInviteNotFound {
			statusCode = http.StatusNotFound
		} else if err == ErrInvalidToken {
			statusCode = http.StatusBadRequest
		}
		apierrors.Error(c, statusCode, apierrors.ErrCodeInternalServer, err.Error(), nil)
		return
	}
	
	apierrors.Success(c, http.StatusOK, "invite details fetched", invite)
}

func (handler *Handler) AcceptInviteByToken(c *gin.Context) {
	token := c.Param("token")
	
	userID, ok := middleware.UserID(c)
	if !ok {
		apierrors.Error(c, http.StatusUnauthorized, apierrors.ErrCodeUnauthorized, "authenticated user is required", nil)
		return
	}
	
	var input AcceptInviteRequest
	if err := validate.BindAndValidate(c, &input); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error(), nil)
		return
	}
	
	invite, err := handler.service.AcceptInviteByToken(c.Request.Context(), token, userID, input.Email)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == ErrInviteNotFound {
			statusCode = http.StatusNotFound
		} else if err == ErrInviteExpired || err == ErrInviteAlreadyUsed || err == ErrEmailMismatch {
			statusCode = http.StatusBadRequest
		}
		apierrors.Error(c, statusCode, apierrors.ErrCodeInternalServer, err.Error(), nil)
		return
	}
	
	apierrors.Success(c, http.StatusOK, "invite accepted successfully", invite)
}

func (handler *Handler) RejectInviteByToken(c *gin.Context) {
	token := c.Param("token")
	
	var input RejectInviteRequest
	if err := validate.BindAndValidate(c, &input); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error(), nil)
		return
	}
	
	invite, err := handler.service.RejectInviteByToken(c.Request.Context(), token, input.Email)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == ErrInviteNotFound {
			statusCode = http.StatusNotFound
		} else if err == ErrEmailMismatch || err == ErrInviteAlreadyUsed {
			statusCode = http.StatusBadRequest
		}
		apierrors.Error(c, statusCode, apierrors.ErrCodeInternalServer, err.Error(), nil)
		return
	}
	
	apierrors.Success(c, http.StatusOK, "invite rejected", invite)
}

func (handler *Handler) RevokeInvite(c *gin.Context) {
	workspaceID := c.Param("workspace_id")
	inviteID := c.Param("invite_id")
	
	userID, ok := middleware.UserID(c)
	if !ok {
		apierrors.Error(c, http.StatusUnauthorized, apierrors.ErrCodeUnauthorized, "authenticated user is required", nil)
		return
	}
	
	if err := handler.service.RevokeInvite(c.Request.Context(), workspaceID, inviteID, userID); err != nil {
		apierrors.Error(c, http.StatusInternalServerError, apierrors.ErrCodeInternalServer, err.Error(), nil)
		return
	}
	
	apierrors.Success(c, http.StatusOK, "invite revoked successfully", nil)
}

func (handler *Handler) GetMyPendingInvites(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		apierrors.Error(c, http.StatusUnauthorized, apierrors.ErrCodeUnauthorized, "authenticated user is required", nil)
		return
	}
	
	// Get user info to fetch email
	user, err := handler.service.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		apierrors.Error(c, http.StatusInternalServerError, apierrors.ErrCodeInternalServer, "could not fetch user info", nil)
		return
	}
	
	invites, err := handler.service.GetPendingInvitesByEmail(c.Request.Context(), user.Email)
	if err != nil {
		apierrors.Error(c, http.StatusInternalServerError, apierrors.ErrCodeInternalServer, err.Error(), nil)
		return
	}
	
	apierrors.Success(c, http.StatusOK, "pending invites fetched", invites)
}
