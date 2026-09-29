package provider

import (
	"context"
	"errors"
)

var ErrGitHubOrganizationScopeRequired = errors.New("GitHub access token is missing the read:org permission; reconnect your GitHub account")

type ServiceProvider interface {
	ExchangeCode(ctx context.Context, code string) (ServiceUser, error)
}

type GitHubRepositoryInspector interface {
	InspectRepository(ctx context.Context, owner, name string) (GitHubRepository, error)
}

type GitHubRepositoryTokenInspector interface {
	InspectRepositoryWithToken(ctx context.Context, owner, name, accessToken string) (GitHubRepository, error)
}

type GitHubRepositoryCatalog interface {
	GetAccountLogin(ctx context.Context, accessToken string) (string, error)
	ListOrganizations(ctx context.Context, accessToken string) ([]GitHubOrganization, error)
	ListRepositories(ctx context.Context, accessToken, organization string) ([]GitHubRepository, error)
}

var _ ServiceProvider = (*GoogleProvider)(nil)
var _ ServiceProvider = (*GitHubProvider)(nil)
