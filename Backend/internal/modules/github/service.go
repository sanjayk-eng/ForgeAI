package github

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

var ErrGitHubAppNotConfigured = errors.New("GitHub App integration is not configured")

type Service interface {
	List(ctx context.Context, organizationID string) ([]Integration, error)
	CreateExportRepository(ctx context.Context, organizationID, projectID, userID string, private bool) (ExportRepository, error)
	GetExportRepository(ctx context.Context, organizationID, projectID string) (ExportRepository, error)
	ExportCredentials(ctx context.Context, organizationID, projectID string) (string, string, error)
	StartInstallation(ctx context.Context, organizationID, userID string) (string, error)
	CompleteInstallation(ctx context.Context, state, installationID string) (Integration, error)
	Disconnect(ctx context.Context, organizationID, integrationID string) error
	HandleWebhook(ctx context.Context, event, signature string, body []byte) error
}

type service struct {
	repository Repository
	github     *githubAppClient
	appConfig  GitHubAppConfig
}

func NewService(repository Repository, config GitHubAppConfig) (Service, error) {
	var client *githubAppClient
	if config.AppID != 0 {
		var err error
		client, err = newGitHubAppClient(config)
		if err != nil {
			return nil, err
		}
	}
	return &service{repository: repository, github: client, appConfig: config}, nil
}

func (s *service) List(ctx context.Context, organizationID string) ([]Integration, error) {
	return s.repository.ListGitHub(ctx, organizationID)
}

func (s *service) CreateExportRepository(ctx context.Context, organizationID, projectID, userID string, private bool) (ExportRepository, error) {
	if s.github == nil {
		return ExportRepository{}, ErrGitHubAppNotConfigured
	}
	exportContext, err := s.repository.FindExportContext(ctx, organizationID, projectID)
	if err != nil {
		return ExportRepository{}, err
	}
	if exportContext.Existing != nil {
		return *exportContext.Existing, nil
	}
	token, err := s.github.InstallationToken(ctx, exportContext.InstallationID)
	if err != nil {
		return ExportRepository{}, err
	}
	created, err := s.github.CreateRepository(ctx, token, exportContext.AccountName, exportContext.AccountType, exportContext.ProjectSlug, private)
	if err != nil {
		return ExportRepository{}, err
	}
	created.IntegrationID = exportContext.IntegrationID
	return s.repository.SaveExportRepository(ctx, projectID, exportContext.IntegrationID, created, userID)
}

func (s *service) GetExportRepository(ctx context.Context, organizationID, projectID string) (ExportRepository, error) {
	return s.repository.FindExportRepository(ctx, organizationID, projectID)
}

func (s *service) ExportCredentials(ctx context.Context, organizationID, projectID string) (string, string, error) {
	if s.github == nil {
		return "", "", ErrGitHubAppNotConfigured
	}
	installationID, repositoryURL, err := s.repository.FindExportCredentials(ctx, organizationID, projectID)
	if err != nil {
		return "", "", err
	}
	token, err := s.github.InstallationToken(ctx, installationID)
	if err != nil {
		return "", "", err
	}
	return token, repositoryURL, nil
}

func (s *service) StartInstallation(ctx context.Context, organizationID, userID string) (string, error) {
	if s.github == nil || s.appConfig.Slug == "" {
		return "", ErrGitHubAppNotConfigured
	}
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", fmt.Errorf("generate GitHub installation state: %w", err)
	}
	state := base64.RawURLEncoding.EncodeToString(stateBytes)
	if err := s.repository.CreateState(ctx, hashState(state), organizationID, userID, time.Now().Add(10*time.Minute)); err != nil {
		return "", err
	}
	installURL := url.URL{Scheme: "https", Host: "github.com", Path: "/apps/" + s.appConfig.Slug + "/installations/new"}
	query := installURL.Query()
	query.Set("state", state)
	installURL.RawQuery = query.Encode()
	return installURL.String(), nil
}

func (s *service) CompleteInstallation(ctx context.Context, state, installationID string) (Integration, error) {
	if s.github == nil {
		return Integration{}, ErrGitHubAppNotConfigured
	}
	state = strings.TrimSpace(state)
	if state == "" {
		return Integration{}, ErrStateInvalid
	}
	stateHash := hashState(state)
	if _, err := s.repository.FindState(ctx, stateHash); err != nil {
		return Integration{}, err
	}
	installation, err := s.github.VerifyInstallation(ctx, installationID)
	if err != nil {
		return Integration{}, err
	}
	return s.repository.CompleteInstallation(ctx, stateHash, installation)
}

func (s *service) Disconnect(ctx context.Context, organizationID, integrationID string) error {
	return s.repository.DisconnectGitHub(ctx, organizationID, integrationID)
}

func (s *service) HandleWebhook(ctx context.Context, event, signature string, body []byte) error {
	if s.appConfig.WebhookSecret == "" {
		return ErrGitHubAppNotConfigured
	}
	if !validWebhookSignature(s.appConfig.WebhookSecret, signature, body) {
		return errors.New("invalid GitHub webhook signature")
	}
	if event != "installation" {
		return nil
	}
	var payload installationWebhook
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("decode GitHub installation webhook: %w", err)
	}
	if payload.Installation.ID < 1 || payload.Installation.AppID != s.appConfig.AppID {
		return errors.New("GitHub webhook installation does not belong to this app")
	}
	status := ""
	switch payload.Action {
	case "deleted":
		status = StatusRevoked
	case "suspend":
		status = StatusInactive
	case "unsuspend", "new_permissions_accepted", "created":
		status = StatusActive
	default:
		return nil
	}
	return s.repository.SetInstallationStatus(ctx, fmt.Sprint(payload.Installation.ID), status)
}

func hashState(state string) string {
	hash := sha256.Sum256([]byte(state))
	return hex.EncodeToString(hash[:])
}

func validWebhookSignature(secret, signature string, body []byte) bool {
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}
	provided, err := hex.DecodeString(strings.TrimPrefix(signature, "sha256="))
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return hmac.Equal(provided, mac.Sum(nil))
}
