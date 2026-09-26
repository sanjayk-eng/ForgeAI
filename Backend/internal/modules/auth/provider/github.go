package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type GitHubProvider struct {
	config     Config
	httpClient *http.Client
}

var _ GitHubRepositoryInspector = (*GitHubProvider)(nil)

func NewGitHubProvider(config Config, httpClient *http.Client) *GitHubProvider {
	return &GitHubProvider{config: config, httpClient: httpClient}
}

func (provider *GitHubProvider) InspectRepository(ctx context.Context, owner, name string) (GitHubRepository, error) {
	return provider.inspectRepository(ctx, owner, name, "")
}

func (provider *GitHubProvider) InspectRepositoryWithToken(ctx context.Context, owner, name, accessToken string) (GitHubRepository, error) {
	return provider.inspectRepository(ctx, owner, name, accessToken)
}

func (provider *GitHubProvider) inspectRepository(ctx context.Context, owner, name, accessToken string) (GitHubRepository, error) {
	var payload struct {
		ID            int64  `json:"id"`
		Name          string `json:"name"`
		HTMLURL       string `json:"html_url"`
		Private       bool   `json:"private"`
		DefaultBranch string `json:"default_branch"`
		Owner         struct {
			Login string `json:"login"`
		} `json:"owner"`
	}
	endpoint := "https://api.github.com/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(name)
	if err := getJSON(ctx, provider.httpClient, endpoint, accessToken, &payload); err != nil {
		return GitHubRepository{}, fmt.Errorf("inspect GitHub repository: %w", err)
	}
	if payload.ID <= 0 || payload.Owner.Login == "" || payload.Name == "" || payload.HTMLURL == "" || payload.DefaultBranch == "" {
		return GitHubRepository{}, fmt.Errorf("GitHub repository response is incomplete")
	}
	var branchesPayload []struct {
		Name string `json:"name"`
	}
	branchesEndpoint := "https://api.github.com/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(name) + "/branches?per_page=100"
	if err := getJSON(ctx, provider.httpClient, branchesEndpoint, accessToken, &branchesPayload); err != nil {
		return GitHubRepository{}, fmt.Errorf("inspect GitHub repository branches: %w", err)
	}
	branches := make([]string, 0, len(branchesPayload))
	for _, branch := range branchesPayload {
		if branch.Name != "" {
			branches = append(branches, branch.Name)
		}
	}
	return GitHubRepository{
		ID: payload.ID, Owner: payload.Owner.Login, Name: payload.Name,
		URL: payload.HTMLURL, Private: payload.Private, DefaultBranch: payload.DefaultBranch, Branches: branches,
	}, nil
}

func (provider *GitHubProvider) ListOrganizations(ctx context.Context, accessToken string) ([]GitHubOrganization, error) {
	var payload []GitHubOrganization
	if err := getJSON(ctx, provider.httpClient, "https://api.github.com/user/orgs?per_page=100", accessToken, &payload); err != nil {
		return nil, fmt.Errorf("list GitHub organizations: %w", err)
	}
	return payload, nil
}

func (provider *GitHubProvider) ListRepositories(ctx context.Context, accessToken, organization string) ([]GitHubRepository, error) {
	endpoint := "https://api.github.com/user/repos?per_page=100&sort=updated&type=all"
	if strings.TrimSpace(organization) != "" {
		endpoint = "https://api.github.com/orgs/" + url.PathEscape(organization) + "/repos?per_page=100&type=all&sort=updated"
	}
	var payload []struct {
		ID            int64  `json:"id"`
		Name          string `json:"name"`
		HTMLURL       string `json:"html_url"`
		DefaultBranch string `json:"default_branch"`
		Owner         struct {
			Login string `json:"login"`
		} `json:"owner"`
		Private bool `json:"private"`
	}
	if err := getJSON(ctx, provider.httpClient, endpoint, accessToken, &payload); err != nil {
		return nil, fmt.Errorf("list GitHub repositories: %w", err)
	}
	repositories := make([]GitHubRepository, 0, len(payload))
	for _, repository := range payload {
		repositories = append(repositories, GitHubRepository{
			ID: repository.ID, Owner: repository.Owner.Login, Name: repository.Name,
			URL: repository.HTMLURL, Private: repository.Private, DefaultBranch: repository.DefaultBranch,
		})
	}
	return repositories, nil
}

func (provider *GitHubProvider) ExchangeCode(ctx context.Context, code string) (ServiceUser, error) {
	var token oauthTokenResponse
	form := url.Values{
		"client_id":     {provider.config.GitHubClientID},
		"client_secret": {provider.config.GitHubClientSecret},
		"code":          {code},
		"redirect_uri":  {callbackURL(provider.config.GitHubRedirectURL, "github")},
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
	if profile.Email == "" {
		var emails []gitHubEmail
		if err := getJSON(ctx, provider.httpClient, "https://api.github.com/user/emails", token.AccessToken, &emails); err != nil {
			return ServiceUser{}, fmt.Errorf("get GitHub email: %w", err)
		}
		for _, email := range emails {
			if email.Primary && email.Verified {
				profile.Email = email.Email
				break
			}
		}
		if profile.Email == "" {
			return ServiceUser{}, fmt.Errorf("GitHub account has no verified email")
		}
	}
	name := profile.Name
	if name == "" {
		name = profile.Login
	}
	return ServiceUser{ProviderID: fmt.Sprintf("%d", profile.ID), Email: profile.Email, Name: name, AvatarURL: profile.AvatarURL, AccessToken: token.AccessToken}, nil
}
