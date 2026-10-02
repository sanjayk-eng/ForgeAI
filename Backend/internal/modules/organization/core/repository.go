package core

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

var (
	ErrOrganizationNotFound = errors.New("organization not found")
	ErrInvalidInput         = errors.New("invalid organization input")
)

type Repository interface {
	Create(ctx context.Context, tx *sqlx.Tx, org Organization) (string, error)
	FindByID(ctx context.Context, id string) (Organization, error)
	FindBySlug(ctx context.Context, slug string) (Organization, error)
	ListByUser(ctx context.Context, userID string) ([]Organization, error)
	Update(ctx context.Context, id string, updates map[string]interface{}) error
	Delete(ctx context.Context, id string) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, tx *sqlx.Tx, org Organization) (string, error) {
	query := `
		INSERT INTO tbl_organization (name, slug, description, settings, status, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (slug) DO NOTHING
		RETURNING id
	`
	var id string
	err := tx.QueryRowContext(ctx, query,
		org.Name, org.Slug, org.Description, org.Settings, org.Status, org.CreatedBy,
	).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrOrganizationSlugTaken
	}
	if err != nil {
		return "", fmt.Errorf("create organization: %w", err)
	}
	return id, nil
}

func (r *repository) FindByID(ctx context.Context, id string) (Organization, error) {
	var org Organization
	query := `SELECT * FROM tbl_organization WHERE id = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &org, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Organization{}, ErrOrganizationNotFound
		}
		return Organization{}, fmt.Errorf("find organization by id: %w", err)
	}
	return org, nil
}

func (r *repository) FindBySlug(ctx context.Context, slug string) (Organization, error) {
	var org Organization
	query := `SELECT * FROM tbl_organization WHERE slug = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &org, query, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Organization{}, ErrOrganizationNotFound
		}
		return Organization{}, fmt.Errorf("find organization by slug: %w", err)
	}
	return org, nil
}

func (r *repository) ListByUser(ctx context.Context, userID string) ([]Organization, error) {
	query := `
		SELECT o.* FROM tbl_organization o
		INNER JOIN tbl_organization_member m ON o.id = m.organization_id
		WHERE m.user_id = $1 AND m.deleted_at IS NULL AND o.deleted_at IS NULL
		ORDER BY o.created_at DESC
	`
	var orgs []Organization
	err := r.db.SelectContext(ctx, &orgs, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list organizations by user: %w", err)
	}
	return orgs, nil
}

func (r *repository) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	sets := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argIndex := 1

	for key, value := range updates {
		sets = append(sets, fmt.Sprintf("%s = $%d", key, argIndex))
		args = append(args, value)
		argIndex++
	}
	args = append(args, id)

	query := fmt.Sprintf("UPDATE tbl_organization SET %s WHERE id = $%d", strings.Join(sets, ", "), argIndex)
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update organization: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check update result: %w", err)
	}
	if rows == 0 {
		return ErrOrganizationNotFound
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	query := `UPDATE tbl_organization SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete organization: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check delete result: %w", err)
	}
	if rows == 0 {
		return ErrOrganizationNotFound
	}
	return nil
}
