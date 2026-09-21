package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type GitHubProvider struct {
	config     Config
	httpClient *http.Client
}

func NewGitHubProvider(config Config, httpClient *http.Client) *GitHubProvider {
	return &GitHubProvider{config: config, httpClient: httpClient}
}

func (provider *GitHubProvider) ExchangeCode(ctx context.Context, code string) (ServiceUser, error) {
	var token oauthTokenResponse
	form := url.Values{
		"client_id":     {provider.config.GitHubClientID},
		"client_secret": {provider.config.GitHubClientSecret},
		"code":          {code},
		"redirect_uri":  {provider.config.GitHubRedirectURL},
	}
	if err := postForm(ctx, provider.httpClient, "https://github.com/login/oauth/access_token", form, &token); err != nil {
		return ServiceUser{}, fmt.Errorf("exchange GitHub code: %w", err)
	}
	if token.AccessToken == "" {
		return ServiceUser{}, fmt.Errorf("exchange GitHub code: %s", token.Error)
	}

	var profile gitHubProfile
	if err := getJSON(ctx, provider.httpClient, "https://api.github.com/user", token.AccessToken, &profile); err != nil {
		return ServiceUser{}, fmt.Errorf("get GitHub user: %w", err)
	}
	name := profile.Name
	if name == "" {
		name = profile.Login
	}
	return ServiceUser{ProviderID: fmt.Sprintf("%d", profile.ID), Email: profile.Email, Name: name, AvatarURL: profile.AvatarURL}, nil
}
