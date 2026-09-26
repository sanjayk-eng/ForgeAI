package sync

import (
	"context"
	"fmt"

	"ai-agent/internal/modules/project/github"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	ClaimRepositoryForSync(ctx context.Context, projectID string) (SyncTarget, error)
	ClaimNextRepositoryForSync(ctx context.Context) (SyncTarget, error)
	MarkRepositorySynced(ctx context.Context, repositoryID string, repository github.Repository) error
	MarkRepositorySyncFailed(ctx context.Context, repositoryID string) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (repo *repository) ClaimRepositoryForSync(ctx context.Context, projectID string) (SyncTarget, error) {
	return repo.claimRepository(ctx, projectID)
}

func (repo *repository) ClaimNextRepositoryForSync(ctx context.Context) (SyncTarget, error) {
	return repo.claimRepository(ctx, "")
}

func (repo *repository) claimRepository(ctx context.Context, projectID string) (SyncTarget, error) {
	var target SyncTarget
	filter := ""
	args := []any{}

	if projectID != "" {
		filter = " AND pr.project_id = $1"
		args = append(args, projectID)
	}

	query := `
		WITH candidate AS (
			SELECT pr.id
			FROM tbl_project_repository pr
			JOIN tbl_enum status_enum ON status_enum.id = pr.sync_status_id
			WHERE status_enum.category = 'PROJECT_REPOSITORY_SYNC_STATUS'
			  AND status_enum.code IN ('PENDING', 'FAILED')` + filter + `
			ORDER BY pr.created_at
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		), claimed AS (
			UPDATE tbl_project_repository pr
			SET sync_status_id = (
				SELECT id FROM tbl_enum
				WHERE category = 'PROJECT_REPOSITORY_SYNC_STATUS' AND code = 'SYNCING'
			), updated_at = NOW()
			FROM candidate
			WHERE pr.id = candidate.id
			RETURNING pr.id, pr.github_owner, pr.github_repository_name
		)
		SELECT id, github_owner, github_repository_name FROM claimed`

	if err := repo.db.GetContext(ctx, &target, query, args...); err != nil {
		return SyncTarget{}, err
	}
	return target, nil
}

func (repo *repository) MarkRepositorySynced(ctx context.Context, repositoryID string, repository github.Repository) error {
	result, err := repo.db.ExecContext(ctx, `
		UPDATE tbl_project_repository
		SET github_owner = $2,
		    github_repository_name = $3,
		    repository_url = $4,
		    sync_status_id = (
			    SELECT id FROM tbl_enum
			    WHERE category = 'PROJECT_REPOSITORY_SYNC_STATUS' AND code = 'SYNCED'
		    ),
		    last_synced_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1`, repositoryID, repository.Owner, repository.Name, repository.URL)
	if err != nil {
		return fmt.Errorf("mark repository synced: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check synced repository update: %w", err)
	}
	if count != 1 {
		return fmt.Errorf("repository %q was not found", repositoryID)
	}
	return nil
}

func (repo *repository) MarkRepositorySyncFailed(ctx context.Context, repositoryID string) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE tbl_project_repository
		SET sync_status_id = (
			SELECT id FROM tbl_enum
			WHERE category = 'PROJECT_REPOSITORY_SYNC_STATUS' AND code = 'FAILED'
		), updated_at = NOW()
		WHERE id = $1`, repositoryID)
	if err != nil {
		return fmt.Errorf("mark repository sync failed: %w", err)
	}
	return nil
}
