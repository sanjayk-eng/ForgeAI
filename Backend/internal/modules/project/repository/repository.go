package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	ConnectRepository(ctx context.Context, tx *sqlx.Tx, projectID, workspaceID string, input ConnectRepositoryRequest) (ProjectRepository, error)
	FindByProjectID(ctx context.Context, projectID string) (ProjectRepository, error)
	UpdateBranch(ctx context.Context, projectID, branch string) (ProjectRepository, error)
	RepositoryExists(ctx context.Context, tx *sqlx.Tx, workspaceID string, githubRepositoryID int64) (bool, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (repo *repository) ConnectRepository(ctx context.Context, tx *sqlx.Tx, projectID, workspaceID string, input ConnectRepositoryRequest) (ProjectRepository, error) {
	var pr ProjectRepository
	query := `
		INSERT INTO tbl_project_repository (
			project_id, workspace_id, github_repository_id, github_owner, github_repository_name,
			repository_url, default_branch, sync_status_id
		)
		SELECT $1, $2, $3, $4, $5, $6, $7, e.id
		FROM tbl_enum e
		WHERE e.category = 'PROJECT_REPOSITORY_SYNC_STATUS'
		  AND e.code = 'PENDING'
		RETURNING id AS repository_id, project_id, github_repository_id, github_owner,
		          github_repository_name, repository_url, default_branch,
		          last_synced_at, created_at AS repository_created_at,
		          updated_at AS repository_updated_at`

	var err error
	if tx != nil {
		err = tx.GetContext(ctx, &pr, query, projectID, workspaceID, input.GitHubRepositoryID, input.GitHubOwner, input.GitHubRepositoryName, input.RepositoryURL, input.DefaultBranch)
	} else {
		err = repo.db.GetContext(ctx, &pr, query, projectID, workspaceID, input.GitHubRepositoryID, input.GitHubOwner, input.GitHubRepositoryName, input.RepositoryURL, input.DefaultBranch)
	}
	if err != nil {
		return ProjectRepository{}, fmt.Errorf("connect project repository: %w", err)
	}
	pr.SyncStatus = SyncStatusPending
	return pr, nil
}

func (repo *repository) FindByProjectID(ctx context.Context, projectID string) (ProjectRepository, error) {
	var pr ProjectRepository
	err := repo.db.GetContext(ctx, &pr, `
		SELECT pr.id AS repository_id, pr.project_id, pr.github_repository_id,
		       pr.github_owner, pr.github_repository_name, pr.repository_url,
		       pr.default_branch, sync_enum.code AS sync_status,
		       pr.last_synced_at, pr.created_at AS repository_created_at,
		       pr.updated_at AS repository_updated_at
		FROM tbl_project_repository pr
		LEFT JOIN tbl_enum sync_enum ON sync_enum.id = pr.sync_status_id
		WHERE pr.project_id = $1`, projectID)
	if err != nil {
		return ProjectRepository{}, fmt.Errorf("find repository by project id: %w", err)
	}
	return pr, nil
}

func (repo *repository) UpdateBranch(ctx context.Context, projectID, branch string) (ProjectRepository, error) {
	var pr ProjectRepository
	err := repo.db.GetContext(ctx, &pr, `
		WITH updated AS (
			UPDATE tbl_project_repository
			SET default_branch = $2::varchar(255),
				sync_status_id = (
					SELECT id FROM tbl_enum
					WHERE category = 'PROJECT_REPOSITORY_SYNC_STATUS' AND code = 'PENDING'
				),
				last_synced_at = NULL,
				updated_at = NOW()
			WHERE project_id = $1
			RETURNING id, project_id, github_repository_id, github_owner,
					  github_repository_name, repository_url, default_branch,
					  last_synced_at, created_at, updated_at
		)
		SELECT updated.id AS repository_id, updated.project_id, updated.github_repository_id,
		       updated.github_owner, updated.github_repository_name, updated.repository_url,
		       updated.default_branch, 'PENDING' AS sync_status, updated.last_synced_at,
		       updated.created_at AS repository_created_at, updated.updated_at AS repository_updated_at
		FROM updated`, projectID, branch)
	if err != nil {
		return ProjectRepository{}, fmt.Errorf("update repository branch: %w", err)
	}
	pr.SyncStatus = SyncStatusPending
	return pr, nil
}

func (repo *repository) RepositoryExists(ctx context.Context, tx *sqlx.Tx, workspaceID string, githubRepositoryID int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM tbl_project_repository WHERE workspace_id = $1 AND github_repository_id = $2)`

	var err error
	if tx != nil {
		err = tx.GetContext(ctx, &exists, query, workspaceID, githubRepositoryID)
	} else {
		err = repo.db.GetContext(ctx, &exists, query, workspaceID, githubRepositoryID)
	}
	return exists, err
}
