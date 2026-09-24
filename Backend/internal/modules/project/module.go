package project

import (
	"context"
	"time"

	"ai-agent/internal/shared/logger"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type ModuleConfig struct {
	Router        gin.IRouter
	Database      *sqlx.DB
	Logger        logger.Logger
	GitHubClient  GitHubRepositoryClient
	GitHubAccount GitHubAccountStore
	SyncContext   context.Context
	SyncInterval  time.Duration
}

type Module struct {
	Repository ProjectRepositoryStore
	Service    Service
	Handler    *Handler
	Sync       *SyncService
}

func LoadModule(config ModuleConfig) *Module {
	repository := NewRepository(config.Database)
	syncService := NewSyncService(repository, config.GitHubClient, config.Logger, config.SyncInterval)
	service := NewServiceWithSync(config.Database, repository, syncService, config.GitHubClient, config.GitHubAccount)
	handler := NewHandler(service)
	RegisterRoutes(config.Router, handler)
	if config.Database != nil {
		syncContext := config.SyncContext
		if syncContext == nil {
			syncContext = context.Background()
		}
		go syncService.Run(syncContext)
	}

	return &Module{Repository: repository, Service: service, Handler: handler, Sync: syncService}
}
