package application

import (
	"context"
	"fmt"
	"strings"

	"ai-agent/internal/modules/terminal/domain"
)

func (service *Service) TerminalShells(ctx context.Context, userID, projectID string) ([]TerminalShell, error) {
	sandbox, err := service.terminalSandbox(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}
	return service.runtime.ListTerminalShells(ctx, sandbox.ContainerID)
}

func (service *Service) StartTerminal(ctx context.Context, userID, projectID, shellID string, cols, rows int) (domain.Sandbox, InteractiveProcess, error) {
	sandbox, err := service.terminalSandbox(ctx, userID, projectID)
	if err != nil {
		return domain.Sandbox{}, nil, err
	}
	if !supportedTerminalShell(shellID) {
		return domain.Sandbox{}, nil, domain.ErrTerminalShellUnavailable
	}
	process, err := service.runtime.StartTerminal(ctx, sandbox.ContainerID, sandbox.WorkspacePath, shellID, cols, rows)
	if err != nil {
		return domain.Sandbox{}, nil, fmt.Errorf("start terminal shell: %w", err)
	}
	return sandbox, process, nil
}

func supportedTerminalShell(shellID string) bool {
	switch shellID {
	case "sh", "bash", "zsh", "powershell":
		return true
	default:
		return false
	}
}

func (service *Service) terminalSandbox(ctx context.Context, userID, projectID string) (domain.Sandbox, error) {
	if service.runtime == nil || service.store == nil {
		return domain.Sandbox{}, domain.ErrInvalidSandbox
	}
	organizationID, err := service.workspaceForUser(ctx, userID, projectID)
	if err != nil {
		return domain.Sandbox{}, err
	}
	sandbox, err := service.store.FindActiveByProject(ctx, strings.TrimSpace(projectID))
	if err != nil {
		return domain.Sandbox{}, err
	}
	if sandbox.ProjectID != strings.TrimSpace(projectID) || sandbox.organizationID != organizationID {
		return domain.Sandbox{}, domain.ErrSandboxAccessDenied
	}
	if sandbox.Status != domain.StatusRunning || sandbox.ContainerID == "" {
		return domain.Sandbox{}, domain.ErrSandboxNotRunning
	}
	return sandbox, nil
}
