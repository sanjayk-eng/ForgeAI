package config

import (
	"fmt"
	"strings"
	"time"
)

type getenv func(string) string

func loadFromEnv(get getenv) (Config, error) {
	config := Config{
		AppEnv:      envOr(get, envAppEnv, defaultAppEnv),
		Host:        envOr(get, envHost, defaultHost),
		DatabaseURL: get(envDatabaseURL),
		JWTSecret:   get(envJWTSecret),
		CORSOrigins: envOr(get, envCORSOrigins, defaultCORSOrigins),
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

	port, err := parsePort(envOr(get, envPort, defaultPort))
	if err != nil {
		return Config{}, err
	}
	config.Port = port
	config.JWTTTL, err = time.ParseDuration(envOr(get, envJWTTTL, defaultJWTTTL))
	if err != nil || config.JWTTTL <= 0 {
		return Config{}, fmt.Errorf("JWT_TTL must be a positive duration")
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
