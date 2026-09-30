package application

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
)

var ErrDefinitionNotFound = errors.New("Go definition was not found in the workspace")

type GoDefinition struct {
	Path   string `json:"path"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

type GoDefinitionResolver interface {
	ResolveGoDefinition(ctx context.Context, containerID, workspacePath, filePath string, line, column int) (GoDefinition, error)
}

func (service *Service) GoDefinition(ctx context.Context, sandboxID, filePath string, line, column int) (GoDefinition, error) {
	sandbox, err := service.runningSandbox(ctx, sandboxID, "resolve Go definitions")
	if err != nil {
		return GoDefinition{}, err
	}
	if line < 1 || column < 1 {
		return GoDefinition{}, fmt.Errorf("line and column must be positive")
	}
	fullPath, err := service.safeWorkspacePath(path.Clean(sandbox.WorkspacePath), filePath)
	if err != nil {
		return GoDefinition{}, err
	}
	if !strings.EqualFold(path.Ext(fullPath), ".go") {
		return GoDefinition{}, fmt.Errorf("Go definitions are only available for .go files")
	}
	resolver, ok := service.runtime.(GoDefinitionResolver)
	if !ok {
		return GoDefinition{}, fmt.Errorf("Go definition resolver is unavailable")
	}
	return resolver.ResolveGoDefinition(ctx, sandbox.ContainerID, sandbox.WorkspacePath, fullPath, line, column)
}
