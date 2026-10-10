package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"ai-agent/internal/modules/github/types"
)

var ErrGitHubAppNotConfigured = errors.New("GitHub App integration is not configured")

// InstallationService handles GitHub App installation lifecycle
type InstallationService struct {
	repo      InstallationRepository
	apiClient GitHubAPIClient
	appConfig types.GitHubAppConfig
}

func NewInstallationService(repo InstallationRepository, client GitHubAPIClient, config types.GitHubAppConfig) *InstallationService {
	return &InstallationService{
		repo:      repo,
		apiClient: client,
		appConfig: config,
	}
}

func (s *InstallationService) StartInstallation(ctx context.Context, organizationID, userID string) (string, error) {
	if s.apiClient == nil || s.appConfig.Slug == "" {
		return "", ErrGitHubAppNotConfigured
	}
	
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", fmt.Errorf("generate GitHub installation state: %w", err)
	}
	
	state := base64.RawURLEncoding.EncodeToString(stateBytes)
	if err := s.repo.CreateState(ctx, hashState(state), organizationID, userID, time.Now().Add(10*time.Minute)); err != nil {
		return "", err
	}
	
	installURL := url.URL{Scheme: "https", Host: "github.com", Path: "/apps/" + s.appConfig.Slug + "/installations/new"}
	query := installURL.Query()
	query.Set("state", state)
	installURL.RawQuery = query.Encode()
	return installURL.String(), nil
}

func (s *InstallationService) CompleteInstallation(ctx context.Context, state, installationID string) (types.Integration, error) {
	if s.apiClient == nil {
		return types.Integration{}, ErrGitHubAppNotConfigured
	}
	
	state = strings.TrimSpace(state)
	if state == "" {
		return types.Integration{}, errors.New("installation state is invalid or expired")
	}
	
	stateHash := hashState(state)
	if _, err := s.repo.FindState(ctx, stateHash); err != nil {
		return types.Integration{}, err
	}
	
	installation, err := s.apiClient.VerifyInstallation(ctx, installationID)
	if err != nil {
		return types.Integration{}, err
	}
	
	return s.repo.CompleteInstallation(ctx, stateHash, installation)
}

func (s *InstallationService) Disconnect(ctx context.Context, organizationID, integrationID string) error {
	return s.repo.DisconnectGitHub(ctx, organizationID, integrationID)
}

func hashState(state string) string {
	hash := sha256.Sum256([]byte(state))
	return hex.EncodeToString(hash[:])
}

// InstallationRepository defines data access for installation operations
type InstallationRepository interface {
	CreateState(ctx context.Context, stateHash, organizationID, userID string, expiresAt time.Time) error
	FindState(ctx context.Context, stateHash string) (types.State, error)
	CompleteInstallation(ctx context.Context, stateHash string, installation types.Installation) (types.Integration, error)
	DisconnectGitHub(ctx context.Context, organizationID, integrationID string) error
}

// GitHubAPIClient defines GitHub API operations
type GitHubAPIClient interface {
	VerifyInstallation(ctx context.Context, installationID string) (types.Installation, error)
	InstallationToken(ctx context.Context, installationID string) (string, error)
	CreateRepository(ctx context.Context, token, accountName, accountType, name string, private bool) (types.ExportRepository, error)
}
