package project

import (
	"context"
	"fmt"
	"time"

	"ai-agent/internal/shared/pagination"

	"github.com/jmoiron/sqlx"
)

type ProjectRepositoryStore interface {
	SyncRepositoryStore
	Create(ctx context.Context, tx *sqlx.Tx, workspaceID, name, slug string, description *string, projectType, createdBy string) (Project, error)
	ListByWorkspace(ctx context.Context, workspaceID string, query pagination.Query) (pagination.Result[Project], error)
	FindByID(ctx context.Context, projectID string) (Project, error)
	FindByWorkspaceSlug(ctx context.Context, workspaceID, slug string) (Project, error)
	Update(ctx context.Context, projectID, name string, description *string, status string) (Project, error)
	ConnectRepository(ctx context.Context, tx *sqlx.Tx, projectID string, input ConnectRepositoryRequest) (ProjectRepository, error)
	IsWorkspaceOwner(ctx context.Context, workspaceID, userID string) (bool, error)
}

func (repo *repository) IsWorkspaceOwner(ctx context.Context, workspaceID, userID string) (bool, error) {
	var exists bool
	err := repo.db.GetContext(ctx, &exists, `
		SELECT EXISTS(
			SELECT 1
			FROM tbl_workspace w
			WHERE w.id = $1
			  AND (
				  w.owner_id = $2
				  OR EXISTS (
					  SELECT 1
					  FROM tbl_workspace_member wm
					  JOIN tbl_enum role_enum ON role_enum.id = wm.role_id
					  WHERE wm.workspace_id = w.id
						AND wm.user_id = $2
						AND role_enum.category = 'WORKSPACE_ROLE'
						AND role_enum.code = 'OWNER'
				  )
				)
		)`, workspaceID, userID)
	return exists, err
}

type repository struct {
	db *sqlx.DB
}

type projectRow struct {
	Project
	RepositoryID         *string     `db:"repository_id"`
	RepositoryProjectID  *string     `db:"repository_project_id"`
	GitHubRepositoryID   *int64      `db:"github_repository_id"`
	GitHubOwner          *string     `db:"github_owner"`
	GitHubRepositoryName *string     `db:"github_repository_name"`
	RepositoryURL        *string     `db:"repository_url"`
	DefaultBranch        *string     `db:"default_branch"`
	SyncStatus           *SyncStatus `db:"sync_status"`
	LastSyncedAt         *time.Time  `db:"last_synced_at"`
	RepositoryCreatedAt  *time.Time  `db:"repository_created_at"`
	RepositoryUpdatedAt  *time.Time  `db:"repository_updated_at"`
}

func NewRepository(db *sqlx.DB) ProjectRepositoryStore {
	return &repository{db: db}
}

const projectSelect = `
	SELECT p.id, p.workspace_id, p.name, p.slug, p.description,
	       type_enum.code AS type, status_enum.code AS status,
	       p.created_by, p.created_at, p.updated_at,
	       pr.id AS repository_id, pr.project_id AS repository_project_id,
	       pr.github_repository_id, pr.github_owner, pr.github_repository_name,
	       pr.repository_url, pr.default_branch,
	       sync_enum.code AS sync_status, pr.last_synced_at,
	       pr.created_at AS repository_created_at,
	       pr.updated_at AS repository_updated_at
	FROM tbl_project p
	JOIN tbl_enum type_enum ON type_enum.id = p.type_id
	JOIN tbl_enum status_enum ON status_enum.id = p.status_id
	LEFT JOIN tbl_project_repository pr ON pr.project_id = p.id
	LEFT JOIN tbl_enum sync_enum ON sync_enum.id = pr.sync_status_id`

