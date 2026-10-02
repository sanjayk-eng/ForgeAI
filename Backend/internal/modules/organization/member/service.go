package member

import (
	"ai-agent/internal/shared/pagination"
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
)

type Service interface {
	List(ctx context.Context, organizationID string, query pagination.Query) ([]MemberWithUser, int, error)
	CanManage(ctx context.Context, userID, organizationID string) error
	UpdateRole(ctx context.Context, organizationID, userID, role string) (Member, error)
	Remove(ctx context.Context, organizationID, userID string) error
}

var (
	ErrManageDenied   = errors.New("organization owner or admin permission is required")
	ErrOwnerProtected = errors.New("organization owner cannot be changed or removed")
	ErrInvalidRole    = errors.New("invalid organization member role")
)

type service struct {
	db   *sqlx.DB
	repo Repository
}

func NewService(db *sqlx.DB, repo Repository) Service {
	return &service{db: db, repo: repo}
}

func (s *service) List(ctx context.Context, organizationID string, query pagination.Query) ([]MemberWithUser, int, error) {
	return s.repo.ListByOrganization(ctx, organizationID, query)
}

func (s *service) CanManage(ctx context.Context, userID, organizationID string) error {
	allowed, err := s.repo.CanManage(ctx, userID, organizationID)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrManageDenied
	}
	return nil
}

func (s *service) UpdateRole(ctx context.Context, organizationID, userID, role string) (Member, error) {
	if role != RoleAdmin && role != RoleMember {
		return Member{}, ErrInvalidRole
	}
	member, err := s.repo.FindByUserAndOrganization(ctx, userID, organizationID)
	if err != nil {
		return Member{}, err
	}
	if member.Role == RoleOwner {
		return Member{}, ErrOwnerProtected
	}
	if err := s.repo.UpdateRole(ctx, member.ID, role); err != nil {
		return Member{}, err
	}
	return s.repo.FindByID(ctx, member.ID)
}

func (s *service) Remove(ctx context.Context, organizationID, userID string) error {
	member, err := s.repo.FindByUserAndOrganization(ctx, userID, organizationID)
	if err != nil {
		return err
	}
	if member.Role == RoleOwner {
		return ErrOwnerProtected
	}
	return s.repo.Remove(ctx, member.ID)
}
