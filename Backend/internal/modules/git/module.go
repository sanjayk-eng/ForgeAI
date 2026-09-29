package git

import (
	"ai-agent/internal/shared/realtime"

	"github.com/gin-gonic/gin"
)

type ModuleConfig struct {
	Router          gin.IRouter
	WorkspaceRoot   string
	SandboxExecutor SandboxExecutor
	SandboxAccess   SandboxAccessValidator
	GitHubAccounts  GitHubAccountStore
	Events          realtime.Publisher
}

type Module struct {
	Service *Service
	Handler *Handler
}

func LoadModule(config ModuleConfig) *Module {
	service := NewService(config.WorkspaceRoot, config.SandboxExecutor)
	handler := NewHandler(service, config.SandboxAccess, config.GitHubAccounts, config.Events)
	RegisterRoutes(config.Router, handler)
	return &Module{Service: service, Handler: handler}
}
