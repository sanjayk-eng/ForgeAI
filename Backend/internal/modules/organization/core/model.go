package core

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

const (
	StatusActive   = "ACTIVE"
	StatusArchived = "ARCHIVED"
)

type JSONMap map[string]interface{}

func (settings *JSONMap) Scan(source interface{}) error {
	if settings == nil {
		return fmt.Errorf("scan organization settings into nil receiver")
	}

	var value []byte
	switch source := source.(type) {
	case nil:
		*settings = JSONMap{}
		return nil
	case []byte:
		value = source
	case string:
		value = []byte(source)
	default:
		return fmt.Errorf("scan organization settings from %T", source)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(value, &decoded); err != nil {
		return fmt.Errorf("decode organization settings: %w", err)
	}
	if decoded == nil {
		decoded = make(map[string]interface{})
	}
	*settings = JSONMap(decoded)
	return nil
}

func (settings JSONMap) Value() (driver.Value, error) {
	if settings == nil {
		return `{}`, nil
	}
	value, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}
	return string(value), nil
}

type Organization struct {
	ID          string     `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Slug        string     `json:"slug" db:"slug"`
	Description string     `json:"description" db:"description"`
	Settings    JSONMap    `json:"settings" db:"settings"`
	Status      string     `json:"status" db:"status"`
	CreatedBy   string     `json:"created_by" db:"created_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
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
