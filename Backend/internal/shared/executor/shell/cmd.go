package shell

import "ai-agent/internal/shared/executor"

func NewCMD() executor.CommandExecutor {
	return New("cmd.exe", "/C")
}
