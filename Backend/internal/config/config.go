package config

import (
	"os"
	"sync"

	"go.uber.org/zap/zapcore"
)

type Config struct {
	AppEnv      string
	Host        string
	Port        int
	DatabaseURL string
	LogLevel    zapcore.Level
	LogFormat   string
	LogSource   bool
}

const (
	envAppEnv      = "APP_ENV"
	envHost        = "HOST"
	envPort        = "PORT"
	envDatabaseURL = "DATABASE_URL"
	envLogLevel    = "LOG_LEVEL"
	envLogFormat   = "LOG_FORMAT"
	envLogSource   = "LOG_SOURCE"

	defaultAppEnv    = "development"
	defaultHost      = "127.0.0.1"
	defaultPort      = "8080"
	defaultLogLevel  = "info"
	defaultLogFormat = "json"
	defaultLogSource = true
)

var (
	loadOnce sync.Once
	loaded   Config
	loadErr  error
)


func Load() (Config, error) {
	loadOnce.Do(func() {
		loaded, loadErr = loadFromEnv(os.Getenv)
	})
	return loaded, loadErr
}
