package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"ai-agent/internal/modules/github/types"
)

// WebhookService handles GitHub webhook events
type WebhookService struct {
	repo          WebhookRepository
	webhookSecret string
	appID         int64
}

type installationWebhook struct {
	Action       string `json:"action"`
	Installation struct {
		ID    int64 `json:"id"`
		AppID int64 `json:"app_id"`
	} `json:"installation"`
}

func NewWebhookService(repo WebhookRepository, webhookSecret string, appID int64) *WebhookService {
	return &WebhookService{
		repo:          repo,
		webhookSecret: webhookSecret,
		appID:         appID,
	}
}

func (s *WebhookService) HandleWebhook(ctx context.Context, event, signature string, body []byte) error {
	if s.webhookSecret == "" {
		return ErrGitHubAppNotConfigured
	}
	
	if !s.validWebhookSignature(signature, body) {
		return errors.New("invalid GitHub webhook signature")
	}
	
	if event != "installation" {
		return nil
	}
	
	var payload installationWebhook
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("decode GitHub installation webhook: %w", err)
	}
	
	if payload.Installation.ID < 1 || payload.Installation.AppID != s.appID {
		return errors.New("GitHub webhook installation does not belong to this app")
	}
	
	status := ""
	switch payload.Action {
	case "deleted":
		status = types.StatusRevoked
	case "suspend":
		status = types.StatusInactive
	case "unsuspend", "new_permissions_accepted", "created":
		status = types.StatusActive
	default:
		return nil
	}
	
	return s.repo.SetInstallationStatus(ctx, fmt.Sprint(payload.Installation.ID), status)
}

func (s *WebhookService) validWebhookSignature(signature string, body []byte) bool {
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}
	
	provided, err := hex.DecodeString(strings.TrimPrefix(signature, "sha256="))
	if err != nil {
		return false
	}
	
	mac := hmac.New(sha256.New, []byte(s.webhookSecret))
	_, _ = mac.Write(body)
	return hmac.Equal(provided, mac.Sum(nil))
}

// WebhookRepository defines data access for webhook operations
type WebhookRepository interface {
	SetInstallationStatus(ctx context.Context, installationID, status string) error
}
