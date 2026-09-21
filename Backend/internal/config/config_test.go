package config

import (
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestLoadFromEnvUsesDefaults(t *testing.T) {
	config, err := loadFromEnv(func(string) string { return "" })
	if err != nil {
		t.Fatalf("loadFromEnv returned error: %v", err)
	}
	if config.AppEnv != "development" || config.Port != 8080 {
		t.Fatalf("unexpected defaults: %+v", config)
	}
	if config.LogLevel != zapcore.InfoLevel || config.LogFormat != "json" || !config.LogSource {
		t.Fatalf("unexpected logging defaults: %+v", config)
	}
}

func TestLoadFromEnvParsesValues(t *testing.T) {
	values := map[string]string{
		"APP_ENV": "production", "HOST": "0.0.0.0", "PORT": "9090",
		"DATABASE_URL": "postgres://localhost/forgeai", "LOG_LEVEL": "debug",
		"LOG_FORMAT": "text", "LOG_SOURCE": "false",
	}
	config, err := loadFromEnv(func(key string) string { return values[key] })
	if err != nil {
		t.Fatalf("loadFromEnv returned error: %v", err)
	}
	if config.Port != 9090 || config.LogLevel != zapcore.DebugLevel || config.LogFormat != "text" || config.LogSource {
		t.Fatalf("unexpected parsed config: %+v", config)
	}
}

func TestLoadFromEnvRejectsInvalidValues(t *testing.T) {
	for key, value := range map[string]string{"PORT": "0", "LOG_LEVEL": "trace", "LOG_FORMAT": "yaml", "LOG_SOURCE": "maybe"} {
		_, err := loadFromEnv(func(candidate string) string {
			if candidate == key {
				return value
			}
			return ""
		})
		if err == nil {
			t.Fatalf("expected %s=%q to be rejected", key, value)
		}
	}
}
