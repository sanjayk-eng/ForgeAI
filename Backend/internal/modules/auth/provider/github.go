package provider

import (
	"context"
	"net/http"

	githubmodule "ai-agent/internal/modules/github"
)

type GitHubProvider struct {
	oauth *githubmodule.OAuthProvider
}

func NewGitHubProvider(config Config, httpClient *http.Client) *GitHubProvider {
	return &GitHubProvider{oauth: githubmodule.NewOAuthProvider(githubmodule.OAuthConfig{
		ClientID: config.GitHubClientID, ClientSecret: config.GitHubClientSecret,
		RedirectURL: config.GitHubRedirectURL,
	}, httpClient)}
}

func (provider *GitHubProvider) ExchangeCode(ctx context.Context, code string) (ServiceUser, error) {
	user, err := provider.oauth.ExchangeCode(ctx, code)
	if err != nil {
		return ServiceUser{}, err
	}
	return ServiceUser{
		ProviderID: user.ProviderID, Email: user.Email, Name: user.Name,
		AvatarURL: user.AvatarURL, AccessToken: user.AccessToken,
	}, nil
}
