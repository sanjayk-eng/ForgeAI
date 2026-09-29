package project

import (
	"ai-agent/internal/modules/project/orchestrator"
	"ai-agent/internal/modules/project/repository"
)

// ProjectWithRepository is an alias for cleaner external usage
type ProjectWithRepository = orchestrator.ProjectWithRepository

// CreateProjectRequest combines project creation with optional repository
type CreateProjectRequest struct {
	Name        string                              `json:"name" binding:"required,min=1,max=150"`
	Description *string                             `json:"description"`
	Type        string                              `json:"type" binding:"required,oneof=REPOSITORY EMPTY"`
	Repository  *repository.ConnectRepositoryRequest `json:"repository,omitempty"`
}

// UpdateProjectRequest for updating project details
type UpdateProjectRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=150"`
	Description *string `json:"description"`
	Status      string  `json:"status,omitempty" binding:"omitempty,oneof=ACTIVE ARCHIVED"`
}

// ConnectRepositoryRequest is an alias for clarity
type ConnectRepositoryRequest repository.ConnectRepositoryRequest

// UpdateRepositoryBranchRequest for updating repository branch
type UpdateRepositoryBranchRequest repository.UpdateRepositoryBranchRequest

// ResolveRepositoryRequest for resolving repository URL
type ResolveRepositoryRequest repository.ResolveRepositoryRequest

// ImportGitHubRepositoriesRequest for importing multiple repos
type ImportGitHubRepositoriesRequest struct {
	Repositories []repository.ConnectRepositoryRequest `json:"repositories" binding:"required,min=1,max=100"`
}

// ImportGitHubRepositoriesResponse for import result
type ImportGitHubRepositoriesResponse struct {
	Projects []ProjectWithRepository `json:"projects"`
	Skipped  int                     `json:"skipped"`
}
