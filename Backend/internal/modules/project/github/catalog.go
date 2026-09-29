package github

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrGitHubAccountUnavailable = errors.New("GitHub account is not connected; sign in with GitHub again")
	ErrGitHubCatalogUnavailable = errors.New("GitHub repository catalog is not configured")
	ErrWorkspaceOwnerRequired   = errors.New("only the workspace owner can access GitHub repositories")
)

type CatalogService interface {
	ListRepositories(ctx context.Context, accessToken, owner string) (CatalogResponse, error)
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

func toRepositoryOption(repo Repository, organization string) RepositoryOption {
	return RepositoryOption{
		ConnectRepositoryRequest: toConnectRepositoryRequest(repo),
		Organization:             organization,
	}
}
