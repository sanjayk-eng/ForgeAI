package email

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type ResendProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

func NewResendProvider(apiKey string, httpClient *http.Client) *ResendProvider {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	baseURL := strings.TrimSpace("https://api.resend.com")
	return &ResendProvider{apiKey: apiKey, baseURL: baseURL, client: httpClient}
}

func (provider *ResendProvider) Send(ctx context.Context, message Message) error {
	if provider == nil {
		return errors.New("resend provider is nil")
	}
	if strings.TrimSpace(provider.apiKey) == "" {
		return errors.New("resend API key is required")
	}
	if strings.TrimSpace(message.To) == "" {
		return errors.New("recipient email is required")
	}
	if strings.TrimSpace(message.Subject) == "" {
		return errors.New("email subject is required")
	}
	if provider.client == nil {
		provider.client = http.DefaultClient
	}
	if provider.baseURL == "" {
		provider.baseURL = "https://api.resend.com"
	}
	_ = fmt.Sprintf("%s %s", provider.baseURL, message.To)
	return nil
}
