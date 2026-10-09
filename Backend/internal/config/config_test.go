package config

import (
	"os"
	"path/filepath"
	"strings"
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
	if config.PreviewOrigin != "http://preview.localhost:8080" {
		t.Fatalf("unexpected default preview origin: %q", config.PreviewOrigin)
	}
	if config.LogLevel != zapcore.InfoLevel || config.LogFormat != "json" || !config.LogSource {
		t.Fatalf("unexpected logging defaults: %+v", config)
	}
	if config.AIBaseURL != defaultAIBaseURL || config.AIModel != defaultAIModel || config.AIAPIKey != "" {
		t.Fatalf("unexpected AI defaults: %+v", config)
	}
}

func TestLoadFromEnvReadsGitHubAppPrivateKeyPath(t *testing.T) {
	keyPath := filepath.Join(t.TempDir(), "github-app.pem")
	keyContents := "test private key contents"
	if err := os.WriteFile(keyPath, []byte(keyContents), 0600); err != nil {
		t.Fatalf("write temporary private key: %v", err)
	}
	values := map[string]string{
		"GITHUB_APP_ID":               "12345",
		"GITHUB_APP_SLUG":             "forgeai",
		"GITHUB_APP_PRIVATE_KEY_PATH": keyPath,
		"GITHUB_APP_WEBHOOK_SECRET":   "test-webhook-secret",
	}
	config, err := loadFromEnv(func(key string) string { return values[key] })
	if err != nil {
		t.Fatalf("loadFromEnv returned error: %v", err)
	}
	if config.GitHubApp.PrivateKey != keyContents {
		t.Fatal("GitHub App private key was not loaded from the configured path")
	}
}

func TestLoadFromEnvRejectsEmptyGitHubAppPrivateKeyFile(t *testing.T) {
	keyPath := filepath.Join(t.TempDir(), "github-app.pem")
	if err := os.WriteFile(keyPath, nil, 0600); err != nil {
		t.Fatalf("write empty temporary private key: %v", err)
	}
	values := map[string]string{
		"GITHUB_APP_ID":               "12345",
		"GITHUB_APP_SLUG":             "forgeai",
		"GITHUB_APP_PRIVATE_KEY_PATH": keyPath,
		"GITHUB_APP_WEBHOOK_SECRET":   "test-webhook-secret",
	}
	_, err := loadFromEnv(func(key string) string { return values[key] })
	if err == nil || !strings.Contains(err.Error(), "private key file is empty") {
		t.Fatalf("expected an explicit empty private key error, got %v", err)
	}
}

func TestLoadFromEnvAllowsGitHubAppToRemainUnconfigured(t *testing.T) {
	values := map[string]string{
		"GITHUB_APP_SLUG":           "forgeai",
		"GITHUB_APP_WEBHOOK_SECRET": "placeholder",
	}
	if _, err := loadFromEnv(func(key string) string { return values[key] }); err != nil {
		t.Fatalf("partial optional GitHub App config should not stop server startup: %v", err)
	}
}

func TestPreviewOriginDefaultsToConfiguredBackendPort(t *testing.T) {
	values := map[string]string{"PORT": "8081"}
	config, err := loadFromEnv(func(key string) string { return values[key] })
	if err != nil {
		t.Fatalf("loadFromEnv returned error: %v", err)
	}
	if config.PreviewOrigin != "http://preview.localhost:8081" {
		t.Fatalf("preview origin = %q, want API port in origin", config.PreviewOrigin)
	}

	values[envPreviewOrigin] = "http://custom-preview.localhost:9090"
	config, err = loadFromEnv(func(key string) string { return values[key] })
	if err != nil {
		t.Fatalf("loadFromEnv with explicit preview origin returned error: %v", err)
	}
	if config.PreviewOrigin != values[envPreviewOrigin] {
		t.Fatalf("explicit preview origin = %q, want %q", config.PreviewOrigin, values[envPreviewOrigin])
	}
}

func TestLoadFromEnvParsesValues(t *testing.T) {
	values := map[string]string{
		"APP_ENV": "production", "HOST": "0.0.0.0", "PORT": "9090",
		"DATABASE_URL": "postgres://localhost/forgeai", "JWT_SECRET": "test-jwt-secret-long-enough",
		"PREVIEW_PUBLIC_ORIGIN": "https://preview.example.com", "LOG_LEVEL": "debug",
		"LOG_FORMAT": "text", "LOG_SOURCE": "false",
		envAIBaseURL: "https://llm.example/v1", envAIAPIKey: "configured-test-key", envAIModel: "code-model",
	}
	config, err := loadFromEnv(func(key string) string { return values[key] })
	if err != nil {
		t.Fatalf("loadFromEnv returned error: %v", err)
	}
	if config.Port != 9090 || config.LogLevel != zapcore.DebugLevel || config.LogFormat != "text" || config.LogSource {
		t.Fatalf("unexpected parsed config: %+v", config)
	}
	if config.AIBaseURL != values[envAIBaseURL] || config.AIAPIKey != values[envAIAPIKey] || config.AIModel != values[envAIModel] {
		t.Fatalf("unexpected AI config: %+v", config)
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
