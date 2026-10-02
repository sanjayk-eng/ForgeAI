package worker

import (
	"context"

	projectcore "ai-agent/internal/modules/project/core"
)

type ProjectAdapter struct {
	coreService projectcore.Service
}

func NewProjectAdapter(coreService projectcore.Service) *ProjectAdapter {
	return &ProjectAdapter{
		coreService: coreService,
	}
}

func (a *ProjectAdapter) FindByID(ctx context.Context, projectID string) (Project, error) {
	coreProject, err := a.coreService.FindByID(ctx, projectID)
	if err != nil {
		return Project{}, err
	}

	project := Project{
		ID:     coreProject.ID,
		Type:   "MANUAL", // All projects are manual now
		UserID: coreProject.CreatedBy,
	}

	return project, nil
}
