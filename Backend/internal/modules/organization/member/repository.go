package member

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

var (
	ErrMemberNotFound = errors.New("member not found")
)

type Repository interface {
	ListByOrganization(ctx context.Context, organizationID string) ([]MemberWithUser, error)
	FindByID(ctx context.Context, id string) (Member, error)
	FindByUserAndOrganization(ctx context.Context, userID, organizationID string) (Member, error)
	UpdateRole(ctx context.Context, id, role string) error
	Remove(ctx context.Context, id string) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) ListByOrganization(ctx context.Context, organizationID string) ([]MemberWithUser, error) {
	query := `
		SELECT 
			m.*,
			u.name as user_name,
			u.email as user_email
		FROM tbl_organization_member m
		INNER JOIN tbl_user u ON m.user_id = u.id
		WHERE m.organization_id = $1 AND m.deleted_at IS NULL
		ORDER BY m.created_at ASC
	`
	var members []MemberWithUser
	err := r.db.SelectContext(ctx, &members, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("list organization members: %w", err)
	}
	return members, nil
}

func (r *repository) FindByID(ctx context.Context, id string) (Member, error) {
	var member Member
	query := `SELECT * FROM tbl_organization_member WHERE id = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &member, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Member{}, ErrMemberNotFound
		}
		return Member{}, fmt.Errorf("find member by id: %w", err)
	}
	return member, nil
}

func (r *repository) FindByUserAndOrganization(ctx context.Context, userID, organizationID string) (Member, error) {
	var member Member
	query := `
		SELECT * FROM tbl_organization_member
		WHERE user_id = $1 AND organization_id = $2 AND deleted_at IS NULL
	`
	err := r.db.GetContext(ctx, &member, query, userID, organizationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Member{}, ErrMemberNotFound
		}
		return Member{}, fmt.Errorf("find member: %w", err)
	}
	return member, nil
}

func (r *repository) UpdateRole(ctx context.Context, id, role string) error {
	query := `
		UPDATE tbl_organization_member
		SET role = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`
	result, err := r.db.ExecContext(ctx, query, role, id)
	if err != nil {
		return fmt.Errorf("update member role: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check update result: %w", err)
	}
	if rows == 0 {
		return ErrMemberNotFound
	}
	return nil
}

func (r *repository) Remove(ctx context.Context, id string) error {
	query := `
		UPDATE tbl_organization_member
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("remove member: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check remove result: %w", err)
	}
	if rows == 0 {
		return ErrMemberNotFound
	}
	return nil
}
