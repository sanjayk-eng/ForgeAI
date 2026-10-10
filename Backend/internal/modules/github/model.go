package github

import "time"

const (
	StatusActive   = "ACTIVE"
	StatusInactive = "INACTIVE"
	StatusExpired  = "EXPIRED"
	StatusRevoked  = "REVOKED"
)

type Integration struct {
	ID                  string    `db:"id" json:"id"`
	OrganizationID      string    `db:"organization_id" json:"-"`
	ProviderCode        string    `db:"provider_code" json:"provider_code"`
	ExternalAccountID   string    `db:"external_account_id" json:"external_account_id"`
	ExternalAccountName string    `db:"external_account_name" json:"external_account_name"`
	ExternalAccountType string    `db:"external_account_type" json:"external_account_type"`
	InstallationID      string    `db:"installation_id" json:"installation_id"`
	Status              string    `db:"status" json:"status"`
	ConnectedAt         time.Time `db:"connected_at" json:"connected_at"`
}

type ExportRepository struct {
	ProjectID      string `db:"project_id" json:"project_id"`
	IntegrationID  string `db:"integration_id" json:"integration_id"`
	RepositoryID   string `db:"external_repository_id" json:"repository_id"`
	RepositoryName string `db:"repository_name" json:"repository_name"`
	FullName       string `db:"full_name" json:"full_name"`
	DefaultBranch  string `db:"default_branch" json:"default_branch"`
	RepositoryURL  string `db:"repository_url" json:"repository_url"`
	Private        bool   `db:"is_private" json:"private"`
}

type ExportContext struct {
	ProjectID      string `db:"project_id"`
	OrganizationID string `db:"organization_id"`
	ProjectSlug    string `db:"project_slug"`
	IntegrationID  string `db:"integration_id"`
	InstallationID string `db:"installation_id"`
	AccountName    string `db:"account_name"`
	AccountType    string `db:"account_type"`
	Existing       *ExportRepository
}

type Installation struct {
	ID          int64
	AppID       int64
	AccountID   int64
	AccountName string
	AccountType string
}

type State struct {
	OrganizationID string `db:"organization_id"`
	UserID         string `db:"user_id"`
}

type GitHubAppConfig struct {
	AppID         int64
	Slug          string
	PrivateKey    string
	WebhookSecret string
}

type installationWebhook struct {
	Action       string `json:"action"`
	Installation struct {
		ID    int64 `json:"id"`
		AppID int64 `json:"app_id"`
	} `json:"installation"`
}

type CreateExportRepositoryRequest struct {
	Private *bool `json:"private" binding:"required"`
}

type CreateExportRepositoryResponse struct {
	Repository ExportRepository `json:"repository"`
}
