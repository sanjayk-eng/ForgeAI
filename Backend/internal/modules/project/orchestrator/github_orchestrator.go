package orchestrator

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"ai-agent/internal/middleware"
	"ai-agent/internal/modules/project/core"
	"ai-agent/internal/modules/project/github"
	projectrepo "ai-agent/internal/modules/project/repository"
	appdatabase "ai-agent/pkg/database"

	"github.com/jmoiron/sqlx"
)

// GitHubOrchestrator handles GitHub-related operations
type GitHubOrchestrator struct {
	db           *sqlx.DB
	coreRepo     core.Repository
	repoRepo     projectrepo.Repository
	githubClient github.Client
	catalogSvc   github.CatalogService
	accountStore GitHubAccountStore
}

func NewGitHubOrchestrator(
	db *sqlx.DB,
	coreRepo core.Repository,
	repoRepo projectrepo.Repository,
	githubClient github.Client,
	catalogSvc github.CatalogService,
	accountStore GitHubAccountStore,
) *GitHubOrchestrator {
	return &GitHubOrchestrator{
		db:           db,
		coreRepo:     coreRepo,
		repoRepo:     repoRepo,
		githubClient: githubClient,
		catalogSvc:   catalogSvc,
		accountStore: accountStore,
	}
}

func (o *GitHubOrchestrator) ResolveRepository(ctx context.Context, repositoryURL string) (projectrepo.ResolvedRepository, error) {
	if o.githubClient == nil {
		return projectrepo.ResolvedRepository{}, github.ErrGitHubCatalogUnavailable
	}

	accessToken := o.getAccessToken(ctx)
	return o.githubClient.ResolveRepository(ctx, repositoryURL, accessToken)
}

func (o *GitHubOrchestrator) ListGitHubRepositories(ctx context.Context, workspaceID, userID string) (github.CatalogResponse, error) {
	if err := o.validateWorkspaceOwner(ctx, workspaceID, userID); err != nil {
		return github.CatalogResponse{}, err
	}

	if o.accountStore == nil {
		return github.CatalogResponse{}, github.ErrGitHubAccountUnavailable
	}

	if o.catalogSvc == nil {
		return github.CatalogResponse{}, github.ErrGitHubCatalogUnavailable
	}

	token, err := o.accountStore.FindGitHubAccessToken(ctx, userID)
	if err != nil {
		return github.CatalogResponse{}, fmt.Errorf("%w: %v", github.ErrGitHubAccountUnavailable, err)
	}

	return o.catalogSvc.ListRepositories(ctx, token)
}

func (o *GitHubOrchestrator) ImportGitHubRepositories(ctx context.Context, workspaceID, userID string, input ImportGitHubRepositoriesInput) (ImportGitHubRepositoriesResponse, error) {
	if err := o.validateWorkspaceOwner(ctx, workspaceID, userID); err != nil {
		return ImportGitHubRepositoriesResponse{}, err
	}

	if len(input.Repositories) == 0 {
		return ImportGitHubRepositoriesResponse{}, core.ErrInvalidProjectInput
	}

	result := ImportGitHubRepositoriesResponse{
		Projects: make([]ProjectWithRepository, 0, len(input.Repositories)),
	}

	err := appdatabase.WithTx(ctx, o.db, func(ctx context.Context, tx *sqlx.Tx) error {
		seen := make(map[int64]struct{}, len(input.Repositories))

		for _, repo := range input.Repositories {
			name := strings.TrimSpace(repo.GitHubRepositoryName)
			if name == "" || repo.GitHubRepositoryID <= 0 {
				return core.ErrInvalidProjectInput
			}

			if _, duplicate := seen[repo.GitHubRepositoryID]; duplicate {
				result.Skipped++
				continue
			}
			seen[repo.GitHubRepositoryID] = struct{}{}

			exists, err := o.repoRepo.RepositoryExists(ctx, tx, workspaceID, repo.GitHubRepositoryID)
			if err != nil {
				return err
			}
			if exists {
				result.Skipped++
				continue
			}

			project, err := o.coreRepo.Create(ctx, tx, workspaceID, name, slugify(repo.GitHubOwner+"-"+name), nil, string(core.ProjectTypeRepository), userID)
			if err != nil {
				return err
			}

			connected, err := o.repoRepo.ConnectRepository(ctx, tx, project.ID, workspaceID, repo)
			if err != nil {
				return err
			}

			result.Projects = append(result.Projects, ProjectWithRepository{
				Project:    project,
				Repository: &connected,
			})
		}

		return nil
	})

	if err != nil {
		return ImportGitHubRepositoriesResponse{}, fmt.Errorf("import GitHub repositories: %w", err)
	}

	return result, nil
}

func (o *GitHubOrchestrator) validateWorkspaceOwner(ctx context.Context, workspaceID, userID string) error {
	owner, err := o.coreRepo.IsWorkspaceOwner(ctx, workspaceID, userID)
	if err != nil {
		return fmt.Errorf("check workspace owner: %w", err)
	}
	if !owner {
		return github.ErrWorkspaceOwnerRequired
	}
	return nil
}

func (o *GitHubOrchestrator) getAccessToken(ctx context.Context) string {
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

type ImportGitHubRepositoriesInput struct {
	Repositories []projectrepo.ConnectRepositoryRequest
}

type ImportGitHubRepositoriesResponse struct {
	Projects []ProjectWithRepository
	Skipped  int
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(value, "-")
	return strings.Trim(strings.TrimSpace(value), "-")
}
