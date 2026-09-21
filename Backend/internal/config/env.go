package config

import "strings"

type getenv func(string) string

func loadFromEnv(get getenv) (Config, error) {
	config := Config{
		AppEnv:      envOr(get, envAppEnv, defaultAppEnv),
		Host:        envOr(get, envHost, defaultHost),
		DatabaseURL: get(envDatabaseURL),
		LogFormat:   strings.ToLower(envOr(get, envLogFormat, defaultLogFormat)),
		LogSource:   defaultLogSource,
	}

	port, err := parsePort(envOr(get, envPort, defaultPort))
	if err != nil {
		return Config{}, err
	}
	config.Port = port

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
