package github

import (
	"context"
	"errors"
	"fmt"

	"ai-agent/internal/modules/project/repository"
)

var (
	ErrGitHubAccountUnavailable = errors.New("GitHub account is not connected; sign in with GitHub again")
	ErrGitHubCatalogUnavailable = errors.New("GitHub repository catalog is not configured")
	ErrWorkspaceOwnerRequired   = errors.New("only the workspace owner can access GitHub repositories")
)

type CatalogService interface {
	ListRepositories(ctx context.Context, accessToken, owner string) (CatalogResponse, error)
	ImportRepositories(ctx context.Context, workspaceID, userID string, repositories []repository.ConnectRepositoryRequest, projectRepo ProjectRepository, coreService CoreService) (ImportResponse, error)
}

type ProjectRepository interface {
	RepositoryExists(ctx context.Context, workspaceID string, githubRepositoryID int64) (bool, error)
}

type CoreService interface {
	IsWorkspaceOwner(ctx context.Context, workspaceID, userID string) (bool, error)
}

type ImportResponse struct {
	Projects []ImportedProject `json:"projects"`
	Skipped  int               `json:"skipped"`
}

type ImportedProject struct {
	ID         string                       `json:"id"`
	Name       string                       `json:"name"`
	Repository repository.ProjectRepository `json:"repository"`
}

type catalogService struct {
	client Client
}

func NewCatalogService(client Client) CatalogService {
	return &catalogService{client: client}
}

func (s *catalogService) ListRepositories(ctx context.Context, accessToken, owner string) (CatalogResponse, error) {
	catalog, ok := s.client.(Catalog)
	if !ok {
		return CatalogResponse{}, ErrGitHubCatalogUnavailable
	}

	result := CatalogResponse{
		Organizations: []string{},
		Repositories:  []RepositoryOption{},
	}

	if owner == "" {
		account, err := catalog.GetAccountLogin(ctx, accessToken)
		if err != nil {
			return CatalogResponse{}, fmt.Errorf("failed to get GitHub account: %w", err)
		}
		organizations, err := catalog.ListOrganizations(ctx, accessToken)
		if err != nil {
			return CatalogResponse{}, fmt.Errorf("failed to list organizations: %w", err)
		}
		result.Account = account
		result.Organizations = organizations
		
		// Add a warning if no organizations found
		if len(organizations) == 0 {
			result.Warning = "No organizations found. If you belong to organizations, try disconnecting and reconnecting your GitHub account to grant organization access permissions."
		}
		
		return result, nil
	}

	account, err := catalog.GetAccountLogin(ctx, accessToken)
	if err != nil {
		return CatalogResponse{}, err
	}
	result.Account = account

	organization := owner
	repositoryOwner := owner
	if owner == PersonalAccountOwner || owner == account {
		organization = ""
		repositoryOwner = account
	}

	repositories, err := catalog.ListRepositories(ctx, accessToken, organization)
	if err != nil {
		return CatalogResponse{}, err
	}
	for _, repo := range repositories {
		result.Repositories = append(result.Repositories, toRepositoryOption(repo, repositoryOwner))
	}

	return result, nil
}

func (s *catalogService) ImportRepositories(ctx context.Context, workspaceID, userID string, repositories []repository.ConnectRepositoryRequest, projectRepo ProjectRepository, coreService CoreService) (ImportResponse, error) {
	// This method signature is defined but implementation will be in the orchestrator service
	return ImportResponse{}, fmt.Errorf("not implemented - use orchestrator service")
}

func toRepositoryOption(repo Repository, organization string) RepositoryOption {
	return RepositoryOption{
		ConnectRepositoryRequest: repository.ConnectRepositoryRequest{
			GitHubRepositoryID:   repo.ID,
			GitHubOwner:          repo.Owner,
			GitHubRepositoryName: repo.Name,
			RepositoryURL:        repo.URL,
			Private:              repo.Private,
			DefaultBranch:        repo.DefaultBranch,
		},
		Organization: organization,
	}
}
