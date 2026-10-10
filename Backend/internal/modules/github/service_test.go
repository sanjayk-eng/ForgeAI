package github

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestValidWebhookSignature(t *testing.T) {
	secret := "webhook-secret"
	body := []byte(`{"action":"deleted"}`)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	if !validWebhookSignature(secret, signature, body) {
		t.Fatal("expected valid GitHub webhook signature to be accepted")
	}
	if validWebhookSignature(secret, signature, []byte(`{"action":"created"}`)) {
		t.Fatal("expected modified webhook body to be rejected")
	}
	if validWebhookSignature(secret, "sha1="+hex.EncodeToString(mac.Sum(nil)), body) {
		t.Fatal("expected unsupported signature scheme to be rejected")
	}
}

func TestDisconnectRemovesOnlyOrganizationAssociation(t *testing.T) {
	repository := &disconnectTestRepository{}
	service := &service{repository: repository}

	if err := service.Disconnect(context.Background(), "forge-org", "integration"); err != nil {
		t.Fatalf("Disconnect() error = %v", err)
	}
	if repository.organizationID != "forge-org" || repository.integrationID != "integration" {
		t.Fatalf("Disconnect() association = %q/%q, want forge-org/integration",
			repository.organizationID, repository.integrationID)
	}
}

type disconnectTestRepository struct {
	Repository
	organizationID string
	integrationID  string
}

func (repository *disconnectTestRepository) DisconnectGitHub(_ context.Context, organizationID, integrationID string) error {
	repository.organizationID = organizationID
	repository.integrationID = integrationID
	return nil
}
