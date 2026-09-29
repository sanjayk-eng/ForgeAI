package application

import (
	"context"
	"time"

	"ai-agent/internal/modules/terminal/domain"
)

type Store interface {
	CanAccessWorkspace(ctx context.Context, workspaceID, userID string) (bool, error)
	FindProjectWorkspace(ctx context.Context, projectID string) (string, error)
	FindActiveByProject(ctx context.Context, projectID string) (domain.Sandbox, error)
	FindByID(ctx context.Context, sandboxID string) (domain.Sandbox, error)
	Create(ctx context.Context, sandbox domain.Sandbox) (string, error)
	AttachContainer(ctx context.Context, sandboxID, containerID string) error
	TransitionStatus(ctx context.Context, sandboxID string, from, to domain.SandboxStatus) error
	SetLastError(ctx context.Context, sandboxID, message string) error
}

type ContainerSpec struct {
	Name            string
	Image           string
	VolumeName      string
	WorkspacePath   string
	NetworkMode     string
	ReadOnlyRootFS  bool
	NoNewPrivileges bool
	CapDrop         []string
	Limits          domain.ResourceLimits
}

type ExecutionResult struct {
	Output   string
	ExitCode int
	Command  string
}

type FileEntry struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	IsDirectory bool   `json:"is_directory"`
	Size        int64  `json:"size,omitempty"`
	ModifiedAt  string `json:"modified_at,omitempty"`
}

type Runtime interface {
	CreateVolume(ctx context.Context, name string) error
	RemoveVolume(ctx context.Context, name string) error
	CreateContainer(ctx context.Context, spec ContainerSpec) (string, error)
	StartContainer(ctx context.Context, containerID string) error
	StopContainer(ctx context.Context, containerID string) error
	RemoveContainer(ctx context.Context, containerID string) error
	Execute(ctx context.Context, containerID, command string, timeoutSeconds int) (ExecutionResult, error)
}

type FileStore interface {
	ListFiles(ctx context.Context, containerID, path string) ([]FileEntry, error)
	ReadFile(ctx context.Context, containerID, path string) (string, error)
	WriteFile(ctx context.Context, containerID, path, content string) error
	CreateDirectory(ctx context.Context, containerID, path string) error
	DeletePath(ctx context.Context, containerID, path string) error
	RenamePath(ctx context.Context, containerID, oldPath, newPath string) error
}

type RepositoryCloner interface {
	CloneRepository(ctx context.Context, volumeName, workspacePath, helperImage, repositoryURL, branch, accessToken string, timeout time.Duration) error
}

type RepositoryPusher interface {
	PushRepository(ctx context.Context, volumeName, workspacePath, helperImage, remote, branch, accessToken string, timeout time.Duration) error
}
