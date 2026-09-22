package invite

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	Create(ctx context.Context, tx *sqlx.Tx, invite Invite) (Invite, error)
	ListByWorkspace(ctx context.Context, workspaceID string) ([]Invite, error)
	UpdateStatus(ctx context.Context, tx *sqlx.Tx, workspaceID, inviteID, status string) (Invite, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (repo *repository) Create(ctx context.Context, tx *sqlx.Tx, invite Invite) (Invite, error) {
	query := `
		INSERT INTO tbl_workspace_invite (workspace_id, email, role_id, status_id, invited_by, token_hash, expires_at)
		SELECT $1, $2, role_enum.id, status_enum.id, $3, $4, $5
		FROM tbl_enum role_enum
		CROSS JOIN tbl_enum status_enum
		WHERE role_enum.category = 'WORKSPACE_ROLE'
		  AND role_enum.code = $6
		  AND status_enum.category = 'WORKSPACE_INVITE_STATUS'
		  AND status_enum.code = $7
		RETURNING id, workspace_id, email, invited_by, token_hash, expires_at, created_at, updated_at`

	var created Invite
	args := []any{invite.WorkspaceID, invite.Email, invite.InvitedBy, invite.TokenHash, invite.ExpiresAt, strings.ToUpper(invite.Role), string(invite.Status)}
	var err error
	if tx != nil {
		err = tx.GetContext(ctx, &created, query, args...)
	} else {
		err = repo.db.GetContext(ctx, &created, query, args...)
	}
	if err != nil {
		return Invite{}, fmt.Errorf("create workspace invite: %w", err)
	}
	created.Role = strings.ToUpper(invite.Role)
	created.Status = invite.Status
	return created, nil
}

func (repo *repository) ListByWorkspace(ctx context.Context, workspaceID string) ([]Invite, error) {
	type inviteRow struct {
		Invite
		InviterEmail string `db:"inviter_email"`
		InviterName  string `db:"inviter_name"`
	}
	var rows []inviteRow
	if err := repo.db.SelectContext(ctx, &rows, `
		SELECT wi.id, wi.workspace_id, wi.email, role_enum.code AS role,
		       status_enum.code AS status, wi.invited_by, wi.token_hash,
		       wi.expires_at, wi.created_at, wi.updated_at,
		       inviter.email AS inviter_email, inviter.name AS inviter_name
		FROM tbl_workspace_invite wi
		JOIN tbl_enum role_enum ON role_enum.id = wi.role_id
		JOIN tbl_enum status_enum ON status_enum.id = wi.status_id
		JOIN tbl_user inviter ON inviter.id = wi.invited_by
		WHERE wi.workspace_id = $1
		ORDER BY wi.created_at DESC`, workspaceID); err != nil {
		return nil, fmt.Errorf("list workspace invites: %w", err)
	}
	invites := make([]Invite, 0, len(rows))
	for _, row := range rows {
		row.Invite.InvitedByUser = MemberUser{ID: row.InvitedBy, Email: row.InviterEmail, Name: row.InviterName}
		invites = append(invites, row.Invite)
	}
	if invites == nil {
		invites = []Invite{}
	}
	return invites, nil
}

func (repo *repository) UpdateStatus(ctx context.Context, tx *sqlx.Tx, workspaceID, inviteID, status string) (Invite, error) {
	query := `
		UPDATE tbl_workspace_invite wi
		SET status_id = status_enum.id,
		    accepted_at = CASE WHEN status_enum.code = 'ACCEPTED' THEN NOW() ELSE wi.accepted_at END,
		    rejected_at = CASE WHEN status_enum.code = 'REJECTED' THEN NOW() ELSE wi.rejected_at END,
		    updated_at = NOW()
		FROM tbl_enum status_enum
		WHERE wi.id = $1
		  AND wi.workspace_id = $2
		  AND status_enum.category = 'WORKSPACE_INVITE_STATUS'
		  AND status_enum.code = $3
		RETURNING wi.id, wi.workspace_id, wi.email, status_enum.code AS status,
		          (SELECT role_enum.code FROM tbl_enum role_enum WHERE role_enum.id = wi.role_id) AS role,
		          wi.invited_by, wi.token_hash, wi.expires_at, wi.created_at, wi.updated_at`

	var updated Invite
	if err := tx.GetContext(ctx, &updated, query, inviteID, workspaceID, strings.ToUpper(status)); err != nil {
		return Invite{}, fmt.Errorf("update workspace invite status: %w", err)
	}
	return updated, nil
}
