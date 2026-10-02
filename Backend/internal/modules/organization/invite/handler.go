package invite

import (
	"errors"
	"net/http"
	"strings"

	"ai-agent/internal/middleware"
	apierrors "ai-agent/internal/shared/errors"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func RegisterRoutes(protected, public gin.IRouter, handler *Handler) {
	protected.POST("/organizations/:organization_id/invites", handler.Create)
	protected.GET("/organizations/:organization_id/invites", handler.List)
	protected.POST("/invites/:token/accept", handler.Accept)
	protected.GET("/invites/my-pending", handler.ListPending)
	protected.PATCH("/invites/:invite_id", handler.UpdateStatus)
	public.GET("/public/invites/:token", handler.GetByToken)
	public.POST("/public/invites/:token/reject", handler.Reject)
}

func (h *Handler) Create(c *gin.Context) {
	var input CreateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "invalid invitation input", nil)
		return
	}
	invite, err := h.service.Create(c.Request.Context(), userID(c), c.Param("organization_id"), input)
	if err != nil {
		h.writeError(c, err)
		return
	}
	apierrors.Success(c, http.StatusCreated, "invitation created", invite)
}

func (h *Handler) List(c *gin.Context) {
	invites, err := h.service.List(c.Request.Context(), userID(c), c.Param("organization_id"), c.Query("status"))
	if err != nil {
		h.writeError(c, err)
		return
	}
	apierrors.Success(c, http.StatusOK, "invitations fetched", invites)
}

func (h *Handler) GetByToken(c *gin.Context) {
	invite, err := h.service.GetByToken(c.Request.Context(), c.Param("token"))
	if err != nil {
		h.writeError(c, err)
		return
	}
	apierrors.Success(c, http.StatusOK, "invitation fetched", invite)
}

func (h *Handler) Accept(c *gin.Context) {
	var input struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "invalid invitation response", nil)
		return
	}
	organizationID, err := h.service.Accept(c.Request.Context(), userID(c), c.Param("token"))
	if err != nil {
		h.writeError(c, err)
		return
	}
	apierrors.Success(c, http.StatusOK, "invitation accepted", gin.H{"organization_id": organizationID})
}

func (h *Handler) UpdateStatus(c *gin.Context) {
	var input struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "invalid invitation status", nil)
		return
	}
	result, err := h.service.UpdateStatus(c.Request.Context(), userID(c), c.Param("invite_id"), input.Status)
	if err != nil {
		h.writeError(c, err)
		return
	}
	apierrors.Success(c, http.StatusOK, "invitation status updated", result)
}

func (h *Handler) Reject(c *gin.Context) {
	var input struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&input); err != nil || strings.TrimSpace(input.Email) == "" {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "email is required", nil)
		return
	}
	if err := h.service.Reject(c.Request.Context(), c.Param("token"), input.Email); err != nil {
		h.writeError(c, err)
		return
	}
	apierrors.Success(c, http.StatusOK, "invitation rejected", nil)
}

func (h *Handler) ListPending(c *gin.Context) {
	invites, err := h.service.ListPending(c.Request.Context(), userID(c))
	if err != nil {
		h.writeError(c, err)
		return
	}
	apierrors.Success(c, http.StatusOK, "pending invitations fetched", invites)
}

func (h *Handler) writeError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	code, message := apierrors.CodeOf(err)
	switch {
	case errors.Is(err, ErrNotFound):
		status, code, message = http.StatusNotFound, apierrors.ErrCodeNotFound, "invitation not found"
	case errors.Is(err, ErrExpired):
		status, code, message = http.StatusGone, apierrors.ErrCodeNotFound, "invitation expired"
	case errors.Is(err, ErrAlreadyUsed), errors.Is(err, ErrDuplicate):
		status, code, message = http.StatusConflict, apierrors.ErrCodeConflict, err.Error()
	case errors.Is(err, ErrEmailMismatch), errors.Is(err, ErrInvalidInvite):
		status, code, message = http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error()
	case errors.Is(err, ErrInvalidStatus):
		status, code, message = http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error()
	case errors.Is(err, ErrInviteForbidden):
		status, code, message = http.StatusForbidden, apierrors.ErrCodeForbidden, "organization owner or admin permission is required"
	}
	apierrors.Error(c, status, code, message, nil)
}

func userID(c *gin.Context) string {
	value, _ := middleware.UserID(c)
	return value
}
