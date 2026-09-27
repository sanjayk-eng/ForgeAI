package worker

import (
	"context"

	projectcore "ai-agent/internal/modules/project/core"
	projectrepo "ai-agent/internal/modules/project/repository"
)

type ProjectAdapter struct {
	coreService projectcore.Service
	repoService projectrepo.Service
}

func NewProjectAdapter(coreService projectcore.Service, repoService projectrepo.Service) *ProjectAdapter {
	return &ProjectAdapter{
		coreService: coreService,
		repoService: repoService,
	}
}

func (a *ProjectAdapter) FindByID(ctx context.Context, projectID string) (Project, error) {
	coreProject, err := a.coreService.FindByID(ctx, projectID)
	if err != nil {
		return Project{}, err
	}

	project := Project{
		ID:   coreProject.ID,
		Type: string(coreProject.Type),
	}

	if coreProject.Type == projectcore.ProjectTypeRepository {
		repo, err := a.repoService.FindByProjectID(ctx, projectID)
		if err == nil {
			project.Repository = &Repository{
				URL:    repo.RepositoryURL,
				Branch: repo.DefaultBranch,
			}
		}
	}

	return project, nil
}
