package provider

import (
	"context"
)

type ServiceProvider interface {
	ExchangeCode(ctx context.Context, code string) (ServiceUser, error)
}

type GitHubRepositoryInspector interface {
	InspectRepository(ctx context.Context, owner, name string) (GitHubRepository, error)
}

type GitHubRepositoryCatalog interface {
	ListOrganizations(ctx context.Context, accessToken string) ([]GitHubOrganization, error)
	ListRepositories(ctx context.Context, accessToken, organization string) ([]GitHubRepository, error)
}

var _ ServiceProvider = (*GoogleProvider)(nil)
var _ ServiceProvider = (*GitHubProvider)(nil)
