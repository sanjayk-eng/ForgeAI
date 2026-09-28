package application

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"ai-agent/internal/modules/terminal/domain"
	"ai-agent/internal/modules/terminal/policy"
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

type Service struct {
	store            Store
	runtime          Runtime
	files            FileStore
	repositoryCloner RepositoryCloner
	policy           policy.Sandbox
}

func NewService(store Store, runtime Runtime, files FileStore, repositoryCloner RepositoryCloner, sandboxPolicy policy.Sandbox) *Service {
	return &Service{
		store:            store,
		runtime:          runtime,
		files:            files,
		repositoryCloner: repositoryCloner,
		policy:           sandboxPolicy,
	}
}

func (service *Service) Create(ctx context.Context, userID, projectID string) (domain.Sandbox, error) {
	userID = strings.TrimSpace(userID)
	projectID = strings.TrimSpace(projectID)
	if userID == "" || projectID == "" || service.store == nil || service.runtime == nil {
		return domain.Sandbox{}, domain.ErrInvalidSandbox
	}
	workspaceID, err := service.workspaceForUser(ctx, userID, projectID)
	if err != nil {
		return domain.Sandbox{}, err
	}
	if active, err := service.store.FindActiveByProject(ctx, projectID); err == nil && active.ID != "" {
		return domain.Sandbox{}, domain.ErrSandboxExists
	} else if err != nil && !errors.Is(err, domain.ErrSandboxNotFound) {
		return domain.Sandbox{}, fmt.Errorf("check active sandbox: %w", err)
	}

	containerName, volumeName, err := domain.ResourceIdentity(projectID)
	if err != nil {
		return domain.Sandbox{}, err
	}
	sandbox := domain.Sandbox{
		WorkspaceID: workspaceID, ProjectID: projectID, ContainerName: containerName,
		VolumeName: volumeName, Image: service.policy.Image, ImageActual: service.policy.Image,
		WorkspacePath: service.policy.WorkspacePath, Status: domain.StatusCreating,
		Limits: service.policy.ResourceLimits(),
	}
	sandbox.ID, err = service.store.Create(ctx, sandbox)
	if err != nil {
		return domain.Sandbox{}, fmt.Errorf("persist sandbox creating state: %w", err)
	}
	fail := func(cause error) (domain.Sandbox, error) {
		_ = service.store.TransitionStatus(ctx, sandbox.ID, domain.StatusCreating, domain.StatusFailed)
		_ = service.store.SetLastError(ctx, sandbox.ID, cause.Error())
		return domain.Sandbox{}, cause
	}
	if err := service.runtime.CreateVolume(ctx, volumeName); err != nil {
		return fail(fmt.Errorf("create sandbox volume: %w", err))
	}
	containerID, err := service.runtime.CreateContainer(ctx, ContainerSpec{
		Name: containerName, Image: service.policy.Image, VolumeName: volumeName,
		WorkspacePath: service.policy.WorkspacePath, NetworkMode: service.policy.NetworkMode,
		ReadOnlyRootFS: service.policy.ReadOnlyRootFS, NoNewPrivileges: service.policy.NoNewPrivileges,
		CapDrop: service.policy.CapDrop, Limits: service.policy.ResourceLimits(),
	})
	if err != nil {
		_ = service.runtime.RemoveVolume(ctx, volumeName)
		return fail(fmt.Errorf("create sandbox container: %w", err))
	}
	if err := service.store.AttachContainer(ctx, sandbox.ID, containerID); err != nil {
		_ = service.runtime.RemoveContainer(ctx, containerID)
		_ = service.runtime.RemoveVolume(ctx, volumeName)
		return fail(fmt.Errorf("persist sandbox container: %w", err))
	}
	if err := service.store.TransitionStatus(ctx, sandbox.ID, domain.StatusCreating, domain.StatusCreated); err != nil {
		_ = service.runtime.RemoveContainer(ctx, containerID)
		_ = service.runtime.RemoveVolume(ctx, volumeName)
		return fail(fmt.Errorf("mark sandbox created: %w", err))
	}
	sandbox.ContainerID = containerID
	sandbox.Status = domain.StatusCreated
	sandbox.CreatedAt = time.Now()
	sandbox.UpdatedAt = sandbox.CreatedAt
	return sandbox, nil
}

