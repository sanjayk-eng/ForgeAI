package project

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	appdatabase "ai-agent/pkg/database"

	"github.com/jmoiron/sqlx"
)

type Service interface {
	Create(ctx context.Context, workspaceID, createdBy string, input CreateProjectRequest) (Project, error)
	ListByWorkspace(ctx context.Context, workspaceID string) ([]Project, error)
	FindByID(ctx context.Context, projectID string) (Project, error)
	FindByWorkspaceSlug(ctx context.Context, workspaceID, slug string) (Project, error)
	Update(ctx context.Context, projectID string, input UpdateProjectRequest) (Project, error)
	ConnectRepository(ctx context.Context, projectID string, input ConnectRepositoryRequest) (ProjectRepository, error)
	SyncProject(ctx context.Context, projectID string) (Project, error)
	ResolveRepository(ctx context.Context, repositoryURL string) (ConnectRepositoryRequest, error)
}

type service struct {
	db     *sqlx.DB
	repo   ProjectRepositoryStore
	sync   *SyncService
	github GitHubRepositoryClient
}

func NewService(db *sqlx.DB, repo ProjectRepositoryStore) Service {
	return &service{db: db, repo: repo}
}

func NewServiceWithSync(db *sqlx.DB, repo ProjectRepositoryStore, syncService *SyncService, github GitHubRepositoryClient) Service {
	return &service{db: db, repo: repo, sync: syncService, github: github}
}

func (service *service) Create(ctx context.Context, workspaceID, createdBy string, input CreateProjectRequest) (Project, error) {
	workspaceID = strings.TrimSpace(workspaceID)
	createdBy = strings.TrimSpace(createdBy)
	name := strings.TrimSpace(input.Name)
	if service.db == nil || workspaceID == "" || createdBy == "" || name == "" {
		return Project{}, ErrInvalidProjectInput
	}
	if input.Type == ProjectTypeRepository && input.Repository == nil {
		return Project{}, ErrInvalidProjectInput
	}
	if input.Type == ProjectTypeEmpty && input.Repository != nil {
		return Project{}, ErrInvalidProjectInput
	}

	var project Project
	err := appdatabase.WithTx(ctx, service.db, func(ctx context.Context, tx *sqlx.Tx) error {
		var err error
		project, err = service.repo.Create(ctx, tx, workspaceID, name, slugify(name), input.Description, string(input.Type), createdBy)
		if err != nil {
			return err
		}
		if input.Repository != nil {
			var repository ProjectRepository
			repository, err = service.repo.ConnectRepository(ctx, tx, project.ID, *input.Repository)
			project.Repository = &repository
		}
		return err
	})
	if err != nil {
		return Project{}, fmt.Errorf("create project: %w", err)
	}
	return project, nil
}

func (service *service) ListByWorkspace(ctx context.Context, workspaceID string) ([]Project, error) {
	if strings.TrimSpace(workspaceID) == "" {
		return nil, ErrInvalidProjectInput
	}
	return service.repo.ListByWorkspace(ctx, workspaceID)
}

func (service *service) FindByID(ctx context.Context, projectID string) (Project, error) {
	if strings.TrimSpace(projectID) == "" {
		return Project{}, ErrInvalidProjectInput
	}
	project, err := service.repo.FindByID(ctx, projectID)
	if err != nil {
		return Project{}, ErrProjectNotFound
	}
	return project, nil
}

func (service *service) FindByWorkspaceSlug(ctx context.Context, workspaceID, slug string) (Project, error) {
	if strings.TrimSpace(workspaceID) == "" || strings.TrimSpace(slug) == "" {
		return Project{}, ErrInvalidProjectInput
	}
	project, err := service.repo.FindByWorkspaceSlug(ctx, workspaceID, slug)
	if err != nil {
		return Project{}, ErrProjectNotFound
	}
	return project, nil
}

func (service *service) Update(ctx context.Context, projectID string, input UpdateProjectRequest) (Project, error) {
	projectID = strings.TrimSpace(projectID)
	name := strings.TrimSpace(input.Name)
	if service.db == nil || projectID == "" || name == "" {
		return Project{}, ErrInvalidProjectInput
	}
	project, err := service.repo.Update(ctx, projectID, name, input.Description, string(input.Status))
	if err != nil {
		return Project{}, ErrProjectNotFound
	}
	return project, nil
}

func (service *service) ConnectRepository(ctx context.Context, projectID string, input ConnectRepositoryRequest) (ProjectRepository, error) {
	if service.db == nil || strings.TrimSpace(projectID) == "" || input.GitHubRepositoryID <= 0 || strings.TrimSpace(input.GitHubOwner) == "" || strings.TrimSpace(input.GitHubRepositoryName) == "" || strings.TrimSpace(input.RepositoryURL) == "" || strings.TrimSpace(input.DefaultBranch) == "" {
		return ProjectRepository{}, ErrInvalidProjectInput
	}
	var repository ProjectRepository
	err := appdatabase.WithTx(ctx, service.db, func(ctx context.Context, tx *sqlx.Tx) error {
		var err error
		repository, err = service.repo.ConnectRepository(ctx, tx, projectID, input)
		return err
	})
	if err != nil {
		return ProjectRepository{}, fmt.Errorf("connect repository: %w", err)
	}
	return repository, nil
}

func (service *service) SyncProject(ctx context.Context, projectID string) (Project, error) {
	if service.sync == nil || strings.TrimSpace(projectID) == "" {
		return Project{}, ErrInvalidProjectInput
	}
	if err := service.sync.SyncProject(ctx, projectID); err != nil {
		return Project{}, err
	}
	return service.FindByID(ctx, projectID)
}

func (service *service) ResolveRepository(ctx context.Context, repositoryURL string) (ConnectRepositoryRequest, error) {
	parsed, err := url.Parse(strings.TrimSpace(repositoryURL))
	if err != nil || !strings.EqualFold(parsed.Scheme, "https") || !strings.EqualFold(parsed.Host, "github.com") || service.github == nil {
		return ConnectRepositoryRequest{}, ErrInvalidProjectInput
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return ConnectRepositoryRequest{}, ErrInvalidProjectInput
	}
	name := strings.TrimSuffix(parts[1], ".git")
	repository, err := service.github.InspectRepository(ctx, parts[0], name)
	if err != nil {
		return ConnectRepositoryRequest{}, fmt.Errorf("resolve GitHub repository: %w", err)
	}
	return ConnectRepositoryRequest{
		GitHubRepositoryID:   repository.ID,
		GitHubOwner:          repository.Owner,
		GitHubRepositoryName: repository.Name,
		RepositoryURL:        repository.URL,
		DefaultBranch:        repository.DefaultBranch,
	}, nil
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(value, "-")
	return strings.Trim(strings.TrimSpace(value), "-")
}
