package invite

import "time"

type InviteStatus string

const (
	InvitePending  InviteStatus = "PENDING"
	InviteAccepted InviteStatus = "ACCEPTED"
	InviteRejected InviteStatus = "REJECTED"
	InviteExpired  InviteStatus = "EXPIRED"
	InviteRevoked  InviteStatus = "REVOKED"
)

type Invite struct {
	ID            string       `json:"id" db:"id"`
	WorkspaceID   string       `json:"workspace_id" db:"workspace_id"`
	Email         string       `json:"email" db:"email"`
	Role          string       `json:"role" db:"role"`
	Status        InviteStatus `json:"status" db:"status"`
	InvitedBy     string       `json:"invited_by" db:"invited_by"`
	InvitedByUser MemberUser   `json:"invited_by_user" db:"-"`
	TokenHash     string       `json:"token_hash" db:"token_hash"`
	ExpiresAt     time.Time    `json:"expires_at" db:"expires_at"`
	CreatedAt     time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at" db:"updated_at"`
}

type MemberUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type CreateInviteRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required"`
}

type UpdateInviteStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=ACCEPTED REJECTED"`
}
