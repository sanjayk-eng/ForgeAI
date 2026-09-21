package provider

import (
	"fmt"
	"net/http"
	"strings"
)

type Config struct {
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURL  string
}

type Factory struct {
	config     Config
	httpClient *http.Client
}

func NewFactory(config Config, httpClient *http.Client) *Factory {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Factory{config: config, httpClient: httpClient}
}

func (factory *Factory) Create(providerType string) (ServiceProvider, error) {
	switch strings.ToLower(strings.TrimSpace(string(providerType))) {
	case "google":
		return NewGoogleProvider(factory.config, factory.httpClient), nil
	case "github":
		return NewGitHubProvider(factory.config, factory.httpClient), nil
	default:
		return nil, fmt.Errorf("unsupported OAuth provider %q", providerType)
	}
}