func (repo *repository) Create(ctx context.Context, tx *sqlx.Tx, workspaceID, name, slug string, description *string, projectType, createdBy string) (Project, error) {
	var row projectRow
	query := `
		WITH inserted AS (
			INSERT INTO tbl_project (workspace_id, name, slug, description, type_id, status_id, created_by)
			SELECT $1, $2, $3, $4, type_enum.id, status_enum.id, $5
			FROM tbl_enum type_enum
			CROSS JOIN tbl_enum status_enum
			WHERE type_enum.category = 'PROJECT_TYPE'
			  AND type_enum.code = $6
			  AND status_enum.category = 'PROJECT_STATUS'
			  AND status_enum.code = 'ACTIVE'
			RETURNING id, workspace_id, name, slug, description, type_id, status_id,
			          created_by, created_at, updated_at
		)
		SELECT inserted.id, inserted.workspace_id, inserted.name, inserted.slug,
		       inserted.description, type_enum.code AS type, status_enum.code AS status,
		       inserted.created_by, inserted.created_at, inserted.updated_at
		FROM inserted
		JOIN tbl_enum type_enum ON type_enum.id = inserted.type_id
		JOIN tbl_enum status_enum ON status_enum.id = inserted.status_id`
	var err error
	if tx != nil {
		err = tx.GetContext(ctx, &row, query, workspaceID, name, slug, description, createdBy, projectType)
	} else {
		err = repo.db.GetContext(ctx, &row, query, workspaceID, name, slug, description, createdBy, projectType)
	}
	if err != nil {
		return Project{}, fmt.Errorf("create project: %w", err)
	}
	return row.project(), nil
}

func (repo *repository) ListByWorkspace(ctx context.Context, workspaceID string, query pagination.Query) (pagination.Result[Project], error) {
	var total int
	countQuery := `SELECT COUNT(*) FROM tbl_project p LEFT JOIN tbl_project_repository pr ON pr.project_id = p.id WHERE p.workspace_id = $1`
	countArgs := []any{workspaceID}
	if query.Search != "" {
		countQuery += ` AND (p.name ILIKE '%' || $2 || '%' OR p.slug ILIKE '%' || $2 || '%' OR pr.github_owner ILIKE '%' || $2 || '%' OR pr.github_repository_name ILIKE '%' || $2 || '%' OR pr.repository_url ILIKE '%' || $2 || '%')`
		countArgs = append(countArgs, query.Search)
	}
	if err := repo.db.GetContext(ctx, &total, countQuery, countArgs...); err != nil {
		return pagination.Result[Project]{}, fmt.Errorf("count projects by workspace: %w", err)
	}

	listQuery := projectSelect + ` WHERE p.workspace_id = $1`
	listArgs := []any{workspaceID}
	if query.Search != "" {
		listQuery += ` AND (p.name ILIKE '%' || $2 || '%' OR p.slug ILIKE '%' || $2 || '%' OR pr.github_owner ILIKE '%' || $2 || '%' OR pr.github_repository_name ILIKE '%' || $2 || '%' OR pr.repository_url ILIKE '%' || $2 || '%')`
		listArgs = append(listArgs, query.Search)
	}
	listQuery += fmt.Sprintf(" ORDER BY p.created_at DESC LIMIT $%d OFFSET $%d", len(listArgs)+1, len(listArgs)+2)
	listArgs = append(listArgs, query.PerPage, query.Offset())
	var rows []projectRow
	if err := repo.db.SelectContext(ctx, &rows, listQuery, listArgs...); err != nil {
		return pagination.Result[Project]{}, fmt.Errorf("list projects by workspace: %w", err)
	}
	projects := make([]Project, 0, len(rows))
	for _, row := range rows {
		projects = append(projects, row.project())
	}
	return pagination.NewResult(projects, query, total), nil
}

func (repo *repository) FindByID(ctx context.Context, projectID string) (Project, error) {
	return repo.find(ctx, projectSelect+` WHERE p.id = $1`, projectID)
}

func (repo *repository) FindByWorkspaceSlug(ctx context.Context, workspaceID, slug string) (Project, error) {
	return repo.find(ctx, projectSelect+` WHERE p.workspace_id = $1 AND p.slug = $2`, workspaceID, slug)
}

