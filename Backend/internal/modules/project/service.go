package project

import (
	"context"

	"ai-agent/internal/modules/project/core"
	"ai-agent/internal/modules/project/github"
	"ai-agent/internal/modules/project/orchestrator"
	projectrepo "ai-agent/internal/modules/project/repository"
	"ai-agent/internal/modules/project/sync"
	"ai-agent/internal/shared/pagination"

	"github.com/jmoiron/sqlx"
)

// Service is the main facade for all project operations
type Service interface {
	// Core project operations
	Create(ctx context.Context, workspaceID, createdBy string, input CreateProjectRequest) (orchestrator.ProjectWithRepository, error)
	ListByWorkspace(ctx context.Context, workspaceID string, query pagination.Query) (pagination.Result[orchestrator.ProjectWithRepository], error)
	FindByID(ctx context.Context, projectID string) (orchestrator.ProjectWithRepository, error)
	FindByWorkspaceSlug(ctx context.Context, workspaceID, slug string) (orchestrator.ProjectWithRepository, error)
	Update(ctx context.Context, projectID, userID string, input UpdateProjectRequest) (orchestrator.ProjectWithRepository, error)
	Delete(ctx context.Context, projectID, userID string) error

	// Repository operations
	ConnectRepository(ctx context.Context, projectID string, input ConnectRepositoryRequest) (projectrepo.ProjectRepository, error)
	UpdateRepositoryBranch(ctx context.Context, projectID, branch string) (projectrepo.ProjectRepository, error)

	// Sync operations
	SyncProject(ctx context.Context, projectID string) (orchestrator.ProjectWithRepository, error)
	SyncWorkspaceProjects(ctx context.Context, workspaceID string) (sync.WorkspaceProjectsSyncResponse, error)

	// GitHub operations
	ResolveRepository(ctx context.Context, repositoryURL string) (projectrepo.ResolvedRepository, error)
	ListGitHubRepositories(ctx context.Context, workspaceID, userID, owner string) (github.CatalogResponse, error)
	ImportGitHubRepositories(ctx context.Context, workspaceID, userID string, input ImportGitHubRepositoriesRequest) (ImportGitHubRepositoriesResponse, error)
}

type service struct {
	db          *sqlx.DB
	core        core.Service
	repoService projectrepo.Service
	projectOrch *orchestrator.ProjectOrchestrator
	githubOrch  *orchestrator.GitHubOrchestrator
	syncOrch    *orchestrator.SyncOrchestrator
}

func NewService(
	db *sqlx.DB,
	coreService core.Service,
	repoService projectrepo.Service,
	projectOrch *orchestrator.ProjectOrchestrator,
	githubOrch *orchestrator.GitHubOrchestrator,
	syncOrch *orchestrator.SyncOrchestrator,
) Service {
	return &service{
		db:          db,
		core:        coreService,
		repoService: repoService,
		projectOrch: projectOrch,
		githubOrch:  githubOrch,
		syncOrch:    syncOrch,
	}
}

func (s *service) Create(ctx context.Context, workspaceID, createdBy string, input CreateProjectRequest) (orchestrator.ProjectWithRepository, error) {
	return s.projectOrch.Create(ctx, workspaceID, createdBy, orchestrator.CreateProjectInput{
		Name:        input.Name,
		Description: input.Description,
		Type:        core.ProjectType(input.Type),
		Repository:  input.Repository,
	})
}

func (s *service) ListByWorkspace(ctx context.Context, workspaceID string, query pagination.Query) (pagination.Result[orchestrator.ProjectWithRepository], error) {
	return s.projectOrch.ListByWorkspace(ctx, workspaceID, query)
}

func (s *service) FindByID(ctx context.Context, projectID string) (orchestrator.ProjectWithRepository, error) {
	return s.projectOrch.FindByID(ctx, projectID)
}

func (s *service) FindByWorkspaceSlug(ctx context.Context, workspaceID, slug string) (orchestrator.ProjectWithRepository, error) {
	return s.projectOrch.FindByWorkspaceSlug(ctx, workspaceID, slug)
}

func (s *service) Update(ctx context.Context, projectID, userID string, input UpdateProjectRequest) (orchestrator.ProjectWithRepository, error) {
	return s.projectOrch.Update(ctx, projectID, userID, core.UpdateProjectRequest{
		Name:        input.Name,
		Description: input.Description,
		Status:      core.ProjectStatus(input.Status),
	})
}

func (s *service) Delete(ctx context.Context, projectID, userID string) error {
	return s.core.Delete(ctx, projectID, userID)
}

func (s *service) ConnectRepository(ctx context.Context, projectID string, input ConnectRepositoryRequest) (projectrepo.ProjectRepository, error) {
	project, err := s.core.FindByID(ctx, projectID)
	if err != nil {
		return projectrepo.ProjectRepository{}, core.ErrProjectNotFound
	}

	return s.repoService.ConnectRepository(ctx, nil, projectID, project.WorkspaceID, projectrepo.ConnectRepositoryRequest(input))
}

func (s *service) UpdateRepositoryBranch(ctx context.Context, projectID, branch string) (projectrepo.ProjectRepository, error) {
	return s.repoService.UpdateBranch(ctx, projectID, branch)
}

func (s *service) SyncProject(ctx context.Context, projectID string) (orchestrator.ProjectWithRepository, error) {
	return s.syncOrch.SyncProject(ctx, projectID)
}

func (s *service) SyncWorkspaceProjects(ctx context.Context, workspaceID string) (sync.WorkspaceProjectsSyncResponse, error) {
	return s.syncOrch.SyncWorkspaceProjects(ctx, workspaceID)
}

func (s *service) ResolveRepository(ctx context.Context, repositoryURL string) (projectrepo.ResolvedRepository, error) {
	return s.githubOrch.ResolveRepository(ctx, repositoryURL)
}

func (s *service) ListGitHubRepositories(ctx context.Context, workspaceID, userID, owner string) (github.CatalogResponse, error) {
	return s.githubOrch.ListGitHubRepositories(ctx, workspaceID, userID, owner)
}

func (s *service) ImportGitHubRepositories(ctx context.Context, workspaceID, userID string, input ImportGitHubRepositoriesRequest) (ImportGitHubRepositoriesResponse, error) {
	result, err := s.githubOrch.ImportGitHubRepositories(ctx, workspaceID, userID, orchestrator.ImportGitHubRepositoriesInput{
		Repositories: input.Repositories,
	})
	if err != nil {
		return ImportGitHubRepositoriesResponse{}, err
	}

	return ImportGitHubRepositoriesResponse{
		Projects: result.Projects,
		Skipped:  result.Skipped,
	}, nil
}
