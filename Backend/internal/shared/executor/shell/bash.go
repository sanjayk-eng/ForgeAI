package shell

import "ai-agent/internal/shared/executor"

func NewBash() executor.CommandExecutor {
	return New("bash", "-c")
}
