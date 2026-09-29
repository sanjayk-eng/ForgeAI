package docker

import (
	"context"
	"fmt"
	"strings"

	"ai-agent/internal/modules/terminal/application"
)

func (runtime *DockerRuntime) ListTerminalShells(ctx context.Context, containerID string) ([]application.TerminalShell, error) {
	output, err := runtime.docker.Run(ctx, "exec", containerID, "/bin/sh", "-lc", `for shell in sh bash zsh pwsh powershell; do printf '%s\t' "$shell"; command -v "$shell" 2>/dev/null || true; done`)
	if err != nil {
		return nil, fmt.Errorf("inspect shells in sandbox: %w", err)
	}
	return parseTerminalShells(output), nil
}

func parseTerminalShells(output string) []application.TerminalShell {
	paths := make(map[string]string)
	for line := range strings.SplitSeq(output, "\n") {
		name, executable, found := strings.Cut(strings.TrimSpace(line), "\t")
		if found && executable != "" {
			paths[name] = executable
		}
	}
	shells := make([]application.TerminalShell, 0, 4)
	for _, candidate := range []struct{ id, name, executable string }{
		{id: "sh", name: "Shell", executable: paths["sh"]},
		{id: "bash", name: "Bash", executable: paths["bash"]},
		{id: "zsh", name: "Zsh", executable: paths["zsh"]},
		{id: "powershell", name: "PowerShell", executable: firstNonEmpty(paths["pwsh"], paths["powershell"])},
	} {
		if candidate.executable == "" {
			continue
		}
		shells = append(shells, application.TerminalShell{
			ID: candidate.id, Name: candidate.name, Available: true,
			Default: candidate.id == "sh", Executable: candidate.executable,
		})
	}
	return shells
}

func (runtime *DockerRuntime) StartTerminal(ctx context.Context, containerID, workspacePath, shellID string, cols, rows int) (application.InteractiveProcess, error) {
	switch shellID {
	case "sh", "bash", "zsh", "powershell":
	default:
		return nil, fmt.Errorf("unsupported shell %q", shellID)
	}
	shells, err := runtime.ListTerminalShells(ctx, containerID)
	if err != nil {
		return nil, err
	}
	var selected *application.TerminalShell
	for index := range shells {
		if shells[index].ID == shellID {
			selected = &shells[index]
			break
		}
	}
	if selected == nil {
		return nil, fmt.Errorf("shell %q is unavailable", shellID)
	}
	if cols < 1 {
		cols = 120
	}
	if rows < 1 {
		rows = 30
	}
	if cols > 65535 {
		cols = 65535
	}
	if rows > 65535 {
		rows = 65535
	}

	args := []string{"exec", "--interactive", "--tty", "--workdir", workspacePath, containerID, selected.Executable}
	switch selected.ID {
	case "sh":
		args = append(args, "-i")
	case "bash":
		args = append(args, "--noprofile", "--norc", "-i")
	case "zsh":
		args = append(args, "-i")
	case "powershell":
		args = append(args, "-NoLogo", "-NoExit")
	}
	return startInteractiveProcess(runtime.docker.binary, args, cols, rows)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
