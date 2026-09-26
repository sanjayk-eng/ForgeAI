package core

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"ai-agent/internal/shared/pagination"

	"github.com/jmoiron/sqlx"
)

type Service interface {
	Create(ctx context.Context, tx *sqlx.Tx, workspaceID, createdBy string, input CreateProjectRequest) (Project, error)
	ListByWorkspace(ctx context.Context, workspaceID string, query pagination.Query) (pagination.Result[Project], error)
	FindByID(ctx context.Context, projectID string) (Project, error)
	FindByWorkspaceSlug(ctx context.Context, workspaceID, slug string) (Project, error)
	Update(ctx context.Context, projectID, userID string, input UpdateProjectRequest) (Project, error)
	Delete(ctx context.Context, projectID, userID string) error
}

type service struct {
	db   *sqlx.DB
	repo Repository
}

func NewService(db *sqlx.DB, repo Repository) Service {
	return &service{db: db, repo: repo}
}

func (s *service) Create(ctx context.Context, tx *sqlx.Tx, workspaceID, createdBy string, input CreateProjectRequest) (Project, error) {
	workspaceID = strings.TrimSpace(workspaceID)
	createdBy = strings.TrimSpace(createdBy)
	name := strings.TrimSpace(input.Name)

	if s.db == nil || workspaceID == "" || createdBy == "" || name == "" {
		return Project{}, ErrInvalidProjectInput
	}

	project, err := s.repo.Create(ctx, tx, workspaceID, name, slugify(name), input.Description, string(input.Type), createdBy)
	if err != nil {
		return Project{}, fmt.Errorf("create project: %w", err)
	}

	return project, nil
}

func (s *service) ListByWorkspace(ctx context.Context, workspaceID string, query pagination.Query) (pagination.Result[Project], error) {
	if strings.TrimSpace(workspaceID) == "" {
		return pagination.Result[Project]{}, ErrInvalidProjectInput
	}
	return s.repo.ListByWorkspace(ctx, workspaceID, query)
}

func (s *service) FindByID(ctx context.Context, projectID string) (Project, error) {
	if strings.TrimSpace(projectID) == "" {
		return Project{}, ErrInvalidProjectInput
	}
	project, err := s.repo.FindByID(ctx, projectID)
	if err != nil {
		return Project{}, ErrProjectNotFound
	}
	return project, nil
}

func (s *service) FindByWorkspaceSlug(ctx context.Context, workspaceID, slug string) (Project, error) {
	if strings.TrimSpace(workspaceID) == "" || strings.TrimSpace(slug) == "" {
		return Project{}, ErrInvalidProjectInput
	}
	project, err := s.repo.FindByWorkspaceSlug(ctx, workspaceID, slug)
	if err != nil {
		return Project{}, ErrProjectNotFound
	}
	return project, nil
}

func (s *service) Update(ctx context.Context, projectID, userID string, input UpdateProjectRequest) (Project, error) {
	projectID = strings.TrimSpace(projectID)
	name := strings.TrimSpace(input.Name)
	status := input.Status
	if status == "" {
		status = ProjectStatusActive
	}

	if s.db == nil || projectID == "" || strings.TrimSpace(userID) == "" || name == "" {
		return Project{}, ErrInvalidProjectInput
	}

	existing, err := s.repo.FindByID(ctx, projectID)
	if err != nil {
		return Project{}, ErrProjectNotFound
	}

	owner, err := s.repo.IsWorkspaceOwner(ctx, existing.WorkspaceID, userID)
	if err != nil {
		return Project{}, fmt.Errorf("check project owner: %w", err)
	}
	if !owner {
		return Project{}, ErrWorkspaceOwnerRequired
	}

	project, err := s.repo.Update(ctx, projectID, name, input.Description, string(status))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Project{}, ErrProjectNotFound
		}
		return Project{}, fmt.Errorf("update project: %w", err)
	}

	return project, nil
}

func (s *service) Delete(ctx context.Context, projectID, userID string) error {
	projectID = strings.TrimSpace(projectID)
	userID = strings.TrimSpace(userID)

	if s.db == nil || projectID == "" || userID == "" {
		return ErrInvalidProjectInput
	}

	project, err := s.repo.FindByID(ctx, projectID)
	if err != nil {
		return ErrProjectNotFound
	}

	owner, err := s.repo.IsWorkspaceOwner(ctx, project.WorkspaceID, userID)
	if err != nil {
		return fmt.Errorf("check project owner: %w", err)
	}
	if !owner {
		return ErrWorkspaceOwnerRequired
	}

	if err := s.repo.Delete(ctx, projectID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrProjectNotFound
		}
		return fmt.Errorf("delete project: %w", err)
	}

	return nil
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(value, "-")
	return strings.Trim(strings.TrimSpace(value), "-")
}
