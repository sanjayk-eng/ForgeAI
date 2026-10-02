package organization

import (
	"errors"
	"net/http"

	"ai-agent/internal/middleware"
	"ai-agent/internal/modules/organization/core"
	"ai-agent/internal/modules/organization/member"
	apierrors "ai-agent/internal/shared/errors"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	coreService   core.Service
	memberService member.Service
}

func NewHandler(coreService core.Service, memberService member.Service) *Handler {
	return &Handler{
		coreService:   coreService,
		memberService: memberService,
	}
}

// Organizations

func (h *Handler) Create(c *gin.Context) {
	var input core.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "invalid input", nil)
		return
	}

	org, err := h.coreService.Create(c.Request.Context(), userID(c), input)
	if err != nil {
		h.writeError(c, err)
		return
	}

	apierrors.Success(c, http.StatusCreated, "organization created", org)
}

func (h *Handler) List(c *gin.Context) {
	orgs, err := h.coreService.ListByUser(c.Request.Context(), userID(c))
	if err != nil {
		h.writeError(c, err)
		return
	}

	apierrors.Success(c, http.StatusOK, "organizations fetched", orgs)
}

func (h *Handler) Get(c *gin.Context) {
	org, err := h.coreService.FindByID(c.Request.Context(), c.Param("organization_id"))
	if err != nil {
		h.writeError(c, err)
		return
	}

	// Validate access
	if err := h.coreService.ValidateAccess(c.Request.Context(), userID(c), org.ID); err != nil {
		h.writeError(c, err)
		return
	}

	apierrors.Success(c, http.StatusOK, "organization fetched", org)
}

func (h *Handler) Update(c *gin.Context) {
	var input core.UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "invalid input", nil)
		return
	}

	org, err := h.coreService.Update(c.Request.Context(), c.Param("organization_id"), userID(c), input)
	if err != nil {
		h.writeError(c, err)
		return
	}

	apierrors.Success(c, http.StatusOK, "organization updated", org)
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.coreService.Delete(c.Request.Context(), c.Param("organization_id"), userID(c)); err != nil {
		h.writeError(c, err)
		return
	}

	apierrors.Success(c, http.StatusOK, "organization deleted", nil)
}

// Members

func (h *Handler) ListMembers(c *gin.Context) {
	// Validate access
	orgID := c.Param("organization_id")
	if err := h.coreService.ValidateAccess(c.Request.Context(), userID(c), orgID); err != nil {
		h.writeError(c, err)
		return
	}

	members, err := h.memberService.List(c.Request.Context(), orgID)
	if err != nil {
		h.writeError(c, err)
		return
	}

	apierrors.Success(c, http.StatusOK, "members fetched", members)
}

func (h *Handler) UpdateMemberRole(c *gin.Context) {
	var input member.UpdateMemberRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "invalid input", nil)
		return
	}

	// Validate access
	orgID := c.Param("organization_id")
	if err := h.coreService.ValidateAccess(c.Request.Context(), userID(c), orgID); err != nil {
		h.writeError(c, err)
		return
	}

	memberData, err := h.memberService.UpdateRole(c.Request.Context(), c.Param("member_id"), input.Role)
	if err != nil {
		h.writeError(c, err)
		return
	}

	apierrors.Success(c, http.StatusOK, "member role updated", memberData)
}

func (h *Handler) RemoveMember(c *gin.Context) {
	// Validate access
	orgID := c.Param("organization_id")
	if err := h.coreService.ValidateAccess(c.Request.Context(), userID(c), orgID); err != nil {
		h.writeError(c, err)
		return
	}

	if err := h.memberService.Remove(c.Request.Context(), c.Param("member_id")); err != nil {
		h.writeError(c, err)
		return
	}

	apierrors.Success(c, http.StatusOK, "member removed", nil)
}

func (h *Handler) writeError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	code, message := apierrors.CodeOf(err)

	switch {
	case errors.Is(err, core.ErrOrganizationNotFound):
		status, code, message = http.StatusNotFound, apierrors.ErrCodeNotFound, "organization not found"
	case errors.Is(err, core.ErrInvalidInput):
		status, code, message = http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error()
	case errors.Is(err, member.ErrMemberNotFound):
		status, code, message = http.StatusNotFound, apierrors.ErrCodeNotFound, "member not found"
	case err.Error() == "access denied to organization":
		status, code, message = http.StatusForbidden, apierrors.ErrCodeForbidden, "access denied"
	}

	apierrors.Error(c, status, code, message, nil)
}

func userID(c *gin.Context) string {
	value, _ := middleware.UserID(c)
	return value
}
