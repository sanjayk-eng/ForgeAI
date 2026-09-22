package member

import (
	"ai-agent/internal/shared/logger"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type ModuleConfig struct {
	Router   gin.IRouter
	Database *sqlx.DB
	Logger   logger.Logger
}

type Module struct {
	Repository Repository
	Service    *Service
	Handler    *Handler
}

func LoadModule(config ModuleConfig) *Module {
	repository := NewRepository(config.Database)
	service := NewService(config.Database, repository)
	handler := NewHandler(service)
	RegisterRoutes(config.Router, handler)

	return &Module{
		Repository: repository,
		Service:    service,
		Handler:    handler,
	}
}
