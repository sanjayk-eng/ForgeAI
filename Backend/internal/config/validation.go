package config

import (
	"fmt"
	"strconv"
	"strings"

	"go.uber.org/zap/zapcore"
)

func parsePort(value string) (int, error) {
	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		if err == nil {
			err = fmt.Errorf("value out of range")
		}
		return 0, fmt.Errorf("PORT must be between 1 and 65535: %w", err)
	}
	return port, nil
}

func parseLevel(value string) (zapcore.Level, error) {
	switch strings.ToLower(value) {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn", "warning":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	default:
		return 0, fmt.Errorf("LOG_LEVEL must be debug, info, warn, or error, got %q", value)
	}
}

func parseBool(get getenv, key string, fallback bool) (bool, error) {
	raw := get(key)
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s must be true or false: %w", key, err)
	}
	return value, nil
}

func validateFormat(format string) error {
	if format != "json" && format != "text" {
		return fmt.Errorf("LOG_FORMAT must be json or text, got %q", format)
	}
	return nil
}
