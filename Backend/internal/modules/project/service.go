package project

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"ai-agent/internal/shared/pagination"
	appdatabase "ai-agent/pkg/database"

	"github.com/jmoiron/sqlx"
)

type Service interface {
	Create(ctx context.Context, workspaceID, createdBy string, input CreateProjectRequest) (Project, error)
	ListByWorkspace(ctx context.Context, workspaceID string, query pagination.Query) (pagination.Result[Project], error)
	FindByID(ctx context.Context, projectID string) (Project, error)
	FindByWorkspaceSlug(ctx context.Context, workspaceID, slug string) (Project, error)
	Update(ctx context.Context, projectID string, input UpdateProjectRequest) (Project, error)
	UpdateRepositoryBranch(ctx context.Context, projectID, branch string) (ProjectRepository, error)
	ConnectRepository(ctx context.Context, projectID string, input ConnectRepositoryRequest) (ProjectRepository, error)
	SyncProject(ctx context.Context, projectID string) (Project, error)
	SyncWorkspaceProjects(ctx context.Context, workspaceID string) (SyncWorkspaceProjectsResponse, error)
	ResolveRepository(ctx context.Context, repositoryURL string) (ResolvedRepository, error)
	ListGitHubRepositories(ctx context.Context, workspaceID, userID string) (GitHubRepositoryCatalogResponse, error)
	ImportGitHubRepositories(ctx context.Context, workspaceID, userID string, input ImportGitHubRepositoriesRequest) (ImportGitHubRepositoriesResponse, error)
}

type service struct {
	db      *sqlx.DB
	repo    ProjectRepositoryStore
	sync    *SyncService
	github  GitHubRepositoryClient
	account GitHubAccountStore
	catalog GitHubRepositoryCatalog
}

type GitHubAccountStore interface {
	FindGitHubAccessToken(ctx context.Context, userID string) (string, error)
}

func NewService(db *sqlx.DB, repo ProjectRepositoryStore) Service {
	return &service{db: db, repo: repo}
}

func NewServiceWithSync(db *sqlx.DB, repo ProjectRepositoryStore, syncService *SyncService, github GitHubRepositoryClient, account GitHubAccountStore) Service {
	catalog, _ := github.(GitHubRepositoryCatalog)
	return &service{db: db, repo: repo, sync: syncService, github: github, account: account, catalog: catalog}
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
			repository, err = service.repo.ConnectRepository(ctx, tx, project.ID, workspaceID, *input.Repository)
			project.Repository = &repository
		}
		return err
	})
	if err != nil {
		return Project{}, fmt.Errorf("create project: %w", err)
	}
	return project, nil
}

func (service *service) ListByWorkspace(ctx context.Context, workspaceID string, query pagination.Query) (pagination.Result[Project], error) {
	if strings.TrimSpace(workspaceID) == "" {
		return pagination.Result[Project]{}, ErrInvalidProjectInput
	}
	return service.repo.ListByWorkspace(ctx, workspaceID, query)
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
	status := input.Status
	if status == "" {
		status = ProjectStatusActive
	}
	if service.db == nil || projectID == "" || name == "" {
		return Project{}, ErrInvalidProjectInput
	}
	project, err := service.repo.Update(ctx, projectID, name, input.Description, string(status))
	if err != nil {
		return Project{}, ErrProjectNotFound
	}
	return project, nil
}

func (service *service) UpdateRepositoryBranch(ctx context.Context, projectID, branch string) (ProjectRepository, error) {
	if service.db == nil || strings.TrimSpace(projectID) == "" || strings.TrimSpace(branch) == "" {
		return ProjectRepository{}, ErrInvalidProjectInput
	}
	return service.repo.UpdateRepositoryBranch(ctx, projectID, strings.TrimSpace(branch))
}

