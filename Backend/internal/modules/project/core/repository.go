package core

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"ai-agent/internal/shared/pagination"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	Create(ctx context.Context, organizationID, name, slug string, description *string, createdBy string) (Project, error)
	ListByOrganization(ctx context.Context, organizationID string, query pagination.Query) (pagination.Result[Project], error)
	FindByID(ctx context.Context, projectID string) (Project, error)
	FindByOrganizationSlug(ctx context.Context, organizationID, slug string) (Project, error)
	Update(ctx context.Context, projectID, name string, description *string, status string) (Project, error)
	Delete(ctx context.Context, projectID string) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

const projectSelect = `
	SELECT id, organization_id, name, slug, description, status,
	       created_by, created_at, updated_at, deleted_at
	FROM tbl_project`

func (repo *repository) Create(ctx context.Context, organizationID, name, slug string, description *string, createdBy string) (Project, error) {
	var project Project
	query := `
		INSERT INTO tbl_project (organization_id, name, slug, description, created_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, organization_id, name, slug, description, status,
		          created_by, created_at, updated_at, deleted_at
	`
	err := repo.db.GetContext(ctx, &project, query, organizationID, name, slug, description, createdBy)
	if err != nil {
		return Project{}, fmt.Errorf("create project: %w", err)
	}
	return project, nil
}

func (repo *repository) ListByOrganization(ctx context.Context, organizationID string, query pagination.Query) (pagination.Result[Project], error) {
	var total int
	countQuery := `SELECT COUNT(*) FROM tbl_project WHERE organization_id = $1 AND deleted_at IS NULL`
	countArgs := []any{organizationID}

	if query.Search != "" {
		countQuery += ` AND name ILIKE '%' || $2 || '%'`
		countArgs = append(countArgs, query.Search)
	}

	if err := repo.db.GetContext(ctx, &total, countQuery, countArgs...); err != nil {
		return pagination.Result[Project]{}, fmt.Errorf("count projects: %w", err)
	}

	listQuery := projectSelect + ` WHERE organization_id = $1 AND deleted_at IS NULL`
	listArgs := []any{organizationID}

	if query.Search != "" {
		listQuery += ` AND name ILIKE '%' || $2 || '%'`
		listArgs = append(listArgs, query.Search)
	}

	listQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", len(listArgs)+1, len(listArgs)+2)
	listArgs = append(listArgs, query.PerPage, query.Offset())

	var projects []Project
	if err := repo.db.SelectContext(ctx, &projects, listQuery, listArgs...); err != nil {
		return pagination.Result[Project]{}, fmt.Errorf("list projects: %w", err)
	}

	return pagination.NewResult(projects, query, total), nil
}

func (repo *repository) FindByID(ctx context.Context, projectID string) (Project, error) {
	var project Project
	if err := repo.db.GetContext(ctx, &project, projectSelect+` WHERE id = $1 AND deleted_at IS NULL`, projectID); err != nil {
		return Project{}, fmt.Errorf("find project: %w", err)
	}
	return project, nil
}

func (repo *repository) FindByOrganizationSlug(ctx context.Context, organizationID, slug string) (Project, error) {
	var project Project
	if err := repo.db.GetContext(ctx, &project, projectSelect+` WHERE organization_id = $1 AND slug = $2 AND deleted_at IS NULL`, organizationID, slug); err != nil {
		return Project{}, fmt.Errorf("find project: %w", err)
	}
	return project, nil
}

func (repo *repository) Update(ctx context.Context, projectID, name string, description *string, status string) (Project, error) {
	var project Project
	query := `
		UPDATE tbl_project
		SET name = $2,
		    slug = $3,
		    description = $4,
		    status = $5,
		    updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, organization_id, name, slug, description, status,
		          created_by, created_at, updated_at, deleted_at
	`
	slug := slugifyProjectName(name)
	if err := repo.db.GetContext(ctx, &project, query, projectID, name, slug, description, status); err != nil {
		return Project{}, fmt.Errorf("update project: %w", err)
	}
	return project, nil
}

func (repo *repository) Delete(ctx context.Context, projectID string) error {
	result, err := repo.db.ExecContext(ctx, `UPDATE tbl_project SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, projectID)
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

func slugifyProjectName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(value, "-")
	return strings.Trim(strings.TrimSpace(value), "-")
}
