package core

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/jmoiron/sqlx"
)

var (
	ErrOwnerRequired         = errors.New("organization owner is required")
	ErrOrganizationSlugTaken = errors.New("organization slug is already taken")
)

type Service interface {
	Create(ctx context.Context, userID string, input CreateOrganizationRequest) (Organization, error)
	FindByID(ctx context.Context, id string) (Organization, error)
	FindBySlug(ctx context.Context, slug string) (Organization, error)
	ListByUser(ctx context.Context, userID string) ([]Organization, error)
	Update(ctx context.Context, id, userID string, input UpdateOrganizationRequest) (Organization, error)
	Delete(ctx context.Context, id, userID string) error
	ValidateAccess(ctx context.Context, userID, organizationID string) error
}

type service struct {
	db   *sqlx.DB
	repo Repository
}

func NewService(db *sqlx.DB, repo Repository) Service {
	return &service{db: db, repo: repo}
}

func (s *service) Create(ctx context.Context, userID string, input CreateOrganizationRequest) (Organization, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return Organization{}, ErrOwnerRequired
	}

	baseSlug := slug.Make(input.Name)
	if baseSlug == "" {
		baseSlug = fmt.Sprintf("org-%d", time.Now().Unix())
	}

	org := Organization{
		Name:        input.Name,
		Slug:        baseSlug,
		Description: input.Description,
		Settings:    make(map[string]interface{}),
		Status:      StatusActive,
		CreatedBy:   userID,
	}

	// Create organization and add owner as member in transaction
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return Organization{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	var orgID string
	for suffix := 0; ; suffix++ {
		org.Slug = baseSlug
		if suffix > 0 {
			org.Slug = fmt.Sprintf("%s-%d", baseSlug, suffix)
		}
		orgID, err = s.repo.Create(ctx, tx, org)
		if !errors.Is(err, ErrOrganizationSlugTaken) {
			break
		}
	}
	if err != nil {
		return Organization{}, err
	}

	// Add creator as owner
	_, err = tx.ExecContext(ctx, `
		INSERT INTO tbl_organization_member (organization_id, user_id, role)
		VALUES ($1, $2, $3)
	`, orgID, userID, "OWNER")
	if err != nil {
		return Organization{}, fmt.Errorf("add owner member: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return Organization{}, fmt.Errorf("commit transaction: %w", err)
	}

	org.ID = orgID
	return org, nil
}

func (s *service) FindByID(ctx context.Context, id string) (Organization, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *service) FindBySlug(ctx context.Context, slug string) (Organization, error) {
	return s.repo.FindBySlug(ctx, slug)
}

func (s *service) ListByUser(ctx context.Context, userID string) ([]Organization, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *service) Update(ctx context.Context, id, userID string, input UpdateOrganizationRequest) (Organization, error) {
	// Validate access
	if err := s.ValidateAccess(ctx, userID, id); err != nil {
		return Organization{}, err
	}

	updates := make(map[string]interface{})
	if input.Name != "" {
		updates["name"] = input.Name
		updates["slug"] = slug.Make(input.Name)
	}
	if input.Description != "" {
		updates["description"] = input.Description
	}
	if input.Status != "" {
		updates["status"] = input.Status
	}

	if err := s.repo.Update(ctx, id, updates); err != nil {
		return Organization{}, err
	}

	return s.repo.FindByID(ctx, id)
}

func (s *service) Delete(ctx context.Context, id, userID string) error {
	// Validate access (must be owner)
	if err := s.ValidateAccess(ctx, userID, id); err != nil {
		return err
	}

	return s.repo.Delete(ctx, id)
}

func (s *service) ValidateAccess(ctx context.Context, userID, organizationID string) error {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM tbl_organization_member
			WHERE organization_id = $1 AND user_id = $2 AND deleted_at IS NULL
		)
	`
	err := s.db.GetContext(ctx, &exists, query, organizationID, userID)
	if err != nil {
		return fmt.Errorf("check organization access: %w", err)
	}
	if !exists {
		return errors.New("access denied to organization")
	}
	return nil
}
