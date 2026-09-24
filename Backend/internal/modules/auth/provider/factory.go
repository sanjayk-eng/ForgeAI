package provider

import (
	"fmt"
	"net/http"
	"net/url"
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

func (factory *Factory) AuthorizationURL(providerType string) (string, error) {
	providerName := strings.ToLower(strings.TrimSpace(providerType))
	values := url.Values{"response_type": {"code"}}

	switch providerName {
	case "github":
		if factory.config.GitHubClientID == "" || factory.config.GitHubRedirectURL == "" {
			return "", fmt.Errorf("GitHub OAuth is not configured")
		}
		values.Set("client_id", factory.config.GitHubClientID)
		values.Set("redirect_uri", callbackURL(factory.config.GitHubRedirectURL, providerName))
		values.Set("scope", "read:user user:email")
		return "https://github.com/login/oauth/authorize?" + values.Encode(), nil
	case "google":
		if factory.config.GoogleClientID == "" || factory.config.GoogleRedirectURL == "" {
			return "", fmt.Errorf("Google OAuth is not configured")
		}
		values.Set("client_id", factory.config.GoogleClientID)
		values.Set("redirect_uri", callbackURL(factory.config.GoogleRedirectURL, providerName))
		values.Set("scope", "openid email profile")
		values.Set("access_type", "offline")
		return "https://accounts.google.com/o/oauth2/v2/auth?" + values.Encode(), nil
	default:
		return "", fmt.Errorf("unsupported OAuth provider %q", providerType)
	}
}

func callbackURL(baseURL, providerType string) string {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}
	query := parsedURL.Query()
	query.Set("type", providerType)
	parsedURL.RawQuery = query.Encode()
	return parsedURL.String()
}
