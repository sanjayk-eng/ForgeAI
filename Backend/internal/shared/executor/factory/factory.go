package factory

import (
	"fmt"

	"ai-agent/internal/shared/executor"
	"ai-agent/internal/shared/executor/shell"
	"ai-agent/internal/shared/executor/spec"
)

type Shell = spec.Shell

const (
	ShellPowerShell = spec.ShellPowerShell
	ShellCMD        = spec.ShellCMD
	ShellBash       = spec.ShellBash
	ShellZsh        = spec.ShellZsh
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


func Resolve(selected Shell) (executor.CommandExecutor, Shell, error) {
	if selected == "" {
		selected = executor.DetectShell()
	}

	created, err := Create(selected)
	if err != nil {
		return nil, "", err
	}
	return created, selected, nil
}
