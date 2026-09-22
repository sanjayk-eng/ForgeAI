package workspace

import (
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

func (handler *Handler) CreateWorkspace(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	var input CreateWorkspaceRequest
	if !bindAndValidate(c, &input) {
		return
	}

	workspace, err := handler.service.CreateWorkspace(c.Request.Context(), userID, input)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	apierrors.Success(c, http.StatusCreated, "workspace created", workspace)
}

func (handler *Handler) GetWorkspace(c *gin.Context) {
	workspace, err := handler.service.GetWorkspace(c.Request.Context(), c.Param("workspace_id"))
	if err != nil {
		handleError(c, http.StatusNotFound, err)
		return
	}
	apierrors.Success(c, http.StatusOK, "workspace fetched", workspace)
}

func (handler *Handler) UpdateWorkspace(c *gin.Context) {
	var input UpdateWorkspaceRequest
	if !bindAndValidate(c, &input) {
		return
	}

	workspace, err := handler.service.UpdateWorkspace(c.Request.Context(), c.Param("workspace_id"), input)
	if err != nil {
		handleError(c, http.StatusNotFound, err)
		return
	}
	apierrors.Success(c, http.StatusOK, "workspace updated", workspace)
}

func (handler *Handler) ListWorkspaces(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	workspaces, err := handler.service.ListWorkspaces(c.Request.Context(), userID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	apierrors.Success(c, http.StatusOK, "workspaces fetched", workspaces)
}

func currentUserID(c *gin.Context) (string, bool) {
	userID, ok := middleware.UserID(c)
	if !ok {
		apierrors.Error(c, http.StatusUnauthorized, apierrors.ErrCodeUnauthorized, "authenticated user is required", nil)
	}
	return userID, ok
}

func bindAndValidate(c *gin.Context, input any) bool {
	if err := validate.BindAndValidate(c, input); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error(), nil)
		return false
	}
	return true
}

func handleError(c *gin.Context, status int, err error) {
	if err == nil {
		return
	}
	code, message := apierrors.CodeOf(err)
	apierrors.Error(c, status, code, message, nil)
}
