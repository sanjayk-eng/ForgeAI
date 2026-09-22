package invite

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	Create(ctx context.Context, tx *sqlx.Tx, invite Invite) (Invite, error)
	ListByWorkspace(ctx context.Context, workspaceID string) ([]Invite, error)
	FindByToken(ctx context.Context, token string) (Invite, error)
	FindByID(ctx context.Context, workspaceID, inviteID string) (Invite, error)
	FindByWorkspaceAndEmail(ctx context.Context, workspaceID, email string) ([]Invite, error)
	UpdateStatusWithTimestamp(ctx context.Context, tx *sqlx.Tx, workspaceID, inviteID, status string, timestamp *time.Time) (Invite, error)
	AddWorkspaceMember(ctx context.Context, tx *sqlx.Tx, workspaceID, userID, role string) error
	GetWorkspaceInfo(ctx context.Context, workspaceID string) (WorkspaceInfo, error)
	GetUserInfo(ctx context.Context, userID string) (UserInfo, error)
}

type WorkspaceInfo struct {
	ID   string `db:"id"`
	Name string `db:"name"`
}

type UserInfo struct {
	ID    string `db:"id"`
	Email string `db:"email"`
	Name  string `db:"name"`
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

func (repo *repository) FindByToken(ctx context.Context, token string) (Invite, error) {
	query := `
		SELECT wi.id, wi.workspace_id, wi.email, role_enum.code AS role,
		       status_enum.code AS status, wi.invited_by, wi.token_hash,
		       wi.expires_at, wi.created_at, wi.updated_at
		FROM tbl_workspace_invite wi
		JOIN tbl_enum role_enum ON role_enum.id = wi.role_id
		JOIN tbl_enum status_enum ON status_enum.id = wi.status_id
		WHERE wi.token_hash = $1`

	var invite Invite
	if err := repo.db.GetContext(ctx, &invite, query, token); err != nil {
		return Invite{}, fmt.Errorf("find invite by token: %w", err)
	}
	return invite, nil
}

func (repo *repository) FindByID(ctx context.Context, workspaceID, inviteID string) (Invite, error) {
	query := `
		SELECT wi.id, wi.workspace_id, wi.email, role_enum.code AS role,
		       status_enum.code AS status, wi.invited_by, wi.token_hash,
		       wi.expires_at, wi.created_at, wi.updated_at
		FROM tbl_workspace_invite wi
		JOIN tbl_enum role_enum ON role_enum.id = wi.role_id
		JOIN tbl_enum status_enum ON status_enum.id = wi.status_id
		WHERE wi.id = $1 AND wi.workspace_id = $2`

	var invite Invite
	if err := repo.db.GetContext(ctx, &invite, query, inviteID, workspaceID); err != nil {
		return Invite{}, fmt.Errorf("find invite by ID: %w", err)
	}
	return invite, nil
}

func (repo *repository) FindByWorkspaceAndEmail(ctx context.Context, workspaceID, email string) ([]Invite, error) {
	query := `
		SELECT wi.id, wi.workspace_id, wi.email, role_enum.code AS role,
		       status_enum.code AS status, wi.invited_by, wi.token_hash,
		       wi.expires_at, wi.created_at, wi.updated_at
		FROM tbl_workspace_invite wi
		JOIN tbl_enum role_enum ON role_enum.id = wi.role_id
		JOIN tbl_enum status_enum ON status_enum.id = wi.status_id
		WHERE wi.workspace_id = $1 AND wi.email = $2
		ORDER BY wi.created_at DESC`

	var invites []Invite
	if err := repo.db.SelectContext(ctx, &invites, query, workspaceID, email); err != nil {
		return nil, fmt.Errorf("find invites by workspace and email: %w", err)
	}
	if invites == nil {
		invites = []Invite{}
	}
	return invites, nil
}

func (repo *repository) UpdateStatusWithTimestamp(ctx context.Context, tx *sqlx.Tx, workspaceID, inviteID, status string, timestamp *time.Time) (Invite, error) {
	query := `
		UPDATE tbl_workspace_invite wi
		SET status_id = status_enum.id,
		    accepted_at = CASE WHEN status_enum.code = 'ACCEPTED' THEN $4 ELSE wi.accepted_at END,
		    rejected_at = CASE WHEN status_enum.code = 'REJECTED' THEN $4 ELSE wi.rejected_at END,
		    revoked_at = CASE WHEN status_enum.code = 'REVOKED' THEN $4 ELSE wi.revoked_at END,
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
	if err := tx.GetContext(ctx, &updated, query, inviteID, workspaceID, strings.ToUpper(status), timestamp); err != nil {
		return Invite{}, fmt.Errorf("update workspace invite status: %w", err)
	}
	return updated, nil
}

func (repo *repository) AddWorkspaceMember(ctx context.Context, tx *sqlx.Tx, workspaceID, userID, role string) error {
	query := `
		INSERT INTO tbl_workspace_member (workspace_id, user_id, role_id)
		SELECT $1, $2, e.id
		FROM tbl_enum e
		WHERE e.category = 'WORKSPACE_ROLE'
		  AND e.code = $3
		ON CONFLICT (workspace_id, user_id) DO NOTHING`

	if _, err := tx.ExecContext(ctx, query, workspaceID, userID, strings.ToUpper(role)); err != nil {
		return fmt.Errorf("add workspace member: %w", err)
	}
	return nil
}

func (repo *repository) GetWorkspaceInfo(ctx context.Context, workspaceID string) (WorkspaceInfo, error) {
	var info WorkspaceInfo
	query := `SELECT id, name FROM tbl_workspace WHERE id = $1`
	if err := repo.db.GetContext(ctx, &info, query, workspaceID); err != nil {
		return WorkspaceInfo{}, fmt.Errorf("get workspace info: %w", err)
	}
	return info, nil
}

func (repo *repository) GetUserInfo(ctx context.Context, userID string) (UserInfo, error) {
	var info UserInfo
	query := `SELECT id, email, name FROM tbl_user WHERE id = $1`
	if err := repo.db.GetContext(ctx, &info, query, userID); err != nil {
		return UserInfo{}, fmt.Errorf("get user info: %w", err)
	}
	return info, nil
}
