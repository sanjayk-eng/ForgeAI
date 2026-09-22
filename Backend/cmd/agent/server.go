package main

import (
	"ai-agent/internal/config"
	"ai-agent/internal/middleware"
	"ai-agent/internal/modules/auth"
	"ai-agent/internal/modules/auth/provider"
	workspaceinvite "ai-agent/internal/modules/workspaces/invite"
	member "ai-agent/internal/modules/workspaces/member"
	workspacecore "ai-agent/internal/modules/workspaces/workspace"
	"ai-agent/internal/shared/email"
	"ai-agent/internal/shared/logger"
	"ai-agent/internal/shared/worker"
	appdatabase "ai-agent/pkg/database"
	appjwt "ai-agent/pkg/jwt"
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type Server struct {
	log logger.Logger
}

func NewServer(log logger.Logger) *Server {
	return &Server{log: log}
}

func runServer() error {
	settings, err := config.Load()
	if err != nil {
		return err
	}

	zapLog, err := logger.Build(settings.LogLevel, settings.LogFormat, settings.LogSource)
	if err != nil {
		return err
	}
	defer zapLog.Sync()

	appLogger := logger.NewZap(zapLog)

	if settings.AppEnv != "production" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	if err := engine.SetTrustedProxies(nil); err != nil {
		return fmt.Errorf("configure trusted proxies: %w", err)
	}
	middleware.Setup(engine, settings.CORSOrigins)
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	oauthFactory := provider.NewFactory(provider.Config{
		GoogleClientID:     settings.OAuth.GoogleClientID,
		GoogleClientSecret: settings.OAuth.GoogleClientSecret,
		GoogleRedirectURL:  settings.OAuth.GoogleRedirectURL,
		GitHubClientID:     settings.OAuth.GitHubClientID,
		GitHubClientSecret: settings.OAuth.GitHubClientSecret,
		GitHubRedirectURL:  settings.OAuth.GitHubRedirectURL,
	}, nil)
	var db *sqlx.DB
	var jwtManager *appjwt.Manager
	if settings.JWTSecret != "" {
		jwtManager, err = appjwt.NewManager(settings.JWTSecret, settings.JWTAccessTTL, settings.JWTRefreshTTL)
		if err != nil {
			return err
		}
	}
	if settings.DatabaseURL != "" {
		db, err = appdatabase.NewPostgres(context.Background(), settings.DatabaseURL)
		if err != nil {
			return err
		}
		defer db.Close()
	}
	auth.LoadModule(auth.ModuleConfig{
		Router:   engine,
		Database: db,
		Provider: oauthFactory,
		JWT:      jwtManager,
		Logger:   appLogger,
	})
	protectedRouter := middleware.ProtectedGroup(engine, jwtManager, appLogger)
	workspacecore.LoadModule(workspacecore.ModuleConfig{
		Router:   protectedRouter,
		Database: db,
		Logger:   appLogger,
	})
	member.LoadModule(member.ModuleConfig{
		Router:   protectedRouter,
		Database: db,
		Logger:   appLogger,
	})
	mailSender := email.NewResendProvider(settings.ResendAPIKey, settings.ResendFromEmail, nil)
	jobQueue := worker.NewInMemoryPubSub(25)
	workerContext, stopWorkers := context.WithCancel(context.Background())
	defer jobQueue.Stop()
	defer stopWorkers()
	jobQueue.Consume(workerContext, 4)
	workspaceinvite.LoadModule(workspaceinvite.ModuleConfig{
		Router:      protectedRouter,
		Database:    db,
		Logger:      appLogger,
		Sender:      mailSender,
		Queue:       jobQueue,
		FrontendURL: settings.FrontendURL,
	})

	address := fmt.Sprintf("%s:%d", settings.Host, settings.Port)
	appLogger.With("component", "agent", "environment", settings.AppEnv).Info(context.Background(), "HTTP server started", "address", address)
	return engine.Run(address)
}
