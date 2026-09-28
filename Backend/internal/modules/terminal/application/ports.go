package application

import (
	"context"
	"time"

	"ai-agent/internal/modules/terminal/domain"
)

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
}

type RepositoryCloner interface {
	CloneRepository(ctx context.Context, volumeName, workspacePath, helperImage, repositoryURL, branch, accessToken string, timeout time.Duration) error
}
