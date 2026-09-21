package shell

import (
	"context"

	"ai-agent/internal/shared/executor"
	"ai-agent/internal/shared/executor/runtime"
)

type Executor struct {
	program string
	args    []string
}

func New(program string, args ...string) executor.CommandExecutor {
	return Executor{program: program, args: args}
}

func (e Executor) Execute(ctx context.Context, command string) (executor.Result, error) {
	return runtime.Run(ctx, e.program, e.args, command)
}
