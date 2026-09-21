package factory

import (
	"fmt"

	"ai-agent/internal/shared/executor"
	"ai-agent/internal/shared/executor/shell"
)

type Shell string

const (
	ShellPowerShell Shell = "powershell"
	ShellCMD        Shell = "cmd"
	ShellBash       Shell = "bash"
	ShellZsh        Shell = "zsh"
)

func Create(selected Shell) (executor.CommandExecutor, error) {
	switch selected {
	case ShellPowerShell:
		return shell.NewPowerShell(), nil
	case ShellCMD:
		return shell.NewCMD(), nil
	case ShellBash:
		return shell.NewBash(), nil
	case ShellZsh:
		return shell.NewZsh(), nil
	default:
		return nil, fmt.Errorf("unsupported shell: %q", selected)
	}
}
