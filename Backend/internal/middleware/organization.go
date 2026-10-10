package middleware

import (
	"context"
	"net/http"

	"ai-agent/internal/shared/errors"

	"github.com/gin-gonic/gin"
)

// OrganizationAccessValidator defines the interface for validating organization access
type OrganizationAccessValidator interface {
	ValidateAccess(ctx context.Context, userID, organizationID string) error
}

// OrganizationManageValidator defines the interface for validating management permissions
type OrganizationManageValidator interface {
	CanManage(ctx context.Context, userID, organizationID string) error
}

// RequireOrgAccess middleware validates that the authenticated user has access to the organization
func RequireOrgAccess(validator OrganizationAccessValidator, paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := UserID(c)
		if !ok {
			errors.Error(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized, "authentication is required", nil)
			c.Abort()
			return
		}

		organizationID := c.Param(paramName)
		if organizationID == "" {
			errors.Error(c, http.StatusBadRequest, errors.ErrCodeBadRequest, "organization ID is required", nil)
			c.Abort()
			return
		}

		if err := validator.ValidateAccess(c.Request.Context(), userID, organizationID); err != nil {
			errors.Error(c, http.StatusForbidden, errors.ErrCodeForbidden, "organization access is required", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireOrgManage middleware validates that the authenticated user can manage the organization
func RequireOrgManage(orgValidator OrganizationAccessValidator, manageValidator OrganizationManageValidator, paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := UserID(c)
		if !ok {
			errors.Error(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized, "authentication is required", nil)
			c.Abort()
			return
		}

		organizationID := c.Param(paramName)
		if organizationID == "" {
			errors.Error(c, http.StatusBadRequest, errors.ErrCodeBadRequest, "organization ID is required", nil)
			c.Abort()
			return
		}

		if err := orgValidator.ValidateAccess(c.Request.Context(), userID, organizationID); err != nil {
			errors.Error(c, http.StatusForbidden, errors.ErrCodeForbidden, "organization access is required", nil)
			c.Abort()
			return
		}

		if err := manageValidator.CanManage(c.Request.Context(), userID, organizationID); err != nil {
			errors.Error(c, http.StatusForbidden, errors.ErrCodeForbidden, "organization owner or admin permission is required", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// ExtractUserID extracts the user ID and writes error response if not found
func ExtractUserID(c *gin.Context) (string, bool) {
	userID, ok := UserID(c)
	if !ok {
		errors.Error(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized, "authentication is required", nil)
		return "", false
	}
	return userID, true
}
