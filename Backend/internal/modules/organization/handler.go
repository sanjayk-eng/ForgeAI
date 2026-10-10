package organization

import (
	"net/http"

	"ai-agent/internal/middleware"
	"ai-agent/internal/modules/organization/core"
	"ai-agent/internal/modules/organization/member"
	basehandler "ai-agent/internal/shared/handler"
	apierrors "ai-agent/internal/shared/errors"
	"ai-agent/internal/shared/pagination"
	"ai-agent/pkg/validate"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	basehandler.BaseHandler
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
	userID, ok := middleware.ExtractUserID(c)
	if !ok {
		return
	}
	
	var input core.CreateOrganizationRequest
	if err := validate.BindAndValidate(c, &input); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error(), nil)
		return
	}

	org, err := h.coreService.Create(c.Request.Context(), userID, input)
	if err != nil {
		h.writeError(c, err)
		return
	}

	apierrors.Success(c, http.StatusCreated, "organization created", org)
}

func (h *Handler) List(c *gin.Context) {
	userID, ok := middleware.ExtractUserID(c)
	if !ok {
		return
	}
	
	orgs, err := h.coreService.ListByUser(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, err)
		return
	}

	apierrors.Success(c, http.StatusOK, "organizations fetched", orgs)
}

func (h *Handler) Get(c *gin.Context) {
	userID, ok := middleware.ExtractUserID(c)
	if !ok {
		return
	}
	
	org, err := h.coreService.FindByID(c.Request.Context(), c.Param("organization_id"))
	if err != nil {
		h.writeError(c, err)
		return
	}

	// Validate access
	if err := h.coreService.ValidateAccess(c.Request.Context(), userID, org.ID); err != nil {
		apierrors.Error(c, http.StatusForbidden, apierrors.ErrCodeForbidden, "access denied", nil)
		return
	}

	apierrors.Success(c, http.StatusOK, "organization fetched", org)
}

func (h *Handler) Update(c *gin.Context) {
	userID, ok := middleware.ExtractUserID(c)
	if !ok {
		return
	}
	
	var input core.UpdateOrganizationRequest
	if err := validate.BindAndValidate(c, &input); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error(), nil)
		return
	}

	org, err := h.coreService.Update(c.Request.Context(), c.Param("organization_id"), userID, input)
	if err != nil {
		h.writeError(c, err)
		return
	}

	apierrors.Success(c, http.StatusOK, "organization updated", org)
}

func (h *Handler) Delete(c *gin.Context) {
	userID, ok := middleware.ExtractUserID(c)
	if !ok {
		return
	}
	
	if err := h.coreService.Delete(c.Request.Context(), c.Param("organization_id"), userID); err != nil {
		h.writeError(c, err)
		return
	}

	apierrors.Success(c, http.StatusOK, "organization deleted", nil)
}

// Members

func (h *Handler) ListMembers(c *gin.Context) {
	userID, ok := middleware.ExtractUserID(c)
	if !ok {
		return
	}
	
	// Validate access
	orgID := c.Param("organization_id")
	if err := h.coreService.ValidateAccess(c.Request.Context(), userID, orgID); err != nil {
		apierrors.Error(c, http.StatusForbidden, apierrors.ErrCodeForbidden, "access denied", nil)
		return
	}

	page := pagination.FromContext(c)
	members, total, err := h.memberService.List(c.Request.Context(), orgID, page)
	if err != nil {
		h.writeError(c, err)
		return
	}

	apierrors.Success(c, http.StatusOK, "members fetched", pagination.NewResult(members, page, total))
}

func (h *Handler) UpdateMemberRole(c *gin.Context) {
	userID, ok := middleware.ExtractUserID(c)
	if !ok {
		return
	}
	
	var input member.UpdateMemberRequest
	if err := validate.BindAndValidate(c, &input); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error(), nil)
		return
	}

	// Validate access
	orgID := c.Param("organization_id")
	if err := h.memberService.CanManage(c.Request.Context(), userID, orgID); err != nil {
		apierrors.Error(c, http.StatusForbidden, apierrors.ErrCodeForbidden, "organization owner or admin permission is required", nil)
		return
	}

	memberData, err := h.memberService.UpdateRole(c.Request.Context(), orgID, c.Param("user_id"), input.Role)
	if err != nil {
		h.writeError(c, err)
		return
	}

	apierrors.Success(c, http.StatusOK, "member role updated", memberData)
}

func (h *Handler) RemoveMember(c *gin.Context) {
	userID, ok := middleware.ExtractUserID(c)
	if !ok {
		return
	}
	
	// Validate access
	orgID := c.Param("organization_id")
	if err := h.memberService.CanManage(c.Request.Context(), userID, orgID); err != nil {
		apierrors.Error(c, http.StatusForbidden, apierrors.ErrCodeForbidden, "organization owner or admin permission is required", nil)
		return
	}

	if err := h.memberService.Remove(c.Request.Context(), orgID, c.Param("user_id")); err != nil {
		h.writeError(c, err)
		return
	}

	apierrors.Success(c, http.StatusOK, "member removed", nil)
}

func (h *Handler) writeError(c *gin.Context, err error) {
	errorMappings := map[error]basehandler.ErrorMapping{
		core.ErrOrganizationNotFound: {http.StatusNotFound, apierrors.ErrCodeNotFound, "organization not found"},
		core.ErrInvalidInput:         {http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error()},
		member.ErrMemberNotFound:     {http.StatusNotFound, apierrors.ErrCodeNotFound, "member not found"},
		member.ErrManageDenied:       {http.StatusForbidden, apierrors.ErrCodeForbidden, "organization owner or admin permission is required"},
		member.ErrOwnerProtected:     {http.StatusConflict, apierrors.ErrCodeConflict, err.Error()},
		member.ErrInvalidRole:        {http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error()},
	}
	
	// Check if it's access denied error
	if err.Error() == "access denied to organization" {
		apierrors.Error(c, http.StatusForbidden, apierrors.ErrCodeForbidden, "access denied", nil)
		return
	}
	
	h.WriteServiceError(c, err, errorMappings)
}
