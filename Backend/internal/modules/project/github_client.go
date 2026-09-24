package project

import (
	"context"
	"fmt"

	"ai-agent/internal/modules/auth/provider"
)

type GitHubRepositoryClient interface {
	InspectRepository(ctx context.Context, owner, name string) (GitHubRepository, error)
}

type GitHubRepositoryCatalog interface {
	ListOrganizations(ctx context.Context, accessToken string) ([]string, error)
	ListRepositories(ctx context.Context, accessToken, organization string) ([]GitHubRepository, error)
}

type GitHubRepository struct {
	ID            int64
	Owner         string
	Name          string
	URL           string
	DefaultBranch string
	Branches      []string
}

type githubRepositoryClient struct {
	provider provider.GitHubRepositoryInspector
	catalog  provider.GitHubRepositoryCatalog
}

func NewGitHubRepositoryClient(client provider.GitHubRepositoryInspector) GitHubRepositoryClient {
	catalog, _ := client.(provider.GitHubRepositoryCatalog)
	return &githubRepositoryClient{provider: client, catalog: catalog}
}

func (client *githubRepositoryClient) InspectRepository(ctx context.Context, owner, name string) (GitHubRepository, error) {
	if client.provider == nil {
		return GitHubRepository{}, fmt.Errorf("GitHub repository provider is not configured")
	}
	repository, err := client.provider.InspectRepository(ctx, owner, name)
	if err != nil {
		return GitHubRepository{}, err
	}
	return GitHubRepository{
		ID: repository.ID, Owner: repository.Owner, Name: repository.Name,
		URL: repository.URL, DefaultBranch: repository.DefaultBranch, Branches: repository.Branches,
	}, nil
}

func (client *githubRepositoryClient) ListOrganizations(ctx context.Context, accessToken string) ([]string, error) {
	if client.catalog == nil {
		return nil, fmt.Errorf("GitHub repository catalog is not configured")
	}
	organizations, err := client.catalog.ListOrganizations(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(organizations))
	for _, organization := range organizations {
		if organization.Login != "" {
			result = append(result, organization.Login)
		}
	}
	return result, nil
}

func (client *githubRepositoryClient) ListRepositories(ctx context.Context, accessToken, organization string) ([]GitHubRepository, error) {
	if client.catalog == nil {
		return nil, fmt.Errorf("GitHub repository catalog is not configured")
	}
	repositories, err := client.catalog.ListRepositories(ctx, accessToken, organization)
	if err != nil {
		return nil, err
	}
	result := make([]GitHubRepository, 0, len(repositories))
	for _, repository := range repositories {
		result = append(result, GitHubRepository{
			ID: repository.ID, Owner: repository.Owner, Name: repository.Name,
			URL: repository.URL, DefaultBranch: repository.DefaultBranch, Branches: repository.Branches,
		})
	}
	return result, nil
}