func (service *Service) Get(ctx context.Context, sandboxID string) (domain.Sandbox, error) {
	if strings.TrimSpace(sandboxID) == "" {
		return domain.Sandbox{}, domain.ErrInvalidSandbox
	}
	return service.store.FindByID(ctx, sandboxID)
}

func (service *Service) Start(ctx context.Context, sandboxID string) (domain.Sandbox, error) {
	sandbox, err := service.Get(ctx, sandboxID)
	if err != nil {
		return domain.Sandbox{}, err
	}
	if _, err := sandbox.Transition(domain.StatusStarting); err != nil {
		return domain.Sandbox{}, err
	}
	if err := service.store.TransitionStatus(ctx, sandbox.ID, domain.StatusCreated, domain.StatusStarting); err != nil {
		return domain.Sandbox{}, err
	}
	if err := service.runtime.StartContainer(ctx, sandbox.ContainerID); err != nil {
		_ = service.store.TransitionStatus(ctx, sandbox.ID, domain.StatusStarting, domain.StatusFailed)
		_ = service.store.SetLastError(ctx, sandbox.ID, fmt.Sprintf("start sandbox: %v", err))
		return domain.Sandbox{}, fmt.Errorf("start sandbox: %w", err)
	}
	if err := service.store.TransitionStatus(ctx, sandbox.ID, domain.StatusStarting, domain.StatusRunning); err != nil {
		return domain.Sandbox{}, err
	}
	sandbox.Status = domain.StatusRunning
	return sandbox, nil
}

func (service *Service) Stop(ctx context.Context, sandboxID string) (domain.Sandbox, error) {
	sandbox, err := service.Get(ctx, sandboxID)
	if err != nil {
		return domain.Sandbox{}, err
	}
	if _, err := sandbox.Transition(domain.StatusStopping); err != nil {
		return domain.Sandbox{}, err
	}
	if err := service.store.TransitionStatus(ctx, sandbox.ID, domain.StatusRunning, domain.StatusStopping); err != nil {
		return domain.Sandbox{}, err
	}
	if err := service.runtime.StopContainer(ctx, sandbox.ContainerID); err != nil {
		_ = service.store.TransitionStatus(ctx, sandbox.ID, domain.StatusStopping, domain.StatusFailed)
		_ = service.store.SetLastError(ctx, sandbox.ID, fmt.Sprintf("stop sandbox: %v", err))
		return domain.Sandbox{}, fmt.Errorf("stop sandbox: %w", err)
	}
	if err := service.store.TransitionStatus(ctx, sandbox.ID, domain.StatusStopping, domain.StatusStopped); err != nil {
		return domain.Sandbox{}, err
	}
	sandbox.Status = domain.StatusStopped
	return sandbox, nil
}

