package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type getenv func(string) string

func loadFromEnv(get getenv) (Config, error) {
	port, err := parsePort(envOr(get, envPort, defaultPort))
	if err != nil {
		return Config{}, err
	}
	config := Config{
		AppEnv:            envOr(get, envAppEnv, defaultAppEnv),
		Host:              envOr(get, envHost, defaultHost),
		Port:              port,
		DatabaseURL:       get(envDatabaseURL),
		JWTSecret:         get(envJWTSecret),
		ResendAPIKey:      get(envResendAPIKey),
		ResendFromEmail:   get(envResendFromEmail),
		FrontendURL:       envOr(get, envFrontendURL, defaultFrontendURL),
		PreviewOrigin:     get(envPreviewOrigin),
		PreviewSigningKey: get(envPreviewSigningKey),
		AIBaseURL:         envOr(get, envAIBaseURL, defaultAIBaseURL),
		AIAPIKey:          get(envAIAPIKey),
		AIModel:           envOr(get, envAIModel, defaultAIModel),
		CORSOrigins:       envOr(get, envCORSOrigins, defaultCORSOrigins),
		OAuth: OAuthConfig{
			GoogleClientID:     get("GOOGLE_CLIENT_ID"),
			GoogleClientSecret: get("GOOGLE_CLIENT_SECRET"),
			GoogleRedirectURL:  get("GOOGLE_REDIRECT_URL"),
			GitHubClientID:     get("GITHUB_CLIENT_ID"),
			GitHubClientSecret: get("GITHUB_CLIENT_SECRET"),
			GitHubRedirectURL:  get("GITHUB_REDIRECT_URL"),
		},
		GitHubApp: GitHubAppConfig{
			Slug:           get("GITHUB_APP_SLUG"),
			PrivateKey:     get("GITHUB_APP_PRIVATE_KEY"),
			PrivateKeyPath: get("GITHUB_APP_PRIVATE_KEY_PATH"),
			WebhookSecret:  get("GITHUB_APP_WEBHOOK_SECRET"),
		},
		LogFormat: strings.ToLower(envOr(get, envLogFormat, defaultLogFormat)),
		LogSource: defaultLogSource,
	}
	if appID := get("GITHUB_APP_ID"); appID != "" {
		config.GitHubApp.AppID, err = strconv.ParseInt(appID, 10, 64)
		if err != nil || config.GitHubApp.AppID < 1 {
			return Config{}, fmt.Errorf("GITHUB_APP_ID must be a positive integer")
		}
	}
	if config.GitHubApp.PrivateKeyPath != "" {
		privateKey, err := os.ReadFile(config.GitHubApp.PrivateKeyPath)
		if err != nil {
			return Config{}, fmt.Errorf("read GitHub App private key from configured path: %w", err)
		}
		if strings.TrimSpace(string(privateKey)) == "" {
			return Config{}, fmt.Errorf("GitHub App private key file is empty")
		}
		if config.GitHubApp.PrivateKey != "" {
			return Config{}, fmt.Errorf("configure either GITHUB_APP_PRIVATE_KEY or GITHUB_APP_PRIVATE_KEY_PATH, not both")
		}
		config.GitHubApp.PrivateKey = string(privateKey)
	}
	if config.GitHubApp.AppID != 0 && (config.GitHubApp.Slug == "" ||
		config.GitHubApp.PrivateKey == "" || config.GitHubApp.WebhookSecret == "") {
		return Config{}, fmt.Errorf("configure GITHUB_APP_SLUG, GITHUB_APP_PRIVATE_KEY or GITHUB_APP_PRIVATE_KEY_PATH, and GITHUB_APP_WEBHOOK_SECRET when GITHUB_APP_ID is set")
	}
	if config.PreviewOrigin == "" {
		config.PreviewOrigin = fmt.Sprintf("http://preview.localhost:%d", config.Port)
	}
	if config.PreviewSigningKey == "" {
		config.PreviewSigningKey = config.JWTSecret
		if config.PreviewSigningKey == "" && config.AppEnv != "production" {
			config.PreviewSigningKey = defaultPreviewSigningKey
		}
	}
	if config.AppEnv == "production" && len(config.PreviewSigningKey) < 16 {
		return Config{}, fmt.Errorf("PREVIEW_SIGNING_KEY or JWT_SECRET must contain at least 16 bytes in production")
	}
	if config.AppEnv == "production" {
		previewOrigin, err := url.Parse(config.PreviewOrigin)
		if err != nil || previewOrigin.Scheme != "https" || previewOrigin.Hostname() == "" || previewOrigin.User != nil ||
			(previewOrigin.Path != "" && previewOrigin.Path != "/") || previewOrigin.RawQuery != "" || previewOrigin.Fragment != "" {
			return Config{}, fmt.Errorf("PREVIEW_PUBLIC_ORIGIN must be an HTTPS origin in production")
		}
	}

	config.JWTAccessTTL, err = time.ParseDuration(envOr(get, envJWTAccessTTL, defaultJWTAccessTTL))
	if err != nil || config.JWTAccessTTL <= 0 {
		return Config{}, fmt.Errorf("JWT_ACCESS_TTL must be a positive duration")
	}
	config.JWTRefreshTTL, err = time.ParseDuration(envOr(get, envJWTRefreshTTL, defaultJWTRefreshTTL))
	if err != nil || config.JWTRefreshTTL <= 0 {
		return Config{}, fmt.Errorf("JWT_REFRESH_TTL must be a positive duration")
	}

	config.LogLevel, err = parseLevel(envOr(get, envLogLevel, defaultLogLevel))
	if err != nil {
		return Config{}, err
	}

	config.LogSource, err = parseBool(get, envLogSource, defaultLogSource)
	if err != nil {
		return Config{}, err
	}
	if err := validateFormat(config.LogFormat); err != nil {
		return Config{}, err
	}

	return config, nil
}

func envOr(get getenv, key string, fallback string) string {
	if value := get(key); value != "" {
		return value
	}
	return fallback
}
