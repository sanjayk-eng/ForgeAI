package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Service interface {
	ConnectRepository(ctx context.Context, tx *sqlx.Tx, projectID, workspaceID string, input ConnectRepositoryRequest) (ProjectRepository, error)
	FindByProjectID(ctx context.Context, projectID string) (ProjectRepository, error)
	UpdateBranch(ctx context.Context, projectID, branch string) (ProjectRepository, error)
	RepositoryExists(ctx context.Context, tx *sqlx.Tx, workspaceID string, githubRepositoryID int64) (bool, error)
}

type service struct {
	db   *sqlx.DB
	repo Repository
}

func NewService(db *sqlx.DB, repo Repository) Service {
	return &service{db: db, repo: repo}
}

func (s *service) ConnectRepository(ctx context.Context, tx *sqlx.Tx, projectID, workspaceID string, input ConnectRepositoryRequest) (ProjectRepository, error) {
	if s.db == nil || strings.TrimSpace(projectID) == "" || input.GitHubRepositoryID <= 0 ||
		strings.TrimSpace(input.GitHubOwner) == "" || strings.TrimSpace(input.GitHubRepositoryName) == "" ||
		strings.TrimSpace(input.RepositoryURL) == "" || strings.TrimSpace(input.DefaultBranch) == "" {
		return ProjectRepository{}, ErrInvalidRepositoryInput
	}

	repository, err := s.repo.ConnectRepository(ctx, tx, projectID, workspaceID, input)
	if err != nil {
		return ProjectRepository{}, fmt.Errorf("connect repository: %w", err)
	}

	return repository, nil
}

func (s *service) FindByProjectID(ctx context.Context, projectID string) (ProjectRepository, error) {
	if strings.TrimSpace(projectID) == "" {
		return ProjectRepository{}, ErrInvalidRepositoryInput
	}

	repository, err := s.repo.FindByProjectID(ctx, projectID)
	if err != nil {
		return ProjectRepository{}, ErrRepositoryNotFound
	}

	return repository, nil
}

func (s *service) UpdateBranch(ctx context.Context, projectID, branch string) (ProjectRepository, error) {
	if s.db == nil || strings.TrimSpace(projectID) == "" || strings.TrimSpace(branch) == "" {
		return ProjectRepository{}, ErrInvalidRepositoryInput
	}

	return s.repo.UpdateBranch(ctx, projectID, strings.TrimSpace(branch))
}

func (s *service) RepositoryExists(ctx context.Context, tx *sqlx.Tx, workspaceID string, githubRepositoryID int64) (bool, error) {
	if strings.TrimSpace(workspaceID) == "" || githubRepositoryID <= 0 {
		return false, ErrInvalidRepositoryInput
	}

	return s.repo.RepositoryExists(ctx, tx, workspaceID, githubRepositoryID)
}
