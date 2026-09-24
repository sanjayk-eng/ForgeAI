package member

import (
	"context"
	"fmt"
	"strings"

	"ai-agent/internal/shared/pagination"
	appdatabase "ai-agent/pkg/database"

	"github.com/jmoiron/sqlx"
)

type Service struct {
	db   *sqlx.DB
	repo Repository
}

func NewService(db *sqlx.DB, repo Repository) *Service {
	return &Service{db: db, repo: repo}
}

func (service *Service) AddMember(ctx context.Context, workspaceID, inviterID, userID, role string) (Member, error) {
	workspaceID = strings.TrimSpace(workspaceID)
	inviterID = strings.TrimSpace(inviterID)
	userID = strings.TrimSpace(userID)
	if service.db == nil || workspaceID == "" || inviterID == "" || userID == "" {
		return Member{}, ErrInvalidMemberInput
	}

	memberRole := NormalizeRole(role)
	member := Member{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Role:        memberRole,
	}

	if err := appdatabase.WithTx(ctx, service.db, func(ctx context.Context, tx *sqlx.Tx) error {
		return service.repo.AddMember(ctx, tx, workspaceID, userID, string(memberRole))
	}); err != nil {
		return Member{}, fmt.Errorf("add workspace member: %w", err)
	}

	return member, nil
}

func (service *Service) ListMembers(ctx context.Context, workspaceID string, query pagination.Query) (pagination.Result[Member], error) {
	workspaceID = strings.TrimSpace(workspaceID)
	if workspaceID == "" {
		return pagination.Result[Member]{}, ErrInvalidMemberInput
	}
	return service.repo.ListMembers(ctx, workspaceID, query)
}

func (service *Service) UpdateMemberRole(ctx context.Context, workspaceID, userID, role string) error {
	workspaceID = strings.TrimSpace(workspaceID)
	userID = strings.TrimSpace(userID)
	if service.db == nil || workspaceID == "" || userID == "" {
		return ErrInvalidMemberInput
	}

	return appdatabase.WithTx(ctx, service.db, func(ctx context.Context, tx *sqlx.Tx) error {
		return service.repo.UpdateMemberRole(ctx, tx, workspaceID, userID, string(NormalizeRole(role)))
	})
}

func (service *Service) RemoveMember(ctx context.Context, workspaceID, userID string) error {
	workspaceID = strings.TrimSpace(workspaceID)
	userID = strings.TrimSpace(userID)
	if service.db == nil || workspaceID == "" || userID == "" {
		return ErrInvalidMemberInput
	}

	return appdatabase.WithTx(ctx, service.db, func(ctx context.Context, tx *sqlx.Tx) error {
		return service.repo.RemoveMember(ctx, tx, workspaceID, userID)
	})
}