func (repo *repository) find(ctx context.Context, query string, args ...any) (Project, error) {
	var row projectRow
	if err := repo.db.GetContext(ctx, &row, query, args...); err != nil {
		return Project{}, fmt.Errorf("find project: %w", err)
	}
	return row.project(), nil
}

func (repo *repository) Update(ctx context.Context, projectID, name string, description *string, status string) (Project, error) {
	var row projectRow
	query := `
		WITH updated AS (
			UPDATE tbl_project
			SET name = $2,
			    slug = regexp_replace(trim(lower($2)), '[^a-z0-9]+', '-', 'g'),
			    description = $3,
			    status_id = (SELECT id FROM tbl_enum WHERE category = 'PROJECT_STATUS' AND code = $4),
			    updated_at = NOW()
			WHERE id = $1
			RETURNING id
		)` + projectSelect + `
		JOIN updated ON updated.id = p.id`
	if err := repo.db.GetContext(ctx, &row, query, projectID, name, description, status); err != nil {
		return Project{}, fmt.Errorf("update project: %w", err)
	}
	return row.project(), nil
}

func (repo *repository) ConnectRepository(ctx context.Context, tx *sqlx.Tx, projectID string, input ConnectRepositoryRequest) (ProjectRepository, error) {
	var repository ProjectRepository
	query := `
		INSERT INTO tbl_project_repository (
			project_id, github_repository_id, github_owner, github_repository_name,
			repository_url, default_branch, sync_status_id
		)
		SELECT $1, $2, $3, $4, $5, $6, e.id
		FROM tbl_enum e
		WHERE e.category = 'PROJECT_REPOSITORY_SYNC_STATUS'
		  AND e.code = 'PENDING'
		RETURNING id AS repository_id, project_id, github_repository_id, github_owner,
		          github_repository_name, repository_url, default_branch,
		          last_synced_at, created_at AS repository_created_at,
		          updated_at AS repository_updated_at`
	var err error
	if tx != nil {
		err = tx.GetContext(ctx, &repository, query, projectID, input.GitHubRepositoryID, input.GitHubOwner, input.GitHubRepositoryName, input.RepositoryURL, input.DefaultBranch)
	} else {
		err = repo.db.GetContext(ctx, &repository, query, projectID, input.GitHubRepositoryID, input.GitHubOwner, input.GitHubRepositoryName, input.RepositoryURL, input.DefaultBranch)
	}
	if err != nil {
		return ProjectRepository{}, fmt.Errorf("connect project repository: %w", err)
	}
	repository.SyncStatus = SyncStatusPending
	return repository, nil
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

func (repo *repository) MarkRepositorySynced(ctx context.Context, repositoryID string, repository GitHubRepository) error {
	result, err := repo.db.ExecContext(ctx, `
		UPDATE tbl_project_repository
		SET github_owner = $2,
		    github_repository_name = $3,
		    repository_url = $4,
		    default_branch = $5,
		    sync_status_id = (
			    SELECT id FROM tbl_enum
			    WHERE category = 'PROJECT_REPOSITORY_SYNC_STATUS' AND code = 'SYNCED'
		    ),
		    last_synced_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1`, repositoryID, repository.Owner, repository.Name, repository.URL, repository.DefaultBranch)
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

func (row projectRow) project() Project {
	project := row.Project
	if row.RepositoryID == nil {
		return project
	}
	project.Repository = &ProjectRepository{
		ID: row.RepositoryIDValue(), ProjectID: *row.RepositoryProjectID,
		GitHubRepositoryID: *row.GitHubRepositoryID, GitHubOwner: *row.GitHubOwner,
		GitHubRepositoryName: *row.GitHubRepositoryName, RepositoryURL: *row.RepositoryURL,
		DefaultBranch: *row.DefaultBranch, SyncStatus: *row.SyncStatus,
		LastSyncedAt: row.LastSyncedAt, CreatedAt: *row.RepositoryCreatedAt, UpdatedAt: *row.RepositoryUpdatedAt,
	}
	return project
}

func (row projectRow) RepositoryIDValue() string { return *row.RepositoryID }
