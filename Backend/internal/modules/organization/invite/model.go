package invite

import "time"

const (
	RoleAdmin      = "ADMIN"
	RoleMember     = "MEMBER"
	StatusPending  = "PENDING"
	
	StatusAccepted = "ACCEPTED"
	StatusRejected = "REJECTED"
	StatusExpired  = "EXPIRED"
	StatusRevoked  = "REVOKED"
)

type Invite struct {
	ID               string    `json:"id" db:"id"`
	OrganizationID   string    `json:"organization_id" db:"organization_id"`
	OrganizationName string    `json:"organization_name,omitempty" db:"organization_name"`
	Email            string    `json:"email" db:"email"`
	Role             string    `json:"role" db:"role"`
	Status           string    `json:"status" db:"status"`
	InvitedBy        string    `json:"invited_by" db:"invited_by"`
	InvitedByUser    *Inviter  `json:"invited_by_user,omitempty" db:"-"`
	InviterName      string    `json:"-" db:"inviter_name"`
	InviterEmail     string    `json:"-" db:"inviter_email"`
	ExpiresAt        time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

type Inviter struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CreateRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (invite *Invite) setInviter() {
	if invite.InviterName == "" && invite.InviterEmail == "" {
		return
	}
	invite.InvitedByUser = &Inviter{ID: invite.InvitedBy, Name: invite.InviterName, Email: invite.InviterEmail}
}
