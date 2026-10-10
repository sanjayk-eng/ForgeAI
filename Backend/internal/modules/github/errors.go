package github

import (
	"errors"

	"ai-agent/internal/modules/github/service"
)

// Re-export common errors
var ErrGitHubAppNotConfigured = service.ErrGitHubAppNotConfigured
var ErrStateInvalid = errors.New("installation state is invalid or expired")
var ErrIntegrationNotFound = errors.New("GitHub integration not found")
var ErrProjectNotFound = errors.New("project not found")
var ErrExportRepositoryNotFound = errors.New("GitHub export repository not found")
