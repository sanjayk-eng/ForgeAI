package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	AppEnv          string
	Host            string
	Port            int
	DatabaseURL     string
	JWTSecret       string
	ResendAPIKey    string
	ResendFromEmail string
	FrontendURL     string
	JWTAccessTTL    time.Duration
	JWTRefreshTTL   time.Duration
	CORSOrigins     string
	OAuth           OAuthConfig
	LogLevel        zapcore.Level
	LogFormat       string
	LogSource       bool
}

type OAuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURL  string
}

const (
	envAppEnv          = "APP_ENV"
	envHost            = "HOST"
	envPort            = "PORT"
	envDatabaseURL     = "DATABASE_URL"
	envJWTSecret       = "JWT_SECRET"
	envResendAPIKey    = "RESEND_API_KEY"
	envResendFromEmail = "RESEND_FROM_EMAIL"
	envFrontendURL     = "FRONTEND_URL"
	envJWTAccessTTL    = "JWT_ACCESS_TTL"
	envJWTRefreshTTL   = "JWT_REFRESH_TTL"
	envCORSOrigins     = "CORS_ALLOWED_ORIGINS"
	envLogLevel        = "LOG_LEVEL"
	envLogFormat       = "LOG_FORMAT"
	envLogSource       = "LOG_SOURCE"

	defaultAppEnv        = "development"
	defaultHost          = "127.0.0.1"
	defaultPort          = "8080"
	defaultCORSOrigins   = "*"
	defaultJWTAccessTTL  = "15m"
	defaultJWTRefreshTTL = "168h"
	defaultLogLevel      = "info"
	defaultLogFormat     = "json"
	defaultLogSource     = true
	defaultFrontendURL   = "http://localhost:5173"
)

var (
	loadOnce sync.Once
	loaded   Config
	loadErr  error
)

func Load() (Config, error) {
	loadOnce.Do(func() {
		loadErr = loadDotEnv()
		if loadErr != nil {
			return
		}
		loaded, loadErr = loadFromEnv(os.Getenv)
	})
	return loaded, loadErr
}

func loadDotEnv() error {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	directory := workingDirectory
	for range 4 {
		path := filepath.Join(directory, ".env")
		if _, err := os.Stat(path); err == nil {
			if err := godotenv.Load(path); err != nil {
				return fmt.Errorf("load %s: %w", path, err)
			}
			return nil
		}
		directory = filepath.Dir(directory)
	}
	return nil
}
