package runtime

import (
	"ai-agent/internal/shared/executor"
	"context"
	"os/exec"
)

func Run(ctx context.Context, program string, args []string, command string) (executor.Result, error) {
	commandArgs := append([]string{}, args...)
	commandArgs = append(commandArgs, command)

	cmd := exec.CommandContext(ctx, program, commandArgs...)
	output, err := cmd.CombinedOutput()

	result := executor.Result{
		Output:   string(output),
		ExitCode: -1,
		Command:  command,
	}
	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}

	return result, err
}
