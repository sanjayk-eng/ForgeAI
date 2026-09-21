package main

import (
	"ai-agent/internal/config"
	"ai-agent/internal/shared/logger"
	"context"
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

	appLogger.With("component", "agent","environment", settings.AppEnv,).Info(context.Background(),"agent started","host", settings.Host,"port", settings.Port)

	return nil
}
