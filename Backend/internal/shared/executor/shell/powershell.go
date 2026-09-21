package shell

import "ai-agent/internal/shared/executor"

func NewPowerShell() executor.CommandExecutor {
	return New("powershell.exe", "-NoProfile", "-Command")
}
