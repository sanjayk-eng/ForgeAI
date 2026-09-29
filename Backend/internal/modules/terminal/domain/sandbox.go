package domain

import (
	"errors"
	"slices"
	"time"
)

type SandboxStatus string

const (
	StatusCreating   SandboxStatus = "CREATING"
	StatusCreated    SandboxStatus = "CREATED"
	StatusStarting   SandboxStatus = "STARTING"
	StatusRunning    SandboxStatus = "RUNNING"
	StatusStopping   SandboxStatus = "STOPPING"
	StatusStopped    SandboxStatus = "STOPPED"
	StatusRestarting SandboxStatus = "RESTARTING"
	StatusFailed     SandboxStatus = "FAILED"
	StatusDestroying SandboxStatus = "DESTROYING"
	StatusDestroyed  SandboxStatus = "DESTROYED"
)

var (
	ErrInvalidTransition        = errors.New("invalid sandbox lifecycle transition")
	ErrSandboxNotFound          = errors.New("sandbox not found")
	ErrSandboxExists            = errors.New("project already has an active sandbox")
	ErrSandboxAccessDenied      = errors.New("sandbox workspace access denied")
	ErrSandboxStateChanged      = errors.New("sandbox state changed concurrently")
	ErrInvalidSandbox           = errors.New("invalid sandbox input")
	ErrSandboxNotRunning        = errors.New("sandbox is not running")
	ErrTerminalShellUnavailable = errors.New("terminal shell is unavailable")
)

type ResourceLimits struct {
	CPUShares   int64 `json:"cpu_shares"`
	MemoryBytes int64 `json:"memory_bytes"`
	PidsLimit   int64 `json:"pids_limit"`
}

type Sandbox struct {
	ID            string         `json:"id"`
	WorkspaceID   string         `json:"workspace_id"`
	ProjectID     string         `json:"project_id"`
	ContainerID   string         `json:"container_id"`
	ContainerName string         `json:"container_name"`
	VolumeName    string         `json:"volume_name"`
	Image         string         `json:"image"`
	ImageActual   string         `json:"image_actual"`
	WorkspacePath string         `json:"workspace_path"`
	Status        SandboxStatus  `json:"status"`
	LastError     string         `json:"last_error,omitempty"`
	Limits        ResourceLimits `json:"limits"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

func (sandbox Sandbox) Transition(next SandboxStatus) (Sandbox, error) {
	if !canTransition(sandbox.Status, next) {
		return Sandbox{}, ErrInvalidTransition
	}
	sandbox.Status = next
	return sandbox, nil
}

func canTransition(current, next SandboxStatus) bool {
	if current == next {
		return true
	}
	allowed := map[SandboxStatus][]SandboxStatus{
		StatusCreating:   {StatusCreated, StatusFailed},
		StatusCreated:    {StatusStarting, StatusDestroying, StatusFailed},
		StatusStarting:   {StatusRunning, StatusFailed},
		StatusRunning:    {StatusStopping, StatusRestarting, StatusDestroying, StatusFailed},
		StatusStopping:   {StatusStopped, StatusFailed},
		StatusStopped:    {StatusStarting, StatusDestroying, StatusFailed},
		StatusRestarting: {StatusRunning, StatusFailed},
		StatusFailed:     {StatusStarting, StatusDestroying, StatusCreating},
		StatusDestroying: {StatusDestroyed, StatusFailed},
		StatusDestroyed:  {},
	}
	return slices.Contains(allowed[current], next)
}
