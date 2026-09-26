package project

import (
	"context"
	"time"

	"ai-agent/internal/modules/auth/provider"
	"ai-agent/internal/modules/project/core"
	"ai-agent/internal/modules/project/github"
	"ai-agent/internal/modules/project/orchestrator"
	projectrepo "ai-agent/internal/modules/project/repository"
	"ai-agent/internal/modules/project/sync"
	"ai-agent/internal/shared/logger"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// GitHubClient is an alias for easier external usage
type GitHubClient = github.Client

type ModuleConfig struct {
	Router        gin.IRouter
	Database      *sqlx.DB
	Logger        logger.Logger
	GitHubClient  github.Client
	GitHubAccount orchestrator.GitHubAccountStore
	SyncContext   context.Context
	SyncInterval  time.Duration
}

type Module struct {
	CoreService    core.Service
	RepoService    projectrepo.Service
	SyncService    *sync.Service
	ProjectService Service
	Handler        *Handler
}

// NewGitHubClient creates a GitHub client from a provider
func NewGitHubClient(inspector provider.GitHubRepositoryInspector) github.Client {
	return github.NewClient(inspector)
}

func LoadModule(config ModuleConfig) *Module {
	// Initialize repositories
	coreRepo := core.NewRepository(config.Database)
	repoRepo := projectrepo.NewRepository(config.Database)
	syncRepo := sync.NewRepository(config.Database)

	// Initialize services
	coreService := core.NewService(config.Database, coreRepo)
	repoService := projectrepo.NewService(config.Database, repoRepo)
	syncService := sync.NewService(syncRepo, config.GitHubClient, config.Logger, config.SyncInterval)

	// Initialize GitHub catalog service
	var catalogSvc github.CatalogService
	if config.GitHubClient != nil {
		catalogSvc = github.NewCatalogService(config.GitHubClient)
	}

	// Initialize orchestrators
	projectOrch := orchestrator.NewProjectOrchestrator(config.Database, coreService, repoService)
	
	githubOrch := orchestrator.NewGitHubOrchestrator(
		config.Database,
		coreRepo,
		repoRepo,
		config.GitHubClient,
		catalogSvc,
		config.GitHubAccount,
	)
	
	syncOrch := orchestrator.NewSyncOrchestrator(
		syncService,
		projectOrch,
		coreService,
		config.GitHubAccount,
	)

	// Initialize main service facade
	projectService := NewService(
		config.Database,
		coreService,
		repoService,
		projectOrch,
		githubOrch,
		syncOrch,
	)

	// Initialize handler
	handler := NewHandler(projectService)

	// Register routes
	RegisterRoutes(config.Router, handler)

	// Start sync service
	if config.Database != nil && syncService != nil {
		syncContext := config.SyncContext
		if syncContext == nil {
			syncContext = context.Background()
		}
		go syncService.Run(syncContext)
	}

	return &Module{
		CoreService:    coreService,
		RepoService:    repoService,
		SyncService:    syncService,
		ProjectService: projectService,
		Handler:        handler,
	}
}
