package workspace

import (
	"errors"
	"strings"
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

type WorkspaceMember struct {
	ID          string        `json:"id" db:"id"`
	WorkspaceID string        `json:"workspace_id" db:"workspace_id"`
	UserID      string        `json:"user_id" db:"user_id"`
	Role        WorkspaceRole `json:"role" db:"role"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" db:"updated_at"`
}

type CreateWorkspaceRequest struct {
	Name string `json:"name" binding:"required,min=2,max=150"`
}

type UpdateWorkspaceRequest struct {
	Name string `json:"name" binding:"required,min=2,max=150"`
}

type AddMemberRequest struct {
	UserID string `json:"user_id" binding:"required,uuid"`
	Role   string `json:"role" binding:"required"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

var (
	ErrInvalidWorkspaceInput   = errors.New("invalid workspace input")
	ErrWorkspaceNotFound       = errors.New("workspace not found")
	ErrWorkspaceMemberExists   = errors.New("user is already a member of this workspace")
	ErrWorkspaceMemberNotFound = errors.New("workspace member not found")
)

func NormalizeRole(role string) WorkspaceRole {
	switch WorkspaceRole(strings.ToUpper(strings.TrimSpace(role))) {
	case RoleAdmin:
		return RoleAdmin
	case RoleOwner:
		return RoleOwner
	default:
		return RoleMember
	}
}
