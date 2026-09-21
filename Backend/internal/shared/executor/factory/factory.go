package factory

import (
	"fmt"
	"os"
	"runtime"

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

func DetectShell() Shell {
	if runtime.GOOS == "windows" {
		if os.Getenv("COMSPEC") != "" {
			return ShellCMD
		}
		return ShellPowerShell
	}

	if os.Getenv("SHELL") == "/bin/zsh" {
		return ShellZsh
	}
	return ShellBash
}

func Resolve(selected Shell) (executor.CommandExecutor, Shell, error) {
	if selected == "" {
		selected = DetectShell()
	}

	created, err := Create(selected)
	if err != nil {
		return nil, "", err
	}
	return created, selected, nil
}
