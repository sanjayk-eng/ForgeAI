package core

import (
	"context"
	"database/sql"
	"fmt"

	"ai-agent/internal/shared/pagination"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	Create(ctx context.Context, tx *sqlx.Tx, workspaceID, name, slug string, description *string, projectType, createdBy string) (Project, error)
	ListByWorkspace(ctx context.Context, workspaceID string, query pagination.Query) (pagination.Result[Project], error)
	FindByID(ctx context.Context, projectID string) (Project, error)
	FindByWorkspaceSlug(ctx context.Context, workspaceID, slug string) (Project, error)
	Update(ctx context.Context, projectID, name string, description *string, status string) (Project, error)
	Delete(ctx context.Context, projectID string) error
	IsWorkspaceOwner(ctx context.Context, workspaceID, userID string) (bool, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

const projectSelect = `
	SELECT p.id, p.workspace_id, p.name, p.slug, p.description,
	       type_enum.code AS type, status_enum.code AS status,
	       p.created_by, p.created_at, p.updated_at
	FROM tbl_project p
	JOIN tbl_enum type_enum ON type_enum.id = p.type_id
	JOIN tbl_enum status_enum ON status_enum.id = p.status_id`

func (repo *repository) Create(ctx context.Context, tx *sqlx.Tx, workspaceID, name, slug string, description *string, projectType, createdBy string) (Project, error) {
	var project Project
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
		err = tx.GetContext(ctx, &project, query, workspaceID, name, slug, description, createdBy, projectType)
	} else {
		err = repo.db.GetContext(ctx, &project, query, workspaceID, name, slug, description, createdBy, projectType)
	}
	if err != nil {
		return Project{}, fmt.Errorf("create project: %w", err)
	}
	return project, nil
}

func (repo *repository) ListByWorkspace(ctx context.Context, workspaceID string, query pagination.Query) (pagination.Result[Project], error) {
	var total int
	countQuery := `SELECT COUNT(*) FROM tbl_project p WHERE p.workspace_id = $1`
	countArgs := []any{workspaceID}

	if query.Search != "" {
		countQuery += ` AND p.name ILIKE '%' || $2 || '%'`
		countArgs = append(countArgs, query.Search)
	}

	if err := repo.db.GetContext(ctx, &total, countQuery, countArgs...); err != nil {
		return pagination.Result[Project]{}, fmt.Errorf("count projects by workspace: %w", err)
	}

	listQuery := projectSelect + ` WHERE p.workspace_id = $1`
	listArgs := []any{workspaceID}

	if query.Search != "" {
		listQuery += ` AND p.name ILIKE '%' || $2 || '%'`
		listArgs = append(listArgs, query.Search)
	}

	listQuery += fmt.Sprintf(" ORDER BY p.created_at DESC LIMIT $%d OFFSET $%d", len(listArgs)+1, len(listArgs)+2)
	listArgs = append(listArgs, query.PerPage, query.Offset())

	var projects []Project
	if err := repo.db.SelectContext(ctx, &projects, listQuery, listArgs...); err != nil {
		return pagination.Result[Project]{}, fmt.Errorf("list projects by workspace: %w", err)
	}

	return pagination.NewResult(projects, query, total), nil
}

func (repo *repository) FindByID(ctx context.Context, projectID string) (Project, error) {
	var project Project
	if err := repo.db.GetContext(ctx, &project, projectSelect+` WHERE p.id = $1`, projectID); err != nil {
		return Project{}, fmt.Errorf("find project: %w", err)
	}
	return project, nil
}

func (repo *repository) FindByWorkspaceSlug(ctx context.Context, workspaceID, slug string) (Project, error) {
	var project Project
	if err := repo.db.GetContext(ctx, &project, projectSelect+` WHERE p.workspace_id = $1 AND p.slug = $2`, workspaceID, slug); err != nil {
		return Project{}, fmt.Errorf("find project: %w", err)
	}
	return project, nil
}

func (repo *repository) Update(ctx context.Context, projectID, name string, description *string, status string) (Project, error) {
	var project Project
	query := `
		WITH updated AS (
			UPDATE tbl_project
			SET name = $2::varchar(255),
			    slug = regexp_replace(trim(lower($2::varchar(255))), '[^a-z0-9]+', '-', 'g'),
			    description = $3,
			    status_id = (SELECT id FROM tbl_enum WHERE category = 'PROJECT_STATUS' AND code = $4),
			    updated_at = NOW()
			WHERE id = $1
			RETURNING id, workspace_id, name, slug, description, type_id, status_id,
			          created_by, created_at, updated_at
		)` + `
		SELECT updated.id, updated.workspace_id, updated.name, updated.slug,
		       updated.description, type_enum.code AS type, status_enum.code AS status,
		       updated.created_by, updated.created_at, updated.updated_at
		FROM updated
		JOIN tbl_enum type_enum ON type_enum.id = updated.type_id
		JOIN tbl_enum status_enum ON status_enum.id = updated.status_id`

	if err := repo.db.GetContext(ctx, &project, query, projectID, name, description, status); err != nil {
		return Project{}, fmt.Errorf("update project: %w", err)
	}
	return project, nil
}

func (repo *repository) Delete(ctx context.Context, projectID string) error {
	result, err := repo.db.ExecContext(ctx, `DELETE FROM tbl_project WHERE id = $1`, projectID)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check deleted project: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("delete project: %w", sql.ErrNoRows)
	}
	return nil
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
