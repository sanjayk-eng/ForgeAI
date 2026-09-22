package member

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

type Member struct {
	ID          string        `json:"id" db:"id"`
	WorkspaceID string        `json:"workspace_id" db:"workspace_id"`
	UserID      string        `json:"user_id" db:"user_id"`
	Role        WorkspaceRole `json:"role" db:"role"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" db:"updated_at"`
}

type AddMemberRequest struct {
	UserID string `json:"user_id" binding:"required,uuid"`
	Role   string `json:"role" binding:"required"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

var (
	ErrInvalidMemberInput  = errors.New("invalid member input")
	ErrMemberNotFound      = errors.New("workspace member not found")
	ErrMemberAlreadyExists = errors.New("user is already a member of this workspace")
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
