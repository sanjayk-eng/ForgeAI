package github

import (
	"context"

	"ai-agent/internal/modules/github/service"
	"ai-agent/internal/modules/github/types"
)

// Service is the main interface for GitHub operations
type Service interface {
	// Integration operations
	List(ctx context.Context, organizationID string) ([]types.Integration, error)
	
	// Export repository operations
	CreateExportRepository(ctx context.Context, organizationID, projectID, userID string, private bool) (types.ExportRepository, error)
	GetExportRepository(ctx context.Context, organizationID, projectID string) (types.ExportRepository, error)
	ExportCredentials(ctx context.Context, organizationID, projectID string) (string, string, error)
	
	// Installation operations
	StartInstallation(ctx context.Context, organizationID, userID string) (string, error)
	CompleteInstallation(ctx context.Context, state, installationID string) (types.Integration, error)
	Disconnect(ctx context.Context, organizationID, integrationID string) error
	
	// Webhook operations
	HandleWebhook(ctx context.Context, event, signature string, body []byte) error
}

type githubService struct {
	integrationSvc  *service.IntegrationService
	installationSvc *service.InstallationService
	repositorySvc   *service.RepositoryService
	webhookSvc      *service.WebhookService
}

func NewService(repository Repository, config types.GitHubAppConfig) (Service, error) {
	var client *GitHubAppClient
	if config.AppID != 0 {
		var err error
		client, err = NewGitHubAppClient(config)
		if err != nil {
			return nil, err
		}
	}
	
	return &githubService{
		integrationSvc:  service.NewIntegrationService(repository),
		installationSvc: service.NewInstallationService(repository, client, config),
		repositorySvc:   service.NewRepositoryService(repository, client),
		webhookSvc:      service.NewWebhookService(repository, config.WebhookSecret, config.AppID),
	}, nil
}

// Integration operations
func (s *githubService) List(ctx context.Context, organizationID string) ([]types.Integration, error) {
	return s.integrationSvc.List(ctx, organizationID)
}

// Export repository operations
func (s *githubService) CreateExportRepository(ctx context.Context, organizationID, projectID, userID string, private bool) (types.ExportRepository, error) {
	return s.repositorySvc.CreateExportRepository(ctx, organizationID, projectID, userID, private)
}

func (s *githubService) GetExportRepository(ctx context.Context, organizationID, projectID string) (types.ExportRepository, error) {
	return s.repositorySvc.GetExportRepository(ctx, organizationID, projectID)
}

func (s *githubService) ExportCredentials(ctx context.Context, organizationID, projectID string) (string, string, error) {
	return s.repositorySvc.ExportCredentials(ctx, organizationID, projectID)
}

// Installation operations
func (s *githubService) StartInstallation(ctx context.Context, organizationID, userID string) (string, error) {
	return s.installationSvc.StartInstallation(ctx, organizationID, userID)
}

func (s *githubService) CompleteInstallation(ctx context.Context, state, installationID string) (types.Integration, error) {
	return s.installationSvc.CompleteInstallation(ctx, state, installationID)
}

func (s *githubService) Disconnect(ctx context.Context, organizationID, integrationID string) error {
	return s.installationSvc.Disconnect(ctx, organizationID, integrationID)
}

// Webhook operations
func (s *githubService) HandleWebhook(ctx context.Context, event, signature string, body []byte) error {
	return s.webhookSvc.HandleWebhook(ctx, event, signature, body)
}
