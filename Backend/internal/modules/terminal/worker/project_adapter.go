package worker

import (
	"context"

	projectcore "ai-agent/internal/modules/project/core"
	projectrepo "ai-agent/internal/modules/project/repository"
)

type ProjectAdapter struct {
	coreService projectcore.Service
	repoService projectrepo.Service
	tokenStore  GitHubTokenStore
}

type GitHubTokenStore interface {
	FindGitHubAccessToken(ctx context.Context, userID string) (string, error)
}

func NewProjectAdapter(coreService projectcore.Service, repoService projectrepo.Service, tokenStore GitHubTokenStore) *ProjectAdapter {
	return &ProjectAdapter{
		coreService: coreService,
		repoService: repoService,
		tokenStore:  tokenStore,
	}
}

func (a *ProjectAdapter) FindByID(ctx context.Context, projectID string) (Project, error) {
	coreProject, err := a.coreService.FindByID(ctx, projectID)
	if err != nil {
		return Project{}, err
	}

	project := Project{
		ID:     coreProject.ID,
		Type:   string(coreProject.Type),
		UserID: coreProject.CreatedBy,
	}

	if coreProject.Type == projectcore.ProjectTypeRepository {
		repo, err := a.repoService.FindByProjectID(ctx, projectID)
		if err == nil {
			project.Repository = &Repository{
				URL:    repo.RepositoryURL,
				Branch: repo.DefaultBranch,
			}
			if a.tokenStore != nil {
				project.Repository.AccessToken, _ = a.tokenStore.FindGitHubAccessToken(ctx, coreProject.CreatedBy)
			}
		}
	}

	return project, nil
}