func (service *service) ConnectRepository(ctx context.Context, projectID string, input ConnectRepositoryRequest) (ProjectRepository, error) {
	if service.db == nil || strings.TrimSpace(projectID) == "" || input.GitHubRepositoryID <= 0 || strings.TrimSpace(input.GitHubOwner) == "" || strings.TrimSpace(input.GitHubRepositoryName) == "" || strings.TrimSpace(input.RepositoryURL) == "" || strings.TrimSpace(input.DefaultBranch) == "" {
		return ProjectRepository{}, ErrInvalidProjectInput
	}
	project, err := service.repo.FindByID(ctx, projectID)
	if err != nil {
		return ProjectRepository{}, ErrProjectNotFound
	}
	var repository ProjectRepository
	err = appdatabase.WithTx(ctx, service.db, func(ctx context.Context, tx *sqlx.Tx) error {
		var err error
		repository, err = service.repo.ConnectRepository(ctx, tx, projectID, project.WorkspaceID, input)
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

func (service *service) SyncWorkspaceProjects(ctx context.Context, workspaceID string) (SyncWorkspaceProjectsResponse, error) {
	if service.sync == nil || strings.TrimSpace(workspaceID) == "" {
		return SyncWorkspaceProjectsResponse{}, ErrInvalidProjectInput
	}
	result := SyncWorkspaceProjectsResponse{}
	query := pagination.Query{Page: 1, PerPage: pagination.MaxPerPage}
	for {
		projects, err := service.repo.ListByWorkspace(ctx, workspaceID, query)
		if err != nil {
			return SyncWorkspaceProjectsResponse{}, err
		}
		for _, project := range projects.Items {
			if project.Repository == nil {
				continue
			}
			if err := service.sync.SyncProject(ctx, project.ID); err != nil {
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

func (service *service) ResolveRepository(ctx context.Context, repositoryURL string) (ResolvedRepository, error) {
	parsed, err := url.Parse(strings.TrimSpace(repositoryURL))
	if err != nil || !strings.EqualFold(parsed.Scheme, "https") || !strings.EqualFold(parsed.Host, "github.com") || service.github == nil {
		return ResolvedRepository{}, ErrInvalidProjectInput
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return ResolvedRepository{}, ErrInvalidProjectInput
	}
	name := strings.TrimSuffix(parts[1], ".git")
	repository, err := service.github.InspectRepository(ctx, parts[0], name)
	if err != nil {
		return ResolvedRepository{}, fmt.Errorf("resolve GitHub repository: %w", err)
	}
	return ResolvedRepository{Repository: ConnectRepositoryRequest{
		GitHubRepositoryID:   repository.ID,
		GitHubOwner:          repository.Owner,
		GitHubRepositoryName: repository.Name,
		RepositoryURL:        repository.URL,
		Private:              repository.Private,
		DefaultBranch:        repository.DefaultBranch,
	}, Branches: repository.Branches}, nil
}

func (service *service) ListGitHubRepositories(ctx context.Context, workspaceID, userID string) (GitHubRepositoryCatalogResponse, error) {
	owner, err := service.isWorkspaceOwner(ctx, workspaceID, userID)
	if err != nil {
		return GitHubRepositoryCatalogResponse{}, fmt.Errorf("check workspace owner: %w", err)
	}
	if !owner {
		return GitHubRepositoryCatalogResponse{}, ErrWorkspaceOwnerRequired
	}
	if service.account == nil {
		return GitHubRepositoryCatalogResponse{}, ErrGitHubAccountUnavailable
	}
	if service.catalog == nil {
		return GitHubRepositoryCatalogResponse{}, ErrGitHubCatalogUnavailable
	}
	token, err := service.account.FindGitHubAccessToken(ctx, userID)
	if err != nil {
		return GitHubRepositoryCatalogResponse{}, fmt.Errorf("%w: %v", ErrGitHubAccountUnavailable, err)
	}
	organizations, err := service.catalog.ListOrganizations(ctx, token)
	if err != nil {
		return GitHubRepositoryCatalogResponse{}, err
	}
	result := GitHubRepositoryCatalogResponse{Organizations: organizations, Repositories: []GitHubRepositoryOption{}}
	seen := make(map[int64]bool)
	personalRepositories, err := service.catalog.ListRepositories(ctx, token, "")
	if err != nil {
		return GitHubRepositoryCatalogResponse{}, err
	}
	for _, repository := range personalRepositories {
		if seen[repository.ID] {
			continue
		}
		seen[repository.ID] = true
		result.Repositories = append(result.Repositories, repositoryOption(repository, "personal"))
	}
	for _, organization := range organizations {
		repositories, listErr := service.catalog.ListRepositories(ctx, token, organization)
		if listErr != nil {
			return GitHubRepositoryCatalogResponse{}, listErr
		}
		for _, repository := range repositories {
			if seen[repository.ID] {
				continue
			}
			seen[repository.ID] = true
			result.Repositories = append(result.Repositories, repositoryOption(repository, organization))
		}
	}
	return result, nil
}

func (service *service) ImportGitHubRepositories(ctx context.Context, workspaceID, userID string, input ImportGitHubRepositoriesRequest) (ImportGitHubRepositoriesResponse, error) {
	owner, err := service.isWorkspaceOwner(ctx, workspaceID, userID)
	if err != nil {
		return ImportGitHubRepositoriesResponse{}, fmt.Errorf("check workspace owner: %w", err)
	}
	if service.db == nil || !owner || len(input.Repositories) == 0 {
		if !owner {
			return ImportGitHubRepositoriesResponse{}, ErrWorkspaceOwnerRequired
		}
		return ImportGitHubRepositoriesResponse{}, ErrInvalidProjectInput
	}
	result := ImportGitHubRepositoriesResponse{Projects: make([]Project, 0, len(input.Repositories))}
	err = appdatabase.WithTx(ctx, service.db, func(ctx context.Context, tx *sqlx.Tx) error {
		seen := make(map[int64]struct{}, len(input.Repositories))
		for _, repository := range input.Repositories {
			name := strings.TrimSpace(repository.GitHubRepositoryName)
			if name == "" || repository.GitHubRepositoryID <= 0 {
				return ErrInvalidProjectInput
			}
			if _, duplicate := seen[repository.GitHubRepositoryID]; duplicate {
				result.Skipped++
				continue
			}
			seen[repository.GitHubRepositoryID] = struct{}{}
			exists, err := service.repo.RepositoryExists(ctx, tx, workspaceID, repository.GitHubRepositoryID)
			if err != nil {
				return err
			}
			if exists {
				result.Skipped++
				continue
			}
			project, err := service.repo.Create(ctx, tx, workspaceID, name, slugify(repository.GitHubOwner+"-"+name), nil, string(ProjectTypeRepository), userID)
			if err != nil {
				return err
			}
			connected, err := service.repo.ConnectRepository(ctx, tx, project.ID, workspaceID, repository)
			if err != nil {
				return err
			}
			project.Repository = &connected
			result.Projects = append(result.Projects, project)
		}
		return nil
	})
	if err != nil {
		return ImportGitHubRepositoriesResponse{}, fmt.Errorf("import GitHub repositories: %w", err)
	}
	return result, nil
}

func (service *service) isWorkspaceOwner(ctx context.Context, workspaceID, userID string) (bool, error) {
	if strings.TrimSpace(workspaceID) == "" || strings.TrimSpace(userID) == "" {
		return false, nil
	}
	owner, err := service.repo.IsWorkspaceOwner(ctx, workspaceID, userID)
	return owner, err
}

func repositoryOption(repository GitHubRepository, organization string) GitHubRepositoryOption {
	return GitHubRepositoryOption{ConnectRepositoryRequest: ConnectRepositoryRequest{
		GitHubRepositoryID: repository.ID, GitHubOwner: repository.Owner, GitHubRepositoryName: repository.Name,
		RepositoryURL: repository.URL, Private: repository.Private, DefaultBranch: repository.DefaultBranch,
	}, Organization: organization}
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(value, "-")
	return strings.Trim(strings.TrimSpace(value), "-")
}
