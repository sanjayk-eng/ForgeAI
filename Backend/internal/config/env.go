package config

import (
	"fmt"
	"net/url"
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
		LogFormat: strings.ToLower(envOr(get, envLogFormat, defaultLogFormat)),
		LogSource: defaultLogSource,
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
