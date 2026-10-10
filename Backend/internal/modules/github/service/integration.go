package service

import (
	"context"

	"ai-agent/internal/modules/github/types"
)

// IntegrationService handles listing and managing GitHub integrations
type IntegrationService struct {
	repo IntegrationRepository
}

func NewIntegrationService(repo IntegrationRepository) *IntegrationService {
	return &IntegrationService{repo: repo}
}

func (s *IntegrationService) List(ctx context.Context, organizationID string) ([]types.Integration, error) {
	return s.repo.ListGitHub(ctx, organizationID)
}

// IntegrationRepository defines data access for integration operations
type IntegrationRepository interface {
	ListGitHub(ctx context.Context, organizationID string) ([]types.Integration, error)
}
