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
	AddMember(ctx context.Context, tx *sqlx.Tx, workspaceID, userID, role string) error
	ListMembers(ctx context.Context, workspaceID string) ([]WorkspaceMember, error)
	UpdateMemberRole(ctx context.Context, tx *sqlx.Tx, workspaceID, userID, role string) error
	RemoveMember(ctx context.Context, tx *sqlx.Tx, workspaceID, userID string) error
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

func (repo *repository) ListMembers(ctx context.Context, workspaceID string) ([]WorkspaceMember, error) {
	var members []WorkspaceMember
	if err := repo.db.SelectContext(ctx, &members, `
		SELECT wm.id, wm.workspace_id, wm.user_id, e.code AS role, wm.created_at, wm.updated_at
		FROM tbl_workspace_member wm
		JOIN tbl_enum e ON e.id = wm.role_id
		WHERE wm.workspace_id = $1
		ORDER BY wm.created_at ASC`, workspaceID); err != nil {
		return nil, fmt.Errorf("list workspace members: %w", err)
	}
	return members, nil
}

func (repo *repository) UpdateMemberRole(ctx context.Context, tx *sqlx.Tx, workspaceID, userID, role string) error {
	query := `
		UPDATE tbl_workspace_member wm
		SET role_id = (
			SELECT e.id
			FROM tbl_enum e
			WHERE e.category = 'WORKSPACE_ROLE' AND e.code = $3
		), updated_at = NOW()
		WHERE wm.workspace_id = $1 AND wm.user_id = $2`

	if _, err := tx.ExecContext(ctx, query, workspaceID, userID, strings.ToUpper(role)); err != nil {
		return fmt.Errorf("update workspace member role: %w", err)
	}
	return nil
}

func (repo *repository) RemoveMember(ctx context.Context, tx *sqlx.Tx, workspaceID, userID string) error {
	query := `DELETE FROM tbl_workspace_member WHERE workspace_id = $1 AND user_id = $2`
	if _, err := tx.ExecContext(ctx, query, workspaceID, userID); err != nil {
		return fmt.Errorf("remove workspace member: %w", err)
	}
	return nil
}
