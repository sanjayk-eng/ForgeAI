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
	ErrInvalidTransition   = errors.New("invalid sandbox lifecycle transition")
	ErrSandboxNotFound     = errors.New("sandbox not found")
	ErrSandboxExists       = errors.New("project already has an active sandbox")
	ErrSandboxStateChanged = errors.New("sandbox state changed concurrently")
	ErrInvalidSandbox      = errors.New("invalid sandbox input")
)

type ResourceLimits struct {
	CPUShares   int64 `json:"cpu_shares"`
	MemoryBytes int64 `json:"memory_bytes"`
	PidsLimit   int64 `json:"pids_limit"`
}

type Sandbox struct {
	ID            string
	WorkspaceID   string
	ProjectID     string
	ContainerID   string
	ContainerName string
	VolumeName    string
	Image         string
	ImageActual   string
	WorkspacePath string
	Status        SandboxStatus
	LastError     string
	Limits        ResourceLimits
	CreatedAt     time.Time
	UpdatedAt     time.Time
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
