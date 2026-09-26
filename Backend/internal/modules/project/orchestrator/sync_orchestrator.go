package orchestrator

import (
	"context"
	"errors"

	"ai-agent/internal/middleware"
	"ai-agent/internal/modules/project/core"
	"ai-agent/internal/modules/project/sync"
	"ai-agent/internal/shared/pagination"
)

// SyncOrchestrator handles GitHub sync operations
type SyncOrchestrator struct {
	syncService    *sync.Service
	projectOrch    *ProjectOrchestrator
	coreService    core.Service
	accountStore   GitHubAccountStore
}

type GitHubAccountStore interface {
	FindGitHubAccessToken(ctx context.Context, userID string) (string, error)
}

func NewSyncOrchestrator(
	syncService *sync.Service,
	projectOrch *ProjectOrchestrator,
	coreService core.Service,
	accountStore GitHubAccountStore,
) *SyncOrchestrator {
	return &SyncOrchestrator{
		syncService:  syncService,
		projectOrch:  projectOrch,
		coreService:  coreService,
		accountStore: accountStore,
	}
}

func (o *SyncOrchestrator) SyncProject(ctx context.Context, projectID string) (ProjectWithRepository, error) {
	if o.syncService == nil {
		return ProjectWithRepository{}, errors.New("sync service not available")
	}

	accessToken := o.getAccessToken(ctx)

	if err := o.syncService.SyncProjectWithToken(ctx, projectID, accessToken); err != nil {
		return ProjectWithRepository{}, err
	}

	return o.projectOrch.FindByID(ctx, projectID)
}

func (o *SyncOrchestrator) SyncWorkspaceProjects(ctx context.Context, workspaceID string) (sync.WorkspaceProjectsSyncResponse, error) {
	if o.syncService == nil {
		return sync.WorkspaceProjectsSyncResponse{}, errors.New("sync service not available")
	}

	result := sync.WorkspaceProjectsSyncResponse{}
	query := pagination.Query{Page: 1, PerPage: pagination.MaxPerPage}

	for {
		projects, err := o.coreService.ListByWorkspace(ctx, workspaceID, query)
		if err != nil {
			return sync.WorkspaceProjectsSyncResponse{}, err
		}

		for _, project := range projects.Items {
			if project.Type != core.ProjectTypeRepository {
				continue
			}

			if err := o.syncService.SyncProject(ctx, project.ID); err != nil {
				result.Failed++
				continue
			}
			result.Triggered++
		}

		if query.Page >= projects.TotalPages || projects.TotalPages == 0 {
			break
		}
		query.Page++
	}

	return result, nil
}

func (o *SyncOrchestrator) getAccessToken(ctx context.Context) string {
	if o.accountStore == nil {
		return ""
	}

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return ""
	}

	token, _ := o.accountStore.FindGitHubAccessToken(ctx, userID)
	return token
}
