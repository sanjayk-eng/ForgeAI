package member

import "time"

const (
	RoleOwner  = "OWNER"
	RoleAdmin  = "ADMIN"
	RoleMember = "MEMBER"
)

type Member struct {
	ID             string     `json:"id" db:"id"`
	OrganizationID string     `json:"organization_id" db:"organization_id"`
	UserID         string     `json:"user_id" db:"user_id"`
	Role           string     `json:"role" db:"role"`
	JoinedAt       time.Time  `json:"joined_at" db:"joined_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

type MemberWithUser struct {
	Member
	User      UserProfile `json:"user" db:"-"`
	UserName  string      `json:"-" db:"user_name"`
	UserEmail string      `json:"-" db:"user_email"`
}

type UserProfile struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UpdateMemberRequest struct {
	Role string `json:"role" validate:"required,oneof=ADMIN MEMBER"`
}

func (Member) TableName() string {
	return "tbl_organization_member"
}
