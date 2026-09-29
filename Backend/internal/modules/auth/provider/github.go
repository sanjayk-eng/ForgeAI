package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	gh "github.com/google/go-github/v92/github"
)

type GitHubProvider struct {
	config     Config
	httpClient *http.Client
}

var _ GitHubRepositoryInspector = (*GitHubProvider)(nil)

func NewGitHubProvider(config Config, httpClient *http.Client) *GitHubProvider {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &GitHubProvider{config: config, httpClient: httpClient}
}

func (provider *GitHubProvider) newAPIClient(accessToken string) (*gh.Client, error) {
	options := []gh.ClientOptionsFunc{gh.WithHTTPClient(provider.httpClient)}
	if strings.TrimSpace(accessToken) != "" {
		options = append(options, gh.WithAuthToken(accessToken))
	}
	return gh.NewClient(options...)
}

func (provider *GitHubProvider) InspectRepository(ctx context.Context, owner, name string) (GitHubRepository, error) {
	return provider.inspectRepository(ctx, owner, name, "")
}

func (provider *GitHubProvider) InspectRepositoryWithToken(ctx context.Context, owner, name, accessToken string) (GitHubRepository, error) {
	return provider.inspectRepository(ctx, owner, name, accessToken)
}

func (provider *GitHubProvider) inspectRepository(ctx context.Context, owner, name, accessToken string) (GitHubRepository, error) {
	client, err := provider.newAPIClient(accessToken)
	if err != nil {
		return GitHubRepository{}, fmt.Errorf("create GitHub API client: %w", err)
	}
	repository, _, err := client.Repositories.Get(ctx, owner, name)
	if err != nil {
		return GitHubRepository{}, fmt.Errorf("inspect GitHub repository: %w", err)
	}
	if repository == nil {
		return GitHubRepository{}, fmt.Errorf("GitHub repository response is empty")
	}
	result := mapGitHubRepository(repository)
	if result.ID <= 0 || result.Owner == "" || result.Name == "" || result.URL == "" || result.DefaultBranch == "" {
		return GitHubRepository{}, fmt.Errorf("GitHub repository response is incomplete")
	}
	result.Branches, err = provider.listBranches(ctx, client, owner, name)
	if err != nil {
		return GitHubRepository{}, fmt.Errorf("inspect GitHub repository branches: %w", err)
	}
	return result, nil
}

func (provider *GitHubProvider) listBranches(ctx context.Context, client *gh.Client, owner, name string) ([]string, error) {
	options := &gh.BranchListOptions{ListOptions: gh.ListOptions{PerPage: 100}}
	branches := make([]string, 0)
	for {
		page, response, err := client.Repositories.ListBranches(ctx, owner, name, options)
		if err != nil {
			return nil, err
		}
		for _, branch := range page {
			if branch != nil && branch.GetName() != "" {
				branches = append(branches, branch.GetName())
			}
		}
		if response == nil || response.NextPage == 0 {
			return branches, nil
		}
		options.Page = response.NextPage
	}
}

func mapGitHubRepository(repository *gh.Repository) GitHubRepository {
	owner := ""
	if repository.Owner != nil {
		owner = repository.Owner.GetLogin()
	}
	return GitHubRepository{
		ID:            repository.GetID(),
		Owner:         owner,
		Name:          repository.GetName(),
		URL:           repository.GetHTMLURL(),
		Private:       repository.GetPrivate(),
		DefaultBranch: repository.GetDefaultBranch(),
	}
}

func mapGitHubRepositories(repositories []*gh.Repository) []GitHubRepository {
	result := make([]GitHubRepository, 0, len(repositories))
	for _, repository := range repositories {
		if repository != nil {
			result = append(result, mapGitHubRepository(repository))
		}
	}
	return result
}

