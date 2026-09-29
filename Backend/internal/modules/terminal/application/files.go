package application

import (
	"context"
	"fmt"
	"path"
	"time"

	"ai-agent/internal/modules/terminal/domain"
	"ai-agent/internal/shared/filesystem"
)

func (service *Service) runningSandbox(ctx context.Context, sandboxID, operation string) (domain.Sandbox, error) {
	sandbox, err := service.Get(ctx, sandboxID)
	if err != nil {
		return domain.Sandbox{}, err
	}
	if sandbox.Status != domain.StatusRunning {
		return domain.Sandbox{}, fmt.Errorf("sandbox must be RUNNING to %s", operation)
	}
	return sandbox, nil
}

func (service *Service) Execute(ctx context.Context, sandboxID, command string) (ExecutionResult, error) {
	sandbox, err := service.runningSandbox(ctx, sandboxID, "execute commands")
	if err != nil {
		return ExecutionResult{}, err
	}
	return service.runtime.Execute(ctx, sandbox.ContainerID, command, int(service.policy.CommandTimeout.Seconds()))
}

func (service *Service) ListFiles(ctx context.Context, sandboxID, filePath string) ([]FileEntry, error) {
	sandbox, err := service.runningSandbox(ctx, sandboxID, "list files")
	if err != nil {
		return nil, err
	}
	return service.files.ListFiles(ctx, sandbox.ContainerID, filePath)
}

func (service *Service) ReadFile(ctx context.Context, sandboxID, filePath string) (string, error) {
	sandbox, err := service.runningSandbox(ctx, sandboxID, "read files")
	if err != nil {
		return "", err
	}
	return service.files.ReadFile(ctx, sandbox.ContainerID, filePath)
}

func (service *Service) safeWorkspacePath(workspaceRoot, rawPath string) (string, error) {
	fullPath, err := filesystem.ResolveWorkspacePath(workspaceRoot, rawPath)
	if err != nil {
		return "", domain.ErrInvalidSandbox
	}
	return fullPath, nil
}

func (service *Service) WriteFile(ctx context.Context, sandboxID, relativePath, content string) error {
	sandbox, err := service.runningSandbox(ctx, sandboxID, "write files")
	if err != nil {
		return err
	}
	fullPath, err := service.safeWorkspacePath(path.Clean(sandbox.WorkspacePath), relativePath)
	if err != nil {
		return err
	}
	if len(content) > 512*1024 {
		return fmt.Errorf("file content exceeds the write limit")
	}
	return service.files.WriteFile(ctx, sandbox.ContainerID, fullPath, content)
}

func (service *Service) CreateDirectory(ctx context.Context, sandboxID, relativePath string) error {
	sandbox, err := service.runningSandbox(ctx, sandboxID, "create directories")
	if err != nil {
		return err
	}
	fullPath, err := service.safeWorkspacePath(path.Clean(sandbox.WorkspacePath), relativePath)
	if err != nil {
		return err
	}
	return service.files.CreateDirectory(ctx, sandbox.ContainerID, fullPath)
}

func (service *Service) DeletePath(ctx context.Context, sandboxID, relativePath string) error {
	sandbox, err := service.runningSandbox(ctx, sandboxID, "delete files")
	if err != nil {
		return err
	}
	fullPath, err := service.safeWorkspacePath(path.Clean(sandbox.WorkspacePath), relativePath)
	if err != nil {
		return err
	}
	return service.files.DeletePath(ctx, sandbox.ContainerID, fullPath)
}

func (service *Service) RenamePath(ctx context.Context, sandboxID, oldPath, newPath string) error {
	sandbox, err := service.runningSandbox(ctx, sandboxID, "rename files")
	if err != nil {
		return err
	}
	workspaceRoot := path.Clean(sandbox.WorkspacePath)
	oldFull, err := service.safeWorkspacePath(workspaceRoot, oldPath)
	if err != nil {
		return err
	}
	newFull, err := service.safeWorkspacePath(workspaceRoot, newPath)
	if err != nil {
		return err
	}
	return service.files.RenamePath(ctx, sandbox.ContainerID, oldFull, newFull)
}

func (service *Service) CloneRepository(ctx context.Context, sandboxID, repositoryURL, branch, accessToken string) error {
	sandbox, err := service.runningSandbox(ctx, sandboxID, "clone a repository")
	if err != nil {
		return err
	}
	return service.repositoryCloner.CloneRepository(ctx, sandbox.VolumeName, sandbox.WorkspacePath, service.policy.GitImage, repositoryURL, branch, accessToken, 5*time.Minute)
}
