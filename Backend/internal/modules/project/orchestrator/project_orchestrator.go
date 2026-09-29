package orchestrator

import (
	"context"
	"fmt"

	"ai-agent/internal/modules/project/core"
	projectrepo "ai-agent/internal/modules/project/repository"
	"ai-agent/internal/shared/pagination"
	appdatabase "ai-agent/pkg/database"

	"github.com/jmoiron/sqlx"
)

// ProjectOrchestrator handles project + repository composite operations
type ProjectOrchestrator struct {
	db              *sqlx.DB
	coreService     core.Service
	repoService     projectrepo.Service
	onCreate        func(ctx context.Context, projectID, userID string)
	onDelete        func(ctx context.Context, projectID string) error
	onBranchUpdated func(ctx context.Context, projectID string) error
}

func NewProjectOrchestrator(db *sqlx.DB, coreService core.Service, repoService projectrepo.Service) *ProjectOrchestrator {
	return &ProjectOrchestrator{
		db:          db,
		coreService: coreService,
		repoService: repoService,
	}
}

func (o *ProjectOrchestrator) SetOnCreate(fn func(ctx context.Context, projectID, userID string)) {
	o.onCreate = fn
}

func (o *ProjectOrchestrator) SetOnDelete(fn func(ctx context.Context, projectID string) error) {
	o.onDelete = fn
}

func (o *ProjectOrchestrator) SetOnBranchUpdated(fn func(ctx context.Context, projectID string) error) {
	o.onBranchUpdated = fn
}

func (o *ProjectOrchestrator) TriggerOnCreate(ctx context.Context, projectID, userID string) {
	if o.onCreate != nil {
		o.onCreate(ctx, projectID, userID)
	}
}

func (o *ProjectOrchestrator) TriggerOnDelete(ctx context.Context, projectID string) error {
	if o.onDelete != nil {
		return o.onDelete(ctx, projectID)
	}
	return nil
}

func (o *ProjectOrchestrator) TriggerOnBranchUpdated(ctx context.Context, projectID string) error {
	if o.onBranchUpdated != nil {
		return o.onBranchUpdated(ctx, projectID)
	}
	return nil
}

type ProjectWithRepository struct {
	core.Project
	Repository *projectrepo.ProjectRepository `json:"repository,omitempty"`
}

func (o *ProjectOrchestrator) Create(ctx context.Context, workspaceID, createdBy string, input CreateProjectInput) (ProjectWithRepository, error) {
	if input.Type == core.ProjectTypeRepository && input.Repository == nil {
		return ProjectWithRepository{}, core.ErrInvalidProjectInput
	}
	if input.Type == core.ProjectTypeEmpty && input.Repository != nil {
		return ProjectWithRepository{}, core.ErrInvalidProjectInput
	}

	var result ProjectWithRepository
	err := appdatabase.WithTx(ctx, o.db, func(ctx context.Context, tx *sqlx.Tx) error {
		project, err := o.coreService.Create(ctx, tx, workspaceID, createdBy, core.CreateProjectRequest{
			Name:        input.Name,
			Description: input.Description,
			Type:        input.Type,
		})
		if err != nil {
			return err
		}

		result = ProjectWithRepository{Project: project}

		if input.Repository != nil {
			repository, err := o.repoService.ConnectRepository(ctx, tx, project.ID, workspaceID, *input.Repository)
			if err != nil {
				return err
			}
			result.Repository = &repository
		}

		return nil
	})

	if err != nil {
		return ProjectWithRepository{}, fmt.Errorf("create project: %w", err)
	}

	o.TriggerOnCreate(ctx, result.ID, createdBy)

	return result, nil
}

func (o *ProjectOrchestrator) ListByWorkspace(ctx context.Context, workspaceID string, query pagination.Query) (pagination.Result[ProjectWithRepository], error) {
	projects, err := o.coreService.ListByWorkspace(ctx, workspaceID, query)
	if err != nil {
		return pagination.Result[ProjectWithRepository]{}, err
	}

	enrichedProjects := make([]ProjectWithRepository, 0, len(projects.Items))
	for _, project := range projects.Items {
		enriched := ProjectWithRepository{Project: project}

		if project.Type == core.ProjectTypeRepository {
			if repo, err := o.repoService.FindByProjectID(ctx, project.ID); err == nil {
				enriched.Repository = &repo
			}
		}

		enrichedProjects = append(enrichedProjects, enriched)
	}

	return pagination.NewResult(enrichedProjects, query, projects.Total), nil
}

func (o *ProjectOrchestrator) FindByID(ctx context.Context, projectID string) (ProjectWithRepository, error) {
	project, err := o.coreService.FindByID(ctx, projectID)
	if err != nil {
		return ProjectWithRepository{}, err
	}

	return o.enrichProject(ctx, project), nil
}

func (o *ProjectOrchestrator) FindByWorkspaceSlug(ctx context.Context, workspaceID, slug string) (ProjectWithRepository, error) {
	project, err := o.coreService.FindByWorkspaceSlug(ctx, workspaceID, slug)
	if err != nil {
		return ProjectWithRepository{}, err
	}

	return o.enrichProject(ctx, project), nil
}

func (o *ProjectOrchestrator) Update(ctx context.Context, projectID, userID string, input core.UpdateProjectRequest) (ProjectWithRepository, error) {
	project, err := o.coreService.Update(ctx, projectID, userID, input)
	if err != nil {
		return ProjectWithRepository{}, err
	}

	return o.enrichProject(ctx, project), nil
}

func (o *ProjectOrchestrator) enrichProject(ctx context.Context, project core.Project) ProjectWithRepository {
	result := ProjectWithRepository{Project: project}

	if project.Type == core.ProjectTypeRepository {
		if repo, err := o.repoService.FindByProjectID(ctx, project.ID); err == nil {
			result.Repository = &repo
		}
	}

	return result
}

type CreateProjectInput struct {
	Name        string
	Description *string
	Type        core.ProjectType
	Repository  *projectrepo.ConnectRepositoryRequest
}
