package github

import (
	"ai-agent/internal/modules/auth/provider"
	"ai-agent/internal/modules/project/repository"
)

func mapRepository(repo provider.GitHubRepository) Repository {
	return Repository{
		ID:            repo.ID,
		Owner:         repo.Owner,
		Name:          repo.Name,
		URL:           repo.URL,
		Private:       repo.Private,
		DefaultBranch: repo.DefaultBranch,
		Branches:      repo.Branches,
	}
}

func toConnectRepositoryRequest(repo Repository) repository.ConnectRepositoryRequest {
	return repository.ConnectRepositoryRequest{
		GitHubRepositoryID:   repo.ID,
		GitHubOwner:          repo.Owner,
		GitHubRepositoryName: repo.Name,
		RepositoryURL:        repo.URL,
		Private:              repo.Private,
		DefaultBranch:        repo.DefaultBranch,
	}
}