func (service *Service) Restart(ctx context.Context, sandboxID string) (domain.Sandbox, error) {
	sandbox, err := service.Get(ctx, sandboxID)
	if err != nil {
		return domain.Sandbox{}, err
	}
	if _, err := sandbox.Transition(domain.StatusRestarting); err != nil {
		return domain.Sandbox{}, err
	}
	if err := service.store.TransitionStatus(ctx, sandbox.ID, domain.StatusRunning, domain.StatusRestarting); err != nil {
		return domain.Sandbox{}, err
	}
	if err := service.runtime.StopContainer(ctx, sandbox.ContainerID); err != nil {
		_ = service.store.TransitionStatus(ctx, sandbox.ID, domain.StatusRestarting, domain.StatusFailed)
		_ = service.store.SetLastError(ctx, sandbox.ID, fmt.Sprintf("restart sandbox stop: %v", err))
		return domain.Sandbox{}, fmt.Errorf("restart sandbox stop: %w", err)
	}
	if err := service.runtime.StartContainer(ctx, sandbox.ContainerID); err != nil {
		_ = service.store.TransitionStatus(ctx, sandbox.ID, domain.StatusRestarting, domain.StatusFailed)
		_ = service.store.SetLastError(ctx, sandbox.ID, fmt.Sprintf("restart sandbox start: %v", err))
		return domain.Sandbox{}, fmt.Errorf("restart sandbox start: %w", err)
	}
	if err := service.store.TransitionStatus(ctx, sandbox.ID, domain.StatusRestarting, domain.StatusRunning); err != nil {
		return domain.Sandbox{}, err
	}
	sandbox.Status = domain.StatusRunning
	return sandbox, nil
}

func (service *Service) Destroy(ctx context.Context, sandboxID string) (domain.Sandbox, error) {
	sandbox, err := service.Get(ctx, sandboxID)
	if err != nil {
		return domain.Sandbox{}, err
	}
	if _, err := sandbox.Transition(domain.StatusDestroying); err != nil {
		return domain.Sandbox{}, err
	}
	if err := service.store.TransitionStatus(ctx, sandbox.ID, sandbox.Status, domain.StatusDestroying); err != nil {
		return domain.Sandbox{}, err
	}
	containerID := sandbox.ContainerID
	if containerID == "" {
		containerID = sandbox.ContainerName
	}
	if containerID != "" {
		if err := service.runtime.RemoveContainer(ctx, containerID); err != nil {
			_ = service.store.TransitionStatus(ctx, sandbox.ID, domain.StatusDestroying, domain.StatusFailed)
			_ = service.store.SetLastError(ctx, sandbox.ID, fmt.Sprintf("destroy sandbox container: %v", err))
			return domain.Sandbox{}, fmt.Errorf("destroy sandbox container: %w", err)
		}
	}
	if err := service.runtime.RemoveVolume(ctx, sandbox.VolumeName); err != nil {
		_ = service.store.TransitionStatus(ctx, sandbox.ID, domain.StatusDestroying, domain.StatusFailed)
		_ = service.store.SetLastError(ctx, sandbox.ID, fmt.Sprintf("destroy sandbox volume: %v", err))
		return domain.Sandbox{}, fmt.Errorf("destroy sandbox volume: %w", err)
	}
	if err := service.store.TransitionStatus(ctx, sandbox.ID, domain.StatusDestroying, domain.StatusDestroyed); err != nil {
		return domain.Sandbox{}, err
	}
	sandbox.Status = domain.StatusDestroyed
	return sandbox, nil
}

func (service *Service) GetByProject(ctx context.Context, projectID string) (domain.Sandbox, error) {
	if strings.TrimSpace(projectID) == "" {
		return domain.Sandbox{}, domain.ErrInvalidSandbox
	}
	return service.store.FindActiveByProject(ctx, projectID)
}

func (service *Service) ValidateProjectAccess(ctx context.Context, userID, projectID string) error {
	_, err := service.workspaceForUser(ctx, userID, projectID)
	return err
}

func (service *Service) ValidateSandboxAccess(ctx context.Context, userID, sandboxID string) error {
	sandbox, err := service.Get(ctx, sandboxID)
	if err != nil {
		return err
	}
	allowed, err := service.store.CanAccessWorkspace(ctx, sandbox.WorkspaceID, strings.TrimSpace(userID))
	if err != nil {
		return fmt.Errorf("check sandbox workspace access: %w", err)
	}
	if !allowed {
		return domain.ErrSandboxAccessDenied
	}
	return nil
}

