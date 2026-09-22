package member

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	AddMember(ctx context.Context, tx *sqlx.Tx, workspaceID, userID, role string) error
	ListMembers(ctx context.Context, workspaceID string) ([]Member, error)
	UpdateMemberRole(ctx context.Context, tx *sqlx.Tx, workspaceID, userID, role string) error
	RemoveMember(ctx context.Context, tx *sqlx.Tx, workspaceID, userID string) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
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

func (repo *repository) ListMembers(ctx context.Context, workspaceID string) ([]Member, error) {
	var members []Member
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
