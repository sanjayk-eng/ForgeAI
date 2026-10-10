package github

import (
	"ai-agent/internal/modules/github/types"
)

// Re-export commonly used types for convenience
type Integration = types.Integration
type ExportRepository = types.ExportRepository
type ExportContext = types.ExportContext
type Installation = types.Installation
type State = types.State
type GitHubAppConfig = types.GitHubAppConfig

// Re-export constants
const (
	StatusActive   = types.StatusActive
	StatusInactive = types.StatusInactive
	StatusExpired  = types.StatusExpired
	StatusRevoked  = types.StatusRevoked
)

// Request/Response DTOs
type installationWebhook struct {
	Action       string `json:"action"`
	Installation struct {
		ID    int64 `json:"id"`
		AppID int64 `json:"app_id"`
	} `json:"installation"`
}

type CreateExportRepositoryRequest struct {
	Private *bool `json:"private" binding:"required" validate:"required"`
}

type CreateExportRepositoryResponse struct {
	Repository ExportRepository `json:"repository"`
}
