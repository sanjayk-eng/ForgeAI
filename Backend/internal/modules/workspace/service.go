package workspace

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	appdatabase "ai-agent/pkg/database"

	"github.com/jmoiron/sqlx"
)

type Service struct {
	db   *sqlx.DB
	repo WorkspaceRepository
}

func NewService(db *sqlx.DB, repo WorkspaceRepository) *Service {
	return &Service{db: db, repo: repo}
}

func (service *Service) CreateWorkspace(ctx context.Context, ownerID string, input CreateWorkspaceRequest) (Workspace, error) {
	if service.db == nil || strings.TrimSpace(ownerID) == "" || strings.TrimSpace(input.Name) == "" {
		return Workspace{}, ErrInvalidWorkspaceInput
	}

	slug := slugify(input.Name)
	if slug == "" {
		return Workspace{}, ErrInvalidWorkspaceInput
	}

	var workspace Workspace
	if err := appdatabase.WithTx(ctx, service.db, func(ctx context.Context, tx *sqlx.Tx) error {
		var err error
		workspace, err = service.repo.Create(ctx, tx, input.Name, slug, ownerID)
		if err != nil {
			return err
		}
		return service.repo.AddMember(ctx, tx, workspace.ID, ownerID, string(RoleOwner))
	}); err != nil {
		return Workspace{}, fmt.Errorf("create workspace: %w", err)
	}

	return workspace, nil
}

func (service *Service) GetWorkspace(ctx context.Context, workspaceID string) (Workspace, error) {
	if strings.TrimSpace(workspaceID) == "" {
		return Workspace{}, ErrInvalidWorkspaceInput
	}
	workspace, err := service.repo.FindByID(ctx, workspaceID)
	if err != nil {
		return Workspace{}, ErrWorkspaceNotFound
	}
	return workspace, nil
}

func (service *Service) ListWorkspaces(ctx context.Context, userID string) ([]Workspace, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrInvalidWorkspaceInput
	}
	return service.repo.ListByUser(ctx, userID)
}

func (service *Service) AddMember(ctx context.Context, workspaceID, inviterID, userID, role string) (WorkspaceMember, error) {
	if service.db == nil || strings.TrimSpace(workspaceID) == "" || strings.TrimSpace(userID) == "" {
		return WorkspaceMember{}, ErrInvalidWorkspaceInput
	}
	if strings.TrimSpace(inviterID) == "" {
		inviterID = userID
	}

	memberRole := NormalizeRole(role)
	member := WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Role:        memberRole,
	}

	if err := appdatabase.WithTx(ctx, service.db, func(ctx context.Context, tx *sqlx.Tx) error {
		return service.repo.AddMember(ctx, tx, workspaceID, userID, string(memberRole))
	}); err != nil {
		return WorkspaceMember{}, fmt.Errorf("add workspace member: %w", err)
	}

	_ = inviterID
	return member, nil
}

func (service *Service) UpdateMemberRole(ctx context.Context, workspaceID, userID, role string) error {
	if service.db == nil || strings.TrimSpace(workspaceID) == "" || strings.TrimSpace(userID) == "" {
		return ErrInvalidWorkspaceInput
	}

	return appdatabase.WithTx(ctx, service.db, func(ctx context.Context, tx *sqlx.Tx) error {
		return service.repo.UpdateMemberRole(ctx, tx, workspaceID, userID, string(NormalizeRole(role)))
	})
}

func (service *Service) RemoveMember(ctx context.Context, workspaceID, userID string) error {
	if service.db == nil || strings.TrimSpace(workspaceID) == "" || strings.TrimSpace(userID) == "" {
		return ErrInvalidWorkspaceInput
	}

	return appdatabase.WithTx(ctx, service.db, func(ctx context.Context, tx *sqlx.Tx) error {
		return service.repo.RemoveMember(ctx, tx, workspaceID, userID)
	})
}

func (service *Service) ListMembers(ctx context.Context, workspaceID string) ([]WorkspaceMember, error) {
	if strings.TrimSpace(workspaceID) == "" {
		return nil, ErrInvalidWorkspaceInput
	}
	return service.repo.ListMembers(ctx, workspaceID)
}

func slugify(value string) string {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.ToLower(trimmed)
	trimmed = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(trimmed, "-")
	trimmed = strings.Trim(trimmed, "-")
	if trimmed == "" {
		return "workspace"
	}
	if len(trimmed) > 80 {
		trimmed = strings.TrimRight(trimmed[:80], "-")
	}
	return trimmed
}

func containsOnlyLettersDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
