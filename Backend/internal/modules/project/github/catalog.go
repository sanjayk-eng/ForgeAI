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
	ListRepositories(ctx context.Context, accessToken string) (CatalogResponse, error)
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
	ID         string                         `json:"id"`
	Name       string                         `json:"name"`
	Repository repository.ProjectRepository `json:"repository"`
}

type catalogService struct {
	client Client
}

func NewCatalogService(client Client) CatalogService {
	return &catalogService{client: client}
}

func (s *catalogService) ListRepositories(ctx context.Context, accessToken string) (CatalogResponse, error) {
	catalog, ok := s.client.(Catalog)
	if !ok {
		return CatalogResponse{}, ErrGitHubCatalogUnavailable
	}

	organizations, err := catalog.ListOrganizations(ctx, accessToken)
	if err != nil {
		return CatalogResponse{}, err
	}

	result := CatalogResponse{
		Organizations: organizations,
		Repositories:  []RepositoryOption{},
	}

	seen := make(map[int64]bool)

	// List personal repositories
	personalRepositories, err := catalog.ListRepositories(ctx, accessToken, "")
	if err != nil {
		return CatalogResponse{}, err
	}

	for _, repo := range personalRepositories {
		if seen[repo.ID] {
			continue
		}
		seen[repo.ID] = true
		result.Repositories = append(result.Repositories, toRepositoryOption(repo, "personal"))
	}

	// List organization repositories
	for _, org := range organizations {
		repositories, listErr := catalog.ListRepositories(ctx, accessToken, org)
		if listErr != nil {
			return CatalogResponse{}, listErr
		}
		for _, repo := range repositories {
			if seen[repo.ID] {
				continue
			}
			seen[repo.ID] = true
			result.Repositories = append(result.Repositories, toRepositoryOption(repo, org))
		}
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
