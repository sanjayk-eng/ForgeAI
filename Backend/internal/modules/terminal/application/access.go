package application

import (
	"context"
	"fmt"
	"strings"

	"ai-agent/internal/modules/terminal/domain"
)

func (service *Service) ValidateProjectAccess(ctx context.Context, userID, projectID string) error {
	_, err := service.workspaceForUser(ctx, userID, projectID)
	return err
}

func (service *Service) ValidateSandboxAccess(ctx context.Context, userID, sandboxID string) error {
	sandbox, err := service.Get(ctx, sandboxID)
	if err != nil {
		return err
	}
	allowed, err := service.store.CanAccessWorkspace(ctx, sandbox.OrganizationID, strings.TrimSpace(userID))
	if err != nil {
		return fmt.Errorf("check sandbox workspace access: %w", err)
	}
	if !allowed {
		return domain.ErrSandboxAccessDenied
	}
	return nil
}

func (service *Service) workspaceForUser(ctx context.Context, userID, projectID string) (string, error) {
	userID = strings.TrimSpace(userID)
	projectID = strings.TrimSpace(projectID)
	if userID == "" || projectID == "" || service.store == nil {
		return "", domain.ErrInvalidSandbox
	}
	organizationID, err := service.store.FindProjectWorkspace(ctx, projectID)
	if err != nil {
		return "", err
	}
	allowed, err := service.store.CanAccessWorkspace(ctx, organizationID, userID)
	if err != nil {
		return "", fmt.Errorf("check sandbox workspace access: %w", err)
	}
	if !allowed {
		return "", domain.ErrSandboxAccessDenied
	}
	return organizationID, nil
}
