package core

import "time"

const (
	StatusActive   = "ACTIVE"
	StatusArchived = "ARCHIVED"
)

type Organization struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Slug        string                 `json:"slug" db:"slug"`
	Description string                 `json:"description" db:"description"`
	Settings    map[string]interface{} `json:"settings" db:"settings"`
	Status      string                 `json:"status" db:"status"`
	CreatedBy   string                 `json:"created_by" db:"created_by"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time             `json:"deleted_at,omitempty" db:"deleted_at"`
}

type CreateOrganizationRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=150"`
	Description string `json:"description" validate:"max=1000"`
}

type UpdateOrganizationRequest struct {
	Name        string `json:"name" validate:"omitempty,min=2,max=150"`
	Description string `json:"description" validate:"max=1000"`
	Status      string `json:"status" validate:"omitempty,oneof=ACTIVE ARCHIVED"`
}

func (Organization) TableName() string {
	return "tbl_organization"
}
