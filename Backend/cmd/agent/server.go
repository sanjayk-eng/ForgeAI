package main

import (
	"ai-agent/internal/config"
	"ai-agent/internal/middleware"
	agentmodule "ai-agent/internal/modules/agent"
	"ai-agent/internal/modules/auth"
	"ai-agent/internal/modules/auth/provider"
	gitmodule "ai-agent/internal/modules/git"
	organizationmodule "ai-agent/internal/modules/organization"
	previewmodule "ai-agent/internal/modules/preview"
	projectmodule "ai-agent/internal/modules/project"
	terminalmodule "ai-agent/internal/modules/terminal"
	terminalworker "ai-agent/internal/modules/terminal/worker"
	"ai-agent/internal/shared/email"
	"ai-agent/internal/shared/logger"
	appdatabase "ai-agent/pkg/database"
	appjwt "ai-agent/pkg/jwt"
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

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
	apiRouter := engine.Group("/api/v1")
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
	// Initialize email module
	emailModule, err := email.NewModule(email.Config{
		Provider:    "resend",
		APIKey:      settings.ResendAPIKey,
		FromEmail:   settings.ResendFromEmail,
		QueueSize:   100,
		WorkerCount: 3,
		Logger:      appLogger,
	})
	if err != nil {
		return fmt.Errorf("initialize email module: %w", err)
	}

	// Start email workers in background
	emailContext, stopEmail := context.WithCancel(context.Background())
	defer stopEmail()
	defer emailModule.Stop()
	emailModule.Start(emailContext)

	appLogger.Info(context.Background(), "email service initialized", "queue_size", 100, "workers", 3)

	authModule := auth.LoadModule(auth.ModuleConfig{
		Router:       apiRouter,
		Database:     db,
		Provider:     oauthFactory,
		JWT:          jwtManager,
		Logger:       appLogger,
		EmailService: emailModule.Service,
		FrontendURL:  settings.FrontendURL,
	})
	protectedRouter := middleware.ProtectedGroup(apiRouter, jwtManager, appLogger)

	// Load organization module (replaces workspace)
	organizationModule := organizationmodule.LoadModule(organizationmodule.ModuleConfig{
		Router:       protectedRouter,
		PublicRouter: apiRouter,
		Database:     db,
		Logger:       appLogger,
		Email:        emailModule.Service,
		FrontendURL:  settings.FrontendURL,
	})

	// Simplified project module (no GitHub client, no sync)
	projectModule := projectmodule.LoadModule(projectmodule.ModuleConfig{
		Router:       protectedRouter,
		Database:     db,
		Logger:       appLogger,
		Organization: organizationModule.CoreService,
	})

	projectAdapter := terminalworker.NewProjectAdapter(
		projectModule.CoreService,
	)
	policyPath, err := sandboxPolicyPath()
	if err != nil {
		return fmt.Errorf("initialize terminal module: %w", err)
	}

	terminalModule, err := terminalmodule.LoadModule(terminalmodule.ModuleConfig{
		Database:         db,
		Logger:           appLogger,
		DockerBinary:     "docker",
		PolicyPath:       policyPath,
		ProjectRepo:      projectAdapter,
		WebSocketOrigins: webSocketOrigins(settings.CORSOrigins, settings.FrontendURL),
	})
	if err != nil {
		return fmt.Errorf("initialize terminal module: %w", err)
	}
	projectModule.SetProjectCreatedHook(terminalModule.OnProjectCreated)
	projectModule.SetProjectDeletedHook(terminalModule.OnProjectDeleted)
	previewHandler, err := previewmodule.NewHandler(terminalModule.Service, terminalModule.Preview, settings.PreviewOrigin, settings.PreviewSigningKey)
	if err != nil {
		return fmt.Errorf("initialize preview module: %w", err)
	}
	previewmodule.RegisterRoutes(protectedRouter, previewHandler)
	engine.NoRoute(previewHandler.Gateway)

	terminalModule.Start(emailContext, 3)
	defer terminalModule.Stop()

	gitmodule.LoadModule(gitmodule.ModuleConfig{
		Router:          protectedRouter,
		WorkspaceRoot:   "/workspace",
		SandboxExecutor: terminalModule.Service,
		SandboxAccess:   terminalModule.Service,
		GitHubAccounts:  authModule.Repository,
		Events:          terminalModule.Events,
	})

	terminalmodule.RegisterRoutes(protectedRouter, terminalModule.Handler)
	agentService := agentmodule.NewService(terminalModule.Service, agentmodule.Config{
		BaseURL: settings.AIBaseURL,
		APIKey:  settings.AIAPIKey,
		Model:   settings.AIModel,
	}, terminalModule.Events)
	agentmodule.RegisterRoutes(protectedRouter, agentmodule.NewHandler(agentService))

	address := fmt.Sprintf("%s:%d", settings.Host, settings.Port)
	appLogger.With("component", "agent", "environment", settings.AppEnv).Info(context.Background(), "HTTP server started", "address", address)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := engine.Run(address); err != nil {
			appLogger.Error(context.Background(), "server error", "error", err)
			quit <- syscall.SIGTERM
		}
	}()

	<-quit
	appLogger.Info(context.Background(), "shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stopEmail()
	terminalModule.Stop()
	emailModule.Stop()

	select {
	case <-shutdownCtx.Done():
		appLogger.Warn(context.Background(), "shutdown timeout exceeded")
	default:
		appLogger.Info(context.Background(), "server shutdown complete")
	}

	return nil
}

func webSocketOrigins(corsOrigins, frontendURL string) []string {
	seen := make(map[string]struct{})
	origins := make([]string, 0)
	for _, origin := range append(strings.Split(corsOrigins, ","), frontendURL) {
		origin = strings.TrimSpace(origin)
		if origin == "" || origin == "*" {
			continue
		}
		if _, exists := seen[origin]; exists {
			continue
		}
		seen[origin] = struct{}{}
		origins = append(origins, origin)
	}
	return origins
}

func sandboxPolicyPath() (string, error) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	directory := workingDirectory
	for range 8 {
		candidate := filepath.Join(directory, "configs", "sandbox.yaml")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			break
		}
		directory = parent
	}
	return "", fmt.Errorf("sandbox policy not found from %s", workingDirectory)
}
