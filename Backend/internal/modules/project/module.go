package project

import (
	"ai-agent/internal/shared/logger"
	"context"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type ModuleConfig struct {
	Router       gin.IRouter
	Database     *sqlx.DB
	Logger       logger.Logger
	Organization OrganizationService
}

type Module struct {
	Service Service
	Handler *Handler
}

func LoadModule(config ModuleConfig) *Module {
	// Initialize repositories
	Repo := NewRepository(config.Database)

	// Initialize services
	Service := NewService(config.Database, Repo, config.Organization)

	// Initialize handler
	handler := NewHandler(Service)

	// Register routes
	RegisterRoutes(config.Router, handler)

	return &Module{
		Service: Service,
		Handler: handler,
	}
}

func (m *Module) SetProjectCreatedHook(hook func(context.Context, string, string)) {
	m.Handler.projectCreatedHook = hook
}

func (m *Module) SetProjectDeletedHook(hook func(context.Context, string)) {
	m.Handler.projectDeletedHook = hook
}
