package agent

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	terminalapp "ai-agent/internal/modules/terminal/application"
	"ai-agent/internal/modules/terminal/domain"
	"ai-agent/internal/shared/realtime"
)

const maxPromptLength = 12_000

var (
	ErrNotConfigured = errors.New("AI model is not configured")
	ErrInvalidPrompt = errors.New("prompt is required and must be within the size limit")
	ErrModelRequest  = errors.New("AI model request failed")
	ErrModelResponse = errors.New("AI model returned an invalid task response")
)

type SandboxService interface {
	ValidateSandboxAccess(ctx context.Context, userID, sandboxID string) error
	Get(ctx context.Context, sandboxID string) (domain.Sandbox, error)
	ListFiles(ctx context.Context, sandboxID, dir string) ([]terminalapp.FileEntry, error)
	ReadFile(ctx context.Context, sandboxID, path string) (string, error)
	WriteFile(ctx context.Context, sandboxID, relativePath, content string) error
}

type Config struct {
	BaseURL string
	APIKey  string
	Model   string
}

type Service struct {
	sandbox SandboxService
	config  Config
	client  *http.Client
	events  realtime.Publisher
}

func NewService(sandbox SandboxService, config Config, publishers ...realtime.Publisher) *Service {
	config.BaseURL = strings.TrimRight(strings.TrimSpace(config.BaseURL), "/")
	config.APIKey = strings.TrimSpace(config.APIKey)
	config.Model = strings.TrimSpace(config.Model)
	service := &Service{
		sandbox: sandbox,
		config:  config,
		client:  &http.Client{Timeout: 90 * time.Second},
	}
	if len(publishers) > 0 {
		service.events = publishers[0]
	}
	return service
}

func (service *Service) Status(ctx context.Context, userID, sandboxID string) (StatusResponse, error) {
	if err := service.sandbox.ValidateSandboxAccess(ctx, userID, sandboxID); err != nil {
		return StatusResponse{}, err
	}
	return StatusResponse{Configured: service.configured(), Model: service.config.Model}, nil
}

func (service *Service) RunTask(ctx context.Context, userID, sandboxID, prompt string) (taskResponse TaskResponse, taskErr error) {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" || len(prompt) > maxPromptLength {
		return TaskResponse{}, ErrInvalidPrompt
	}
	if !service.configured() {
		return TaskResponse{}, ErrNotConfigured
	}
	if err := service.sandbox.ValidateSandboxAccess(ctx, userID, sandboxID); err != nil {
		return TaskResponse{}, err
	}
	sandbox, err := service.sandbox.Get(ctx, sandboxID)
	if err != nil {
		return TaskResponse{}, err
	}
	service.publishAgentEvent(sandbox, "agent.started", "running", "Agent task started")
	defer func() {
		status, message := "completed", "Agent task completed"
		if taskErr != nil {
			status, message = "failed", "Agent task failed"
		}
		service.publishAgentEvent(sandbox, "agent.completed", status, message)
	}()

	projectContext, err := service.projectContext(ctx, sandboxID)
	if err != nil {
		return TaskResponse{}, fmt.Errorf("read project context: %w", err)
	}
	service.publishAgentEvent(sandbox, "agent.progress", "running", "Project context prepared")
	result, err := service.completeTask(ctx, prompt, projectContext)
	if err != nil {
		return TaskResponse{}, err
	}
	service.publishAgentEvent(sandbox, "agent.progress", "running", "Applying validated file changes")

	changes, err := validateChanges(result.Changes)
	if err != nil {
		return TaskResponse{}, err
	}
	changedFiles := make([]string, 0, len(changes))
	for _, change := range changes {
		if err := service.sandbox.WriteFile(ctx, sandboxID, change.Path, change.Content); err != nil {
			return TaskResponse{}, fmt.Errorf("apply agent change to %s: %w", change.Path, err)
		}
		changedFiles = append(changedFiles, change.Path)
	}

	message := strings.TrimSpace(result.Message)
	if message == "" {
		message = "Task completed."
	}
	return TaskResponse{Message: message, ChangedFiles: changedFiles}, nil
}

func (service *Service) publishAgentEvent(sandbox domain.Sandbox, eventType, status, message string) {
	if service.events == nil || sandbox.ProjectID == "" {
		return
	}
	service.events.Publish(realtime.Event{
		Version: 1, Event: eventType, WorkspaceID: sandbox.WorkspaceID,
		ProjectID: sandbox.ProjectID, SandboxID: sandbox.ID, Status: status, Message: message,
	})
}

func (service *Service) configured() bool {
	if service.sandbox == nil || service.config.APIKey == "" || service.config.Model == "" {
		return false
	}
	parsed, err := url.ParseRequestURI(service.config.BaseURL)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return false
	}
	return parsed.Scheme == "https" || parsed.Scheme == "http" && isLoopbackHost(parsed.Hostname())
}

func isLoopbackHost(host string) bool {
	return strings.EqualFold(host, "localhost") || host == "127.0.0.1" || host == "::1"
}