func (provider *GitHubProvider) ListOrganizations(ctx context.Context, accessToken string) ([]GitHubOrganization, error) {
	client, err := provider.newAPIClient(accessToken)
	if err != nil {
		return nil, fmt.Errorf("create GitHub API client: %w", err)
	}
	options := &gh.ListOptions{PerPage: 100}
	organizations := make([]GitHubOrganization, 0)
	scopes := ""
	for {
		page, response, err := client.Organizations.List(ctx, "", options)
		if err != nil {
			return nil, fmt.Errorf("list GitHub organizations: %w", err)
		}
		if response != nil {
			scopes = response.Header.Get("X-OAuth-Scopes")
		}
		for _, organization := range page {
			if organization != nil && organization.GetLogin() != "" {
				organizations = append(organizations, GitHubOrganization{Login: organization.GetLogin()})
			}
		}
		if response == nil || response.NextPage == 0 {
			break
		}
		options.Page = response.NextPage
	}
	if !hasOAuthScope(scopes, "read:org") {
		return nil, ErrGitHubOrganizationScopeRequired
	}
	return organizations, nil
}

func hasOAuthScope(scopesHeader, requiredScope string) bool {
	for _, scope := range strings.Split(scopesHeader, ",") {
		if strings.EqualFold(strings.TrimSpace(scope), requiredScope) {
			return true
		}
	}
	return false
}

func (provider *GitHubProvider) GetAccountLogin(ctx context.Context, accessToken string) (string, error) {
	client, err := provider.newAPIClient(accessToken)
	if err != nil {
		return "", fmt.Errorf("create GitHub API client: %w", err)
	}
	profile, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return "", fmt.Errorf("get GitHub account: %w", err)
	}
	if profile == nil || strings.TrimSpace(profile.GetLogin()) == "" {
		return "", fmt.Errorf("GitHub account response is missing login")
	}
	return profile.GetLogin(), nil
}

func (provider *GitHubProvider) ListRepositories(ctx context.Context, accessToken, organization string) ([]GitHubRepository, error) {
	client, err := provider.newAPIClient(accessToken)
	if err != nil {
		return nil, fmt.Errorf("create GitHub API client: %w", err)
	}
	organization = strings.TrimSpace(organization)
	userOptions := &gh.RepositoryListByAuthenticatedUserOptions{
		Type:        "all",
		Sort:        "updated",
		ListOptions: gh.ListOptions{PerPage: 100},
	}
	organizationOptions := &gh.RepositoryListByOrgOptions{
		Type:        "all",
		Sort:        "updated",
		ListOptions: gh.ListOptions{PerPage: 100},
	}
	repositories := make([]*gh.Repository, 0)
	pageNumber := 0
	for {
		var page []*gh.Repository
		var response *gh.Response
		if organization == "" {
			userOptions.Page = pageNumber
			page, response, err = client.Repositories.ListByAuthenticatedUser(ctx, userOptions)
		} else {
			organizationOptions.Page = pageNumber
			page, response, err = client.Repositories.ListByOrg(ctx, organization, organizationOptions)
		}
		if err != nil {
			return nil, fmt.Errorf("list GitHub repositories: %w", err)
		}
		repositories = append(repositories, page...)
		if response == nil || response.NextPage == 0 {
			break
		}
		pageNumber = response.NextPage
	}
	return mapGitHubRepositories(repositories), nil
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

	client, err := provider.newAPIClient(token.AccessToken)
	if err != nil {
		return ServiceUser{}, fmt.Errorf("create GitHub API client: %w", err)
	}
	profile, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return ServiceUser{}, fmt.Errorf("get GitHub user: %w", err)
	}
	if profile == nil {
		return ServiceUser{}, fmt.Errorf("get GitHub user: response is empty")
	}
	emailAddress := profile.GetEmail()
	if emailAddress == "" {
		options := &gh.ListOptions{PerPage: 100}
		for {
			emails, response, err := client.Users.ListEmails(ctx, options)
			if err != nil {
				return ServiceUser{}, fmt.Errorf("get GitHub email: %w", err)
			}
			for _, email := range emails {
				if email != nil && email.GetPrimary() && email.GetVerified() {
					emailAddress = email.GetEmail()
					break
				}
			}
			if emailAddress != "" || response == nil || response.NextPage == 0 {
				break
			}
			options.Page = response.NextPage
		}
		if emailAddress == "" {
			return ServiceUser{}, fmt.Errorf("GitHub account has no verified email")
		}
	}
	name := profile.GetName()
	if name == "" {
		name = profile.GetLogin()
	}
	return ServiceUser{
		ProviderID:  fmt.Sprintf("%d", profile.GetID()),
		Email:       emailAddress,
		Name:        name,
		AvatarURL:   profile.GetAvatarURL(),
		AccessToken: token.AccessToken,
	}, nil
}
