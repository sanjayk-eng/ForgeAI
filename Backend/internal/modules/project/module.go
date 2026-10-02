package project

import (
	"context"
	
	"ai-agent/internal/modules/project/core"
	"ai-agent/internal/shared/logger"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type OrganizationService interface {
	ValidateAccess(ctx context.Context, userID, organizationID string) error
}

type ModuleConfig struct {
	Router       gin.IRouter
	Database     *sqlx.DB
	Logger       logger.Logger
	Organization OrganizationService
}

type Module struct {
	CoreService core.Service
	Handler     *Handler
}

func LoadModule(config ModuleConfig) *Module {
	// Initialize repositories
	coreRepo := core.NewRepository(config.Database)

	// Initialize services
	coreService := core.NewService(config.Database, coreRepo, config.Organization)

	// Initialize handler
	handler := NewHandler(coreService)

	// Register routes
	RegisterRoutes(config.Router, handler)

	return &Module{
		CoreService: coreService,
		Handler:     handler,
	}
}