func (service *Service) workspaceForUser(ctx context.Context, userID, projectID string) (string, error) {
	userID = strings.TrimSpace(userID)
	projectID = strings.TrimSpace(projectID)
	if userID == "" || projectID == "" || service.store == nil {
		return "", domain.ErrInvalidSandbox
	}
	workspaceID, err := service.store.FindProjectWorkspace(ctx, projectID)
	if err != nil {
		return "", err
	}
	allowed, err := service.store.CanAccessWorkspace(ctx, workspaceID, userID)
	if err != nil {
		return "", fmt.Errorf("check sandbox workspace access: %w", err)
	}
	if !allowed {
		return "", domain.ErrSandboxAccessDenied
	}
	return workspaceID, nil
}

func (service *Service) Execute(ctx context.Context, sandboxID, command string) (ExecutionResult, error) {
	sandbox, err := service.Get(ctx, sandboxID)
	if err != nil {
		return ExecutionResult{}, err
	}
	if sandbox.Status != domain.StatusRunning {
		return ExecutionResult{}, fmt.Errorf("sandbox must be RUNNING to execute commands")
	}
	return service.runtime.Execute(ctx, sandbox.ContainerID, command, int(service.policy.CommandTimeout.Seconds()))
}

func (service *Service) ListFiles(ctx context.Context, sandboxID, path string) ([]FileEntry, error) {
	sandbox, err := service.Get(ctx, sandboxID)
	if err != nil {
		return nil, err
	}
	if sandbox.Status != domain.StatusRunning {
		return nil, fmt.Errorf("sandbox must be RUNNING to list files")
	}
	return service.files.ListFiles(ctx, sandbox.ContainerID, path)
}

func (service *Service) ReadFile(ctx context.Context, sandboxID, path string) (string, error) {
	sandbox, err := service.Get(ctx, sandboxID)
	if err != nil {
		return "", err
	}
	if sandbox.Status != domain.StatusRunning {
		return "", fmt.Errorf("sandbox must be RUNNING to read files")
	}
	return service.files.ReadFile(ctx, sandbox.ContainerID, path)
}

func (service *Service) WriteFile(ctx context.Context, sandboxID, relativePath, content string) error {
	sandbox, err := service.Get(ctx, sandboxID)
	if err != nil {
		return err
	}
	if sandbox.Status != domain.StatusRunning {
		return fmt.Errorf("sandbox must be RUNNING to write files")
	}
	if strings.Contains(relativePath, "\\") || strings.HasPrefix(relativePath, "/") {
		return domain.ErrInvalidSandbox
	}
	cleanPath := path.Clean(relativePath)
	if cleanPath == "." || cleanPath == ".." || strings.HasPrefix(cleanPath, "../") {
		return domain.ErrInvalidSandbox
	}
	for _, segment := range strings.Split(cleanPath, "/") {
		if strings.EqualFold(segment, ".git") {
			return domain.ErrInvalidSandbox
		}
	}
	workspaceRoot := path.Clean(sandbox.WorkspacePath)
	fullPath := path.Join(workspaceRoot, cleanPath)
	if !strings.HasPrefix(fullPath, workspaceRoot+"/") {
		return domain.ErrInvalidSandbox
	}
	if len(content) > 512*1024 {
		return fmt.Errorf("file content exceeds the write limit")
	}
	return service.files.WriteFile(ctx, sandbox.ContainerID, fullPath, content)
}

func (service *Service) CloneRepository(ctx context.Context, sandboxID, repositoryURL, branch, accessToken string) error {
	sandbox, err := service.Get(ctx, sandboxID)
	if err != nil {
		return err
	}
	if sandbox.Status != domain.StatusRunning {
		return fmt.Errorf("sandbox must be RUNNING to clone a repository")
	}
	return service.repositoryCloner.CloneRepository(ctx, sandbox.VolumeName, sandbox.WorkspacePath, service.policy.GitImage, repositoryURL, branch, accessToken, 5*time.Minute)
}
