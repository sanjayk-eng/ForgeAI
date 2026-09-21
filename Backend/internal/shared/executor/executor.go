package executor

import "context"

type CommandExecutor interface {
	Execute(ctx context.Context, command string) (Result, error)
}
