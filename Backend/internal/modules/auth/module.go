package auth

import (
	"ai-agent/internal/shared/logger"
	appjwt "ai-agent/pkg/jwt"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type ModuleConfig struct {
	Router   gin.IRouter
	Database *sqlx.DB
	Provider ProviderFactory
	JWT      *appjwt.Manager
	Logger   logger.Logger
}

type Module struct {
	Repository AuthRepository
	Service    *Service
	Handler    *Handler
}

func LoadModule(config ModuleConfig) *Module {
	var repository AuthRepository
	if config.Database != nil {
		repository = NewRepository(config.Database)
	}

	service := NewService(config.Provider, config.Database, repository, config.JWT, config.Logger)
	handler := NewHandler(service)
	RegisterRoutes(config.Router, handler)
	RegisterProtectedRoutes(config.Router, handler, config.JWT, config.Logger)

	return &Module{
		Repository: repository,
		Service:    service,
		Handler:    handler,
	}
}
