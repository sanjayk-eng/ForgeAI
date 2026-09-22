package workspace

import (
	"errors"
	"time"
)

type WorkspaceRole string

const (
	RoleOwner  WorkspaceRole = "OWNER"
	RoleAdmin  WorkspaceRole = "ADMIN"
	RoleMember WorkspaceRole = "MEMBER"
)

type Workspace struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Slug      string    `json:"slug" db:"slug"`
	OwnerID   string    `json:"owner_id" db:"owner_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CreateWorkspaceRequest struct {
	Name string `json:"name" binding:"required,min=2,max=150"`
}

type UpdateWorkspaceRequest struct {
	Name string `json:"name" binding:"required,min=2,max=150"`
}

var (
	ErrInvalidWorkspaceInput = errors.New("invalid workspace input")
	ErrWorkspaceNotFound     = errors.New("workspace not found")
)
