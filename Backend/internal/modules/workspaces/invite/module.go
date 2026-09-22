package invite

import (
	"ai-agent/internal/shared/email"
	"ai-agent/internal/shared/logger"
	"ai-agent/internal/shared/worker"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type ModuleConfig struct {
	Router   gin.IRouter
	Database *sqlx.DB
	Logger   logger.Logger
	Sender   email.Sender
	Queue    worker.Worker
}

type Module struct {
	Service *Service
	Handler *Handler
}

func LoadModule(config ModuleConfig) *Module {
	service := NewService(config.Database, config.Sender, config.Queue)
	handler := NewHandler(service)
	RegisterRoutes(config.Router, handler)

	return &Module{
		Service: service,
		Handler: handler,
	}
}
