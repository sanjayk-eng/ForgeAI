package project

import (
	"errors"
	"time"
)

type ProjectType string

const (
	ProjectTypeRepository ProjectType = "REPOSITORY"
	ProjectTypeEmpty      ProjectType = "EMPTY"
)

type ProjectStatus string

const (
	ProjectStatusActive   ProjectStatus = "ACTIVE"
	ProjectStatusArchived ProjectStatus = "ARCHIVED"
)

type SyncStatus string

const (
	SyncStatusPending SyncStatus = "PENDING"
	SyncStatusSyncing SyncStatus = "SYNCING"
	SyncStatusSynced  SyncStatus = "SYNCED"
	SyncStatusFailed  SyncStatus = "FAILED"
)

type Project struct {
	ID          string             `json:"id" db:"id"`
	WorkspaceID string             `json:"workspace_id" db:"workspace_id"`
	Name        string             `json:"name" db:"name"`
	Slug        string             `json:"slug" db:"slug"`
	Description *string            `json:"description,omitempty" db:"description"`
	Type        ProjectType        `json:"type" db:"type"`
	Status      ProjectStatus      `json:"status" db:"status"`
	CreatedBy   string             `json:"created_by" db:"created_by"`
	CreatedAt   time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" db:"updated_at"`
	Repository  *ProjectRepository `json:"repository,omitempty"`
}

type ProjectRepository struct {
	ID                   string     `json:"id" db:"repository_id"`
	ProjectID            string     `json:"project_id" db:"project_id"`
	GitHubRepositoryID   int64      `json:"github_repository_id" db:"github_repository_id"`
	GitHubOwner          string     `json:"github_owner" db:"github_owner"`
	GitHubRepositoryName string     `json:"github_repository_name" db:"github_repository_name"`
	RepositoryURL        string     `json:"repository_url" db:"repository_url"`
	DefaultBranch        string     `json:"default_branch" db:"default_branch"`
	SyncStatus           SyncStatus `json:"sync_status" db:"sync_status"`
	LastSyncedAt         *time.Time `json:"last_synced_at,omitempty" db:"last_synced_at"`
	CreatedAt            time.Time  `json:"created_at" db:"repository_created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"repository_updated_at"`
}

type CreateProjectRequest struct {
	Name        string                    `json:"name" binding:"required,min=1,max=150"`
	Description *string                   `json:"description"`
	Type        ProjectType               `json:"type" binding:"required,oneof=REPOSITORY EMPTY"`
	Repository  *ConnectRepositoryRequest `json:"repository,omitempty"`
}

type UpdateProjectRequest struct {
	Name        string        `json:"name" binding:"required,min=1,max=150"`
	Description *string       `json:"description"`
	Status      ProjectStatus `json:"status" binding:"required,oneof=ACTIVE ARCHIVED"`
}

type ConnectRepositoryRequest struct {
	GitHubRepositoryID   int64  `json:"github_repository_id" binding:"required"`
	GitHubOwner          string `json:"github_owner" binding:"required,max=255"`
	GitHubRepositoryName string `json:"github_repository_name" binding:"required,max=255"`
	RepositoryURL        string `json:"repository_url" binding:"required"`
	DefaultBranch        string `json:"default_branch" binding:"required,max=255"`
}

type ResolveRepositoryRequest struct {
	RepositoryURL string `json:"repository_url" binding:"required,url"`
}

type ResolvedRepository struct {
	Repository ConnectRepositoryRequest `json:"repository"`
	Branches   []string                 `json:"branches"`
}

var (
	ErrInvalidProjectInput = errors.New("invalid project input")
	ErrProjectNotFound     = errors.New("project not found")
	ErrRepositoryConflict  = errors.New("repository is already connected")
)
