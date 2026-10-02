package project

import (
	"errors"
	"time"
)

type ProjectStatus string

const (
	ProjectStatusActive   ProjectStatus = "ACTIVE"
	ProjectStatusArchived ProjectStatus = "ARCHIVED"
)

type Project struct {
	ID             string        `json:"id" db:"id"`
	OrganizationID string        `json:"organization_id" db:"organization_id"`
	Name           string        `json:"name" db:"name"`
	Slug           string        `json:"slug" db:"slug"`
	Description    *string       `json:"description,omitempty" db:"description"`
	Status         ProjectStatus `json:"status" db:"status"`
	CreatedBy      string        `json:"created_by" db:"created_by"`
	CreatedAt      time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at" db:"updated_at"`
	DeletedAt      *time.Time    `json:"deleted_at,omitempty" db:"deleted_at"`
}

type CreateProjectRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=150"`
	Description *string `json:"description"`
}

type UpdateProjectRequest struct {
	Name        string        `json:"name" binding:"omitempty,min=1,max=150"`
	Description *string       `json:"description"`
	Status      ProjectStatus `json:"status,omitempty" binding:"omitempty,oneof=ACTIVE ARCHIVED"`
}

var (
	ErrInvalidProjectInput        = errors.New("invalid project input")
	ErrProjectNotFound            = errors.New("project not found")
	ErrOrganizationOwnerRequired  = errors.New("only the organization owner can perform this action")
	ErrOrganizationAccessRequired = errors.New("organization access required")
)
