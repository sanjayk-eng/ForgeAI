package project

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

type OrganizationService interface {
	ValidateAccess(ctx context.Context, userID, organizationID string) error
}

type Service interface {
	Create(ctx context.Context, organizationID, createdBy string, input CreateProjectRequest) (Project, error)
	ListByOrganization(ctx context.Context, organizationID string, query pagination.Query) (pagination.Result[Project], error)
	FindByID(ctx context.Context, projectID string) (Project, error)
	FindByOrganizationSlug(ctx context.Context, organizationID, slug string) (Project, error)
	Update(ctx context.Context, projectID, userID string, input UpdateProjectRequest) (Project, error)
	Delete(ctx context.Context, projectID, userID string) error
}

type service struct {
	db   *sqlx.DB
	repo Repository
	org  OrganizationService
}

func NewService(db *sqlx.DB, repo Repository, org OrganizationService) Service {
	return &service{db: db, repo: repo, org: org}
}

func (s *service) Create(ctx context.Context, organizationID, createdBy string, input CreateProjectRequest) (Project, error) {
	organizationID = strings.TrimSpace(organizationID)
	createdBy = strings.TrimSpace(createdBy)
	name := strings.TrimSpace(input.Name)

	if s.db == nil || organizationID == "" || createdBy == "" || name == "" {
		return Project{}, ErrInvalidProjectInput
	}

	// Validate organization access
	if err := s.org.ValidateAccess(ctx, createdBy, organizationID); err != nil {
		return Project{}, ErrOrganizationAccessRequired
	}

	project, err := s.repo.Create(ctx, organizationID, name, slugify(name), input.Description, createdBy)
	if err != nil {
		return Project{}, fmt.Errorf("create project: %w", err)
	}

	return project, nil
}

func (s *service) ListByOrganization(ctx context.Context, organizationID string, query pagination.Query) (pagination.Result[Project], error) {
	if strings.TrimSpace(organizationID) == "" {
		return pagination.Result[Project]{}, ErrInvalidProjectInput
	}
	return s.repo.ListByOrganization(ctx, organizationID, query)
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

func (s *service) FindByOrganizationSlug(ctx context.Context, organizationID, slug string) (Project, error) {
	if strings.TrimSpace(organizationID) == "" || strings.TrimSpace(slug) == "" {
		return Project{}, ErrInvalidProjectInput
	}
	project, err := s.repo.FindByOrganizationSlug(ctx, organizationID, slug)
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

	// Validate organization access
	if err := s.org.ValidateAccess(ctx, userID, existing.OrganizationID); err != nil {
		return Project{}, ErrOrganizationAccessRequired
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

	// Validate organization access
	if err := s.org.ValidateAccess(ctx, userID, project.OrganizationID); err != nil {
		return ErrOrganizationAccessRequired
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
