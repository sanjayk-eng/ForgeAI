package member

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Service interface {
	List(ctx context.Context, organizationID string) ([]MemberWithUser, error)
	UpdateRole(ctx context.Context, memberID, role string) (Member, error)
	Remove(ctx context.Context, memberID string) error
}

type service struct {
	db   *sqlx.DB
	repo Repository
}

func NewService(db *sqlx.DB, repo Repository) Service {
	return &service{db: db, repo: repo}
}

func (s *service) List(ctx context.Context, organizationID string) ([]MemberWithUser, error) {
	return s.repo.ListByOrganization(ctx, organizationID)
}

func (s *service) UpdateRole(ctx context.Context, memberID, role string) (Member, error) {
	// Validate role
	if role != RoleAdmin && role != RoleMember {
		return Member{}, fmt.Errorf("invalid role: %s", role)
	}

	if err := s.repo.UpdateRole(ctx, memberID, role); err != nil {
		return Member{}, err
	}

	return s.repo.FindByID(ctx, memberID)
}

func (s *service) Remove(ctx context.Context, memberID string) error {
	return s.repo.Remove(ctx, memberID)
}
