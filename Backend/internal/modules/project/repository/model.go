package repository

import (
	"errors"
	"time"
)

type SyncStatus string

const (
	SyncStatusPending SyncStatus = "PENDING"
	SyncStatusSyncing SyncStatus = "SYNCING"
	SyncStatusSynced  SyncStatus = "SYNCED"
	SyncStatusFailed  SyncStatus = "FAILED"
)

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

type ConnectRepositoryRequest struct {
	GitHubRepositoryID   int64  `json:"github_repository_id" binding:"required"`
	GitHubOwner          string `json:"github_owner" binding:"required,max=255"`
	GitHubRepositoryName string `json:"github_repository_name" binding:"required,max=255"`
	RepositoryURL        string `json:"repository_url" binding:"required"`
	Private              bool   `json:"private"`
	DefaultBranch        string `json:"default_branch" binding:"required,max=255"`
}

type UpdateRepositoryBranchRequest struct {
	DefaultBranch string `json:"default_branch" binding:"required,max=255"`
}

type ResolveRepositoryRequest struct {
	RepositoryURL string `json:"repository_url" binding:"required,url"`
}

type ResolvedRepository struct {
	Repository ConnectRepositoryRequest `json:"repository"`
	Branches   []string                 `json:"branches"`
}

var (
	ErrInvalidRepositoryInput = errors.New("invalid repository input")
	ErrRepositoryNotFound     = errors.New("repository not found")
	ErrRepositoryConflict     = errors.New("repository is already connected")
)
