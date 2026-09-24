package project

import (
	"context"
	"fmt"

	"ai-agent/internal/modules/auth/provider"
)

type GitHubRepositoryClient interface {
	InspectRepository(ctx context.Context, owner, name string) (GitHubRepository, error)
}

type GitHubRepository struct {
	ID            int64
	Owner         string
	Name          string
	URL           string
	DefaultBranch string
}

type githubRepositoryClient struct {
	provider provider.GitHubRepositoryInspector
}

func NewGitHubRepositoryClient(client provider.GitHubRepositoryInspector) GitHubRepositoryClient {
	return &githubRepositoryClient{provider: client}
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
		URL: repository.URL, DefaultBranch: repository.DefaultBranch,
	}, nil
}
