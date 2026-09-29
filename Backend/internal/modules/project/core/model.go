package core

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

type Project struct {
	ID          string        `json:"id" db:"id"`
	WorkspaceID string        `json:"workspace_id" db:"workspace_id"`
	Name        string        `json:"name" db:"name"`
	Slug        string        `json:"slug" db:"slug"`
	Description *string       `json:"description,omitempty" db:"description"`
	Type        ProjectType   `json:"type" db:"type"`
	Status      ProjectStatus `json:"status" db:"status"`
	CreatedBy   string        `json:"created_by" db:"created_by"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" db:"updated_at"`
}

type CreateProjectRequest struct {
	Name        string      `json:"name" binding:"required,min=1,max=150"`
	Description *string     `json:"description"`
	Type        ProjectType `json:"type" binding:"required,oneof=REPOSITORY EMPTY"`
}

type UpdateProjectRequest struct {
	Name        string        `json:"name" binding:"required,min=1,max=150"`
	Description *string       `json:"description"`
	Status      ProjectStatus `json:"status,omitempty" binding:"omitempty,oneof=ACTIVE ARCHIVED"`
}

var (
	ErrInvalidProjectInput    = errors.New("invalid project input")
	ErrProjectNotFound        = errors.New("project not found")
	ErrWorkspaceOwnerRequired = errors.New("only the workspace owner can perform this action")
)
