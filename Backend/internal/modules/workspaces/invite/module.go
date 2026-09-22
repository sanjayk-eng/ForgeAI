package invite

import (
	"context"

	"ai-agent/internal/shared/logger"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// EmailService interface for sending workspace invite emails
type EmailService interface {
	SendWorkspaceInvite(ctx context.Context, to, workspaceName, inviterName, inviteLink, role string) error
}

type ModuleConfig struct {
	ProtectedRouter gin.IRouter      // For authenticated routes
	PublicRouter    gin.IRouter      // For public routes  
	Database        *sqlx.DB
	Logger          logger.Logger
	EmailService    EmailService     // Interface, not concrete type
	FrontendURL     string
}

type Module struct {
	Service *Service
	Handler *Handler
}

func LoadModule(config ModuleConfig) *Module {
	service := NewService(config.Database, config.EmailService, config.FrontendURL, config.Logger)
	handler := NewHandler(service)
	
	// Register protected routes
	if config.ProtectedRouter != nil {
		RegisterRoutes(config.ProtectedRouter, handler)
	}
	
	// Register public routes
	if config.PublicRouter != nil {
		RegisterPublicRoutes(config.PublicRouter, handler)
	}

	return &Module{
		Service: service,
		Handler: handler,
	}
}
