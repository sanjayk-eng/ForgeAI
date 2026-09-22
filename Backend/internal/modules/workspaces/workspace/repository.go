package workspace

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type WorkspaceRepository interface {
	Create(ctx context.Context, tx *sqlx.Tx, name, slug, ownerID string) (Workspace, error)
	FindByID(ctx context.Context, workspaceID string) (Workspace, error)
	ListByUser(ctx context.Context, userID string) ([]Workspace, error)
	Update(ctx context.Context, workspaceID, name, slug string) (Workspace, error)
	AddMember(ctx context.Context, tx *sqlx.Tx, workspaceID, userID, role string) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) WorkspaceRepository {
	return &repository{db: db}
}

func (repo *repository) Create(ctx context.Context, tx *sqlx.Tx, name, slug, ownerID string) (Workspace, error) {
	var workspace Workspace
	query := `
		INSERT INTO tbl_workspace (name, slug, owner_id)
		VALUES ($1, $2, $3)
		RETURNING id, name, slug, owner_id, created_at, updated_at`

	var err error
	if tx != nil {
		err = tx.GetContext(ctx, &workspace, query, name, slug, ownerID)
	} else {
		err = repo.db.GetContext(ctx, &workspace, query, name, slug, ownerID)
	}
	if err != nil {
		return Workspace{}, fmt.Errorf("create workspace: %w", err)
	}
	return workspace, nil
}

func (repo *repository) FindByID(ctx context.Context, workspaceID string) (Workspace, error) {
	var workspace Workspace
	if err := repo.db.GetContext(ctx, &workspace, `
		SELECT id, name, slug, owner_id, created_at, updated_at
		FROM tbl_workspace
		WHERE id = $1`, workspaceID); err != nil {
		return Workspace{}, fmt.Errorf("find workspace by id: %w", err)
	}
	return workspace, nil
}

func (repo *repository) ListByUser(ctx context.Context, userID string) ([]Workspace, error) {
	var workspaces []Workspace
	if err := repo.db.SelectContext(ctx, &workspaces, `
		SELECT w.id, w.name, w.slug, w.owner_id, w.created_at, w.updated_at
		FROM tbl_workspace w
		JOIN tbl_workspace_member wm ON wm.workspace_id = w.id
		WHERE wm.user_id = $1
		ORDER BY w.created_at DESC`, userID); err != nil {
		return nil, fmt.Errorf("list workspaces by user: %w", err)
	}
	return workspaces, nil
}

func (repo *repository) Update(ctx context.Context, workspaceID, name, slug string) (Workspace, error) {
	var workspace Workspace
	if err := repo.db.GetContext(ctx, &workspace, `
		UPDATE tbl_workspace
		SET name = $2, slug = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, slug, owner_id, created_at, updated_at`, workspaceID, name, slug); err != nil {
		return Workspace{}, fmt.Errorf("update workspace: %w", err)
	}
	return workspace, nil
}

func (repo *repository) AddMember(ctx context.Context, tx *sqlx.Tx, workspaceID, userID, role string) error {
	query := `
		INSERT INTO tbl_workspace_member (workspace_id, user_id, role_id)
		SELECT $1, $2, e.id
		FROM tbl_enum e
		WHERE e.category = 'WORKSPACE_ROLE' AND e.code = $3
		ON CONFLICT (workspace_id, user_id)
		DO UPDATE SET role_id = EXCLUDED.role_id, updated_at = NOW()`

	if _, err := tx.ExecContext(ctx, query, workspaceID, userID, strings.ToUpper(role)); err != nil {
		return fmt.Errorf("add workspace member: %w", err)
	}
	return nil
}
