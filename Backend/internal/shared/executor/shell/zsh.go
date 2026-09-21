package shell

import "ai-agent/internal/shared/executor"

func NewZsh() executor.CommandExecutor {
	return New("zsh", "-c")
}
