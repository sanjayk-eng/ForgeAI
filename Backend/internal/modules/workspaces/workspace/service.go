package workspace

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	appdatabase "ai-agent/pkg/database"

	"github.com/jmoiron/sqlx"
)

type Service interface {
	CreateWorkspace(ctx context.Context, ownerID string, input CreateWorkspaceRequest) (Workspace, error)
	GetWorkspace(ctx context.Context, workspaceID string) (Workspace, error)
	ListWorkspaces(ctx context.Context, userID string) ([]Workspace, error)
	UpdateWorkspace(ctx context.Context, workspaceID string, input UpdateWorkspaceRequest) (Workspace, error)
}

type service struct {
	db   *sqlx.DB
	repo WorkspaceRepository
}

func NewService(db *sqlx.DB, repo WorkspaceRepository) Service {
	return &service{db: db, repo: repo}
}

func (service *service) CreateWorkspace(ctx context.Context, ownerID string, input CreateWorkspaceRequest) (Workspace, error) {
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

func (service *service) GetWorkspace(ctx context.Context, workspaceID string) (Workspace, error) {
	if strings.TrimSpace(workspaceID) == "" {
		return Workspace{}, ErrInvalidWorkspaceInput
	}
	workspace, err := service.repo.FindByID(ctx, workspaceID)
	if err != nil {
		return Workspace{}, ErrWorkspaceNotFound
	}
	return workspace, nil
}

func (service *service) ListWorkspaces(ctx context.Context, userID string) ([]Workspace, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrInvalidWorkspaceInput
	}
	return service.repo.ListByUser(ctx, userID)
}

func (service *service) UpdateWorkspace(ctx context.Context, workspaceID string, input UpdateWorkspaceRequest) (Workspace, error) {
	workspaceID = strings.TrimSpace(workspaceID)
	name := strings.TrimSpace(input.Name)
	if service.db == nil || workspaceID == "" || name == "" {
		return Workspace{}, ErrInvalidWorkspaceInput
	}

	slug := slugify(name)
	if slug == "" {
		return Workspace{}, ErrInvalidWorkspaceInput
	}

	workspace, err := service.repo.Update(ctx, workspaceID, name, slug)
	if err != nil {
		return Workspace{}, ErrWorkspaceNotFound
	}
	return workspace, nil
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
