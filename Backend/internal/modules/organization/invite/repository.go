package invite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var (
	ErrNotFound      = errors.New("invitation not found")
	ErrExpired       = errors.New("invitation has expired")
	ErrAlreadyUsed   = errors.New("invitation is no longer pending")
	ErrEmailMismatch = errors.New("invitation email does not match the signed-in user")
	ErrDuplicate     = errors.New("a pending invitation already exists for this email")
)

type Repository interface {
	Create(context.Context, Invite, string) (Invite, error)
	ListByOrganization(context.Context, string, string) ([]Invite, error)
	FindByToken(context.Context, string) (Invite, error)
	ListPendingByEmail(context.Context, string) ([]Invite, error)
	Revoke(context.Context, string, string) error
	OrganizationIDByID(context.Context, string) (string, error)
	Reject(context.Context, string, string) error
	RejectByID(context.Context, string, string) error
	Accept(context.Context, string, string, string) (string, error)
	AcceptByID(context.Context, string, string, string) (string, error)
	UserEmail(context.Context, string) (string, error)
	UserName(context.Context, string) (string, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

const inviteSelect = `
	SELECT i.id, i.organization_id, o.name AS organization_name, i.email, i.role,
	       i.status, i.invited_by, u.name AS inviter_name, u.email AS inviter_email,
	       i.expires_at, i.created_at, i.updated_at
	FROM tbl_organization_invite i
	JOIN tbl_organization o ON o.id = i.organization_id
	JOIN tbl_user u ON u.id = i.invited_by
`

func (r *repository) Create(ctx context.Context, invite Invite, tokenHash string) (Invite, error) {
	query := `
		INSERT INTO tbl_organization_invite
			(organization_id, email, role, status, invited_by, token_hash, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowxContext(ctx, query,
		invite.OrganizationID, invite.Email, invite.Role, StatusPending,
		invite.InvitedBy, tokenHash, invite.ExpiresAt,
	).Scan(&invite.ID, &invite.CreatedAt, &invite.UpdatedAt)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return Invite{}, ErrDuplicate
		}
		return Invite{}, fmt.Errorf("create organization invitation: %w", err)
	}
	invite.Status = StatusPending
	invite.setInviter()
	return invite, nil
}

func (r *repository) ListByOrganization(ctx context.Context, organizationID, status string) ([]Invite, error) {
	if _, err := r.db.ExecContext(ctx, `
		UPDATE tbl_organization_invite SET status = $1, updated_at = NOW()
		WHERE organization_id = $2 AND status = $3 AND expires_at <= NOW()`,
		StatusExpired, organizationID, StatusPending); err != nil {
		return nil, fmt.Errorf("expire organization invitations: %w", err)
	}
	invites := make([]Invite, 0)
	err := r.db.SelectContext(ctx, &invites, inviteSelect+`
		WHERE i.organization_id = $1 AND ($2 = '' OR i.status = $2)
		ORDER BY i.created_at DESC`, organizationID, status)
	if err != nil {
		return nil, fmt.Errorf("list organization invitations: %w", err)
	}
	for index := range invites {
		invites[index].setInviter()
	}
	return invites, nil
}

func (r *repository) FindByToken(ctx context.Context, tokenHash string) (Invite, error) {
	_, err := r.db.ExecContext(ctx, `
		UPDATE tbl_organization_invite SET status = $1, updated_at = NOW()
		WHERE token_hash = $2 AND status = $3 AND expires_at <= NOW()`,
		StatusExpired, tokenHash, StatusPending)
	if err != nil {
		return Invite{}, fmt.Errorf("expire organization invitation: %w", err)
	}
	var invite Invite
	err = r.db.GetContext(ctx, &invite, inviteSelect+` WHERE i.token_hash = $1`, tokenHash)
	if errors.Is(err, sql.ErrNoRows) {
		return Invite{}, ErrNotFound
	}
	if err != nil {
		return Invite{}, fmt.Errorf("find organization invitation: %w", err)
	}
	invite.setInviter()
	return invite, nil
}

func (r *repository) ListPendingByEmail(ctx context.Context, email string) ([]Invite, error) {
	invites := make([]Invite, 0)
	err := r.db.SelectContext(ctx, &invites, inviteSelect+`
		WHERE lower(i.email) = lower($1) AND i.status = $2 AND i.expires_at > NOW()
		ORDER BY i.created_at DESC`, email, StatusPending)
	if err != nil {
		return nil, fmt.Errorf("list pending invitations: %w", err)
	}
	for index := range invites {
		invites[index].setInviter()
	}
	return invites, nil
}

func (r *repository) Revoke(ctx context.Context, organizationID, inviteID string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE tbl_organization_invite SET status = $1, updated_at = NOW()
		WHERE organization_id = $2 AND id = $3 AND status = $4`,
		StatusRevoked, organizationID, inviteID, StatusPending)
	if err != nil {
		return fmt.Errorf("revoke organization invitation: %w", err)
	}
	return requireRows(result)
}

func (r *repository) OrganizationIDByID(ctx context.Context, inviteID string) (string, error) {
	var organizationID string
	err := r.db.GetContext(ctx, &organizationID, `SELECT organization_id FROM tbl_organization_invite WHERE id = $1`, inviteID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("find invitation organization: %w", err)
	}
	return organizationID, nil
}

func (r *repository) Reject(ctx context.Context, tokenHash, email string) error {
	return r.reject(ctx, "token_hash", tokenHash, email)
}

func (r *repository) RejectByID(ctx context.Context, inviteID, email string) error {
	return r.reject(ctx, "id", inviteID, email)
}

func (r *repository) reject(ctx context.Context, keyColumn, key, email string) error {
	if keyColumn != "token_hash" && keyColumn != "id" {
		return ErrNotFound
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE tbl_organization_invite SET status = $1, updated_at = NOW()
		WHERE `+keyColumn+` = $2 AND status = $3 AND expires_at <= NOW()`,
		StatusExpired, key, StatusPending)
	if err != nil {
		return fmt.Errorf("expire organization invitation: %w", err)
	}
	result, err := r.db.ExecContext(ctx, `
		UPDATE tbl_organization_invite SET status = $1, updated_at = NOW()
		WHERE `+keyColumn+` = $2 AND lower(email) = lower($3) AND status = $4 AND expires_at > NOW()`,
		StatusRejected, key, strings.TrimSpace(email), StatusPending)
	if err != nil {
		return fmt.Errorf("reject organization invitation: %w", err)
	}
	return requireRows(result)
}

func (r *repository) Accept(ctx context.Context, tokenHash, userID, email string) (string, error) {
	return r.accept(ctx, "token_hash", tokenHash, userID, email)
}

func (r *repository) AcceptByID(ctx context.Context, inviteID, userID, email string) (string, error) {
	return r.accept(ctx, "id", inviteID, userID, email)
}

func (r *repository) accept(ctx context.Context, keyColumn, key, userID, email string) (string, error) {
	if keyColumn != "token_hash" && keyColumn != "id" {
		return "", ErrNotFound
	}
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin accept invitation transaction: %w", err)
	}
	defer tx.Rollback()

	var inviteID, organizationID, inviteEmail, role, status string
	var expiresAt time.Time
	err = tx.QueryRowxContext(ctx, `
		SELECT id, organization_id, email, role, status, expires_at
		FROM tbl_organization_invite WHERE `+keyColumn+` = $1 FOR UPDATE`, key,
	).Scan(&inviteID, &organizationID, &inviteEmail, &role, &status, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("load organization invitation: %w", err)
	}
	if status != StatusPending {
		return "", ErrAlreadyUsed
	}
	if !expiresAt.After(time.Now()) {
		if _, err := tx.ExecContext(ctx, `UPDATE tbl_organization_invite SET status = $1, updated_at = NOW() WHERE id = $2`, StatusExpired, inviteID); err != nil {
			return "", fmt.Errorf("expire organization invitation: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return "", fmt.Errorf("commit expired invitation: %w", err)
		}
		return "", ErrExpired
	}
	if !strings.EqualFold(strings.TrimSpace(inviteEmail), strings.TrimSpace(email)) {
		return "", ErrEmailMismatch
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO tbl_organization_member (organization_id, user_id, role)
		VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`, organizationID, userID, role); err != nil {
		return "", fmt.Errorf("add invited organization member: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE tbl_organization_invite
		SET status = $1, accepted_by = $2, accepted_at = NOW(), updated_at = NOW()
		WHERE id = $3`, StatusAccepted, userID, inviteID); err != nil {
		return "", fmt.Errorf("mark organization invitation accepted: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit accepted invitation: %w", err)
	}
	return organizationID, nil
}

func (r *repository) UserEmail(ctx context.Context, userID string) (string, error) {
	var email string
	if err := r.db.GetContext(ctx, &email, `SELECT email FROM tbl_user WHERE id = $1`, userID); err != nil {
		return "", fmt.Errorf("find invitation user email: %w", err)
	}
	return email, nil
}

func (r *repository) UserName(ctx context.Context, userID string) (string, error) {
	var name string
	if err := r.db.GetContext(ctx, &name, `SELECT name FROM tbl_user WHERE id = $1`, userID); err != nil {
		return "", fmt.Errorf("find invitation sender name: %w", err)
	}
	return name, nil
}

func requireRows(result sql.Result) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check invitation update: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
