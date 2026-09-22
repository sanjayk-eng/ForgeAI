package invite

import (
	"ai-agent/internal/shared/logger"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type ModuleConfig struct {
	Router       gin.IRouter
	Database     *sqlx.DB
	Logger       logger.Logger
	EmailService EmailService
	FrontendURL  string
}

type Module struct {
	Service *Service
	Handler *Handler
}

func LoadModule(config ModuleConfig) *Module {
	service := NewService(config.Database, config.EmailService, config.FrontendURL, config.Logger)
	handler := NewHandler(service)
	RegisterRoutes(config.Router, handler)

	return &Module{
		Service: service,
		Handler: handler,
	}
}
