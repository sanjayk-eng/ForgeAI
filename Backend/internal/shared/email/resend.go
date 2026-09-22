package email

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type ResendProvider struct {
	apiKey  string
	from    string
	baseURL string
	client  *http.Client
}

func NewResendProvider(apiKey, from string, httpClient *http.Client) *ResendProvider {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	baseURL := strings.TrimSpace("https://api.resend.com")
	return &ResendProvider{apiKey: strings.TrimSpace(apiKey), from: strings.TrimSpace(from), baseURL: baseURL, client: httpClient}
}

func (provider *ResendProvider) Send(ctx context.Context, message Message) error {
	if provider == nil {
		return errors.New("resend provider is nil")
	}
	if strings.TrimSpace(provider.apiKey) == "" {
		return errors.New("resend API key is required")
	}
	if strings.TrimSpace(provider.from) == "" {
		return errors.New("resend sender email is required")
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

	payload := struct {
		From    string   `json:"from"`
		To      []string `json:"to"`
		Subject string   `json:"subject"`
		Text    string   `json:"text,omitempty"`
		HTML    string   `json:"html,omitempty"`
	}{
		From:    provider.from,
		To:      []string{message.To},
		Subject: message.Subject,
		Text:    message.Text,
		HTML:    message.HTML,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal resend email: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, provider.baseURL+"/emails", strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("create resend request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+provider.apiKey)
	request.Header.Set("Content-Type", "application/json")

	response, err := provider.client.Do(request)
	if err != nil {
		return fmt.Errorf("send resend email: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		responseBody, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return fmt.Errorf("resend email failed with status %d: %s", response.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	return nil
}
