package main

import (
	"ai-agent/internal/config"
	"ai-agent/internal/middleware"
	"ai-agent/internal/modules/auth"
	"ai-agent/internal/modules/auth/provider"
	"ai-agent/internal/shared/logger"
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
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
	auth.RegisterRoutes(engine, auth.NewHandler(auth.NewService(oauthFactory, appLogger)))

	address := fmt.Sprintf("%s:%d", settings.Host, settings.Port)
	appLogger.With("component", "agent", "environment", settings.AppEnv).Info(context.Background(), "HTTP server started", "address", address)
	return engine.Run(address)
}
