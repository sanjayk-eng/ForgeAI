package github

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"ai-agent/internal/modules/auth/provider"
	"ai-agent/internal/modules/project/repository"
)

type Client interface {
	InspectRepository(ctx context.Context, owner, name string) (Repository, error)
	InspectRepositoryWithToken(ctx context.Context, owner, name, accessToken string) (Repository, error)
	ResolveRepository(ctx context.Context, repositoryURL, accessToken string) (repository.ResolvedRepository, error)
}

type Catalog interface {
	ListOrganizations(ctx context.Context, accessToken string) ([]string, error)
	ListRepositories(ctx context.Context, accessToken, organization string) ([]Repository, error)
}

type Repository struct {
	ID            int64
	Owner         string
	Name          string
	URL           string
	Private       bool
	DefaultBranch string
	Branches      []string
}

type RepositoryOption struct {
	repository.ConnectRepositoryRequest
	Organization string `json:"organization"`
}

type CatalogResponse struct {
	Organizations []string           `json:"organizations"`
	Repositories  []RepositoryOption `json:"repositories"`
}

type client struct {
	provider provider.GitHubRepositoryInspector
	catalog  provider.GitHubRepositoryCatalog
}

func NewClient(providerClient provider.GitHubRepositoryInspector) Client {
	catalog, _ := providerClient.(provider.GitHubRepositoryCatalog)
	return &client{provider: providerClient, catalog: catalog}
}

func (c *client) InspectRepository(ctx context.Context, owner, name string) (Repository, error) {
	return c.inspectRepository(ctx, owner, name, "")
}

func (c *client) InspectRepositoryWithToken(ctx context.Context, owner, name, accessToken string) (Repository, error) {
	return c.inspectRepository(ctx, owner, name, accessToken)
}

func (c *client) inspectRepository(ctx context.Context, owner, name, accessToken string) (Repository, error) {
	if c.provider == nil {
		return Repository{}, fmt.Errorf("GitHub repository provider is not configured")
	}

	var repo provider.GitHubRepository
	var err error

	if tokenClient, ok := c.provider.(provider.GitHubRepositoryTokenInspector); ok {
		repo, err = tokenClient.InspectRepositoryWithToken(ctx, owner, name, accessToken)
	} else {
		repo, err = c.provider.InspectRepository(ctx, owner, name)
	}

	if err != nil {
		return Repository{}, err
	}

	return Repository{
		ID:            repo.ID,
		Owner:         repo.Owner,
		Name:          repo.Name,
		URL:           repo.URL,
		Private:       repo.Private,
		DefaultBranch: repo.DefaultBranch,
		Branches:      repo.Branches,
	}, nil
}

func (c *client) ResolveRepository(ctx context.Context, repositoryURL, accessToken string) (repository.ResolvedRepository, error) {
	parsed, err := url.Parse(strings.TrimSpace(repositoryURL))
	if err != nil || !strings.EqualFold(parsed.Scheme, "https") || !strings.EqualFold(parsed.Host, "github.com") {
		return repository.ResolvedRepository{}, fmt.Errorf("invalid repository URL")
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return repository.ResolvedRepository{}, fmt.Errorf("invalid repository URL format")
	}

	name := strings.TrimSuffix(parts[1], ".git")
	repo, err := c.InspectRepositoryWithToken(ctx, parts[0], name, accessToken)
	if err != nil {
		return repository.ResolvedRepository{}, fmt.Errorf("resolve GitHub repository: %w", err)
	}

	return repository.ResolvedRepository{
		Repository: repository.ConnectRepositoryRequest{
			GitHubRepositoryID:   repo.ID,
			GitHubOwner:          repo.Owner,
			GitHubRepositoryName: repo.Name,
			RepositoryURL:        repo.URL,
			Private:              repo.Private,
			DefaultBranch:        repo.DefaultBranch,
		},
		Branches: repo.Branches,
	}, nil
}

func (c *client) ListOrganizations(ctx context.Context, accessToken string) ([]string, error) {
	if c.catalog == nil {
		return nil, fmt.Errorf("GitHub repository catalog is not configured")
	}

	organizations, err := c.catalog.ListOrganizations(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(organizations))
	for _, org := range organizations {
		if org.Login != "" {
			result = append(result, org.Login)
		}
	}
	return result, nil
}

func (c *client) ListRepositories(ctx context.Context, accessToken, organization string) ([]Repository, error) {
	if c.catalog == nil {
		return nil, fmt.Errorf("GitHub repository catalog is not configured")
	}

	repositories, err := c.catalog.ListRepositories(ctx, accessToken, organization)
	if err != nil {
		return nil, err
	}

	result := make([]Repository, 0, len(repositories))
	for _, repo := range repositories {
		result = append(result, Repository{
			ID:            repo.ID,
			Owner:         repo.Owner,
			Name:          repo.Name,
			URL:           repo.URL,
			Private:       repo.Private,
			DefaultBranch: repo.DefaultBranch,
			Branches:      repo.Branches,
		})
	}
	return result, nil
}
