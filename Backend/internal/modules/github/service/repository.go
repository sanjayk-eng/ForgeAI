package service

import (
	"context"

	"ai-agent/internal/modules/github/types"
)

// RepositoryService handles GitHub export repository operations
type RepositoryService struct {
	repo      ExportRepository
	apiClient GitHubAPIClient
}

func NewRepositoryService(repo ExportRepository, client GitHubAPIClient) *RepositoryService {
	return &RepositoryService{
		repo:      repo,
		apiClient: client,
	}
}

func (s *RepositoryService) CreateExportRepository(ctx context.Context, organizationID, projectID, userID string, private bool) (types.ExportRepository, error) {
	if s.apiClient == nil {
		return types.ExportRepository{}, ErrGitHubAppNotConfigured
	}
	
	exportContext, err := s.repo.FindExportContext(ctx, organizationID, projectID)
	if err != nil {
		return types.ExportRepository{}, err
	}
	
	if exportContext.Existing != nil {
		return *exportContext.Existing, nil
	}
	
	token, err := s.apiClient.InstallationToken(ctx, exportContext.InstallationID)
	if err != nil {
		return types.ExportRepository{}, err
	}
	
	created, err := s.apiClient.CreateRepository(ctx, token, exportContext.AccountName, exportContext.AccountType, exportContext.ProjectSlug, private)
	if err != nil {
		return types.ExportRepository{}, err
	}
	
	created.IntegrationID = exportContext.IntegrationID
	return s.repo.SaveExportRepository(ctx, projectID, exportContext.IntegrationID, created, userID)
}

func (s *RepositoryService) GetExportRepository(ctx context.Context, organizationID, projectID string) (types.ExportRepository, error) {
	return s.repo.FindExportRepository(ctx, organizationID, projectID)
}

func (s *RepositoryService) ExportCredentials(ctx context.Context, organizationID, projectID string) (string, string, error) {
	if s.apiClient == nil {
		return "", "", ErrGitHubAppNotConfigured
	}
	
	installationID, repositoryURL, err := s.repo.FindExportCredentials(ctx, organizationID, projectID)
	if err != nil {
		return "", "", err
	}
	
	token, err := s.apiClient.InstallationToken(ctx, installationID)
	if err != nil {
		return "", "", err
	}
	
	return token, repositoryURL, nil
}

// ExportRepository defines data access for export repository operations
type ExportRepository interface {
	FindExportContext(ctx context.Context, organizationID, projectID string) (types.ExportContext, error)
	SaveExportRepository(ctx context.Context, projectID, integrationID string, repository types.ExportRepository, userID string) (types.ExportRepository, error)
	FindExportRepository(ctx context.Context, organizationID, projectID string) (types.ExportRepository, error)
	FindExportCredentials(ctx context.Context, organizationID, projectID string) (string, string, error)
}
