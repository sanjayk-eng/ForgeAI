package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	gh "github.com/google/go-github/v92/github"
)

type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type OAuthUser struct {
	ProviderID  string
	Email       string
	Name        string
	AvatarURL   string
	AccessToken string
}

type OAuthProvider struct {
	config     OAuthConfig
	httpClient *http.Client
}

func NewOAuthProvider(config OAuthConfig, httpClient *http.Client) *OAuthProvider {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &OAuthProvider{config: config, httpClient: httpClient}
}

func (provider *OAuthProvider) ExchangeCode(ctx context.Context, code string) (OAuthUser, error) {
	redirectURL, err := url.Parse(provider.config.RedirectURL)
	if err != nil {
		return OAuthUser{}, fmt.Errorf("parse GitHub OAuth redirect URL: %w", err)
	}
	query := redirectURL.Query()
	query.Set("type", "github")
	redirectURL.RawQuery = query.Encode()

	form := url.Values{
		"client_id":     {provider.config.ClientID},
		"client_secret": {provider.config.ClientSecret},
		"code":          {code},
		"redirect_uri":  {redirectURL.String()},
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://github.com/login/oauth/access_token", strings.NewReader(form.Encode()))
	if err != nil {
		return OAuthUser{}, fmt.Errorf("create GitHub OAuth token request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")
	response, err := provider.httpClient.Do(request)
	if err != nil {
		return OAuthUser{}, fmt.Errorf("exchange GitHub OAuth code: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return OAuthUser{}, fmt.Errorf("GitHub OAuth endpoint returned HTTP %d", response.StatusCode)
	}
	var token struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(response.Body).Decode(&token); err != nil {
		return OAuthUser{}, fmt.Errorf("decode GitHub OAuth token: %w", err)
	}
	if token.AccessToken == "" {
		return OAuthUser{}, fmt.Errorf("exchange GitHub code: %s", token.Error)
	}

	client, err := gh.NewClient(gh.WithHTTPClient(provider.httpClient), gh.WithAuthToken(token.AccessToken))
	if err != nil {
		return OAuthUser{}, fmt.Errorf("create GitHub API client: %w", err)
	}
	profile, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return OAuthUser{}, fmt.Errorf("get GitHub user: %w", err)
	}
	if profile == nil {
		return OAuthUser{}, fmt.Errorf("get GitHub user: response is empty")
	}

	emailAddress := profile.GetEmail()
	if emailAddress == "" {
		options := &gh.ListOptions{PerPage: 100}
		for {
			emails, page, err := client.Users.ListEmails(ctx, options)
			if err != nil {
				return OAuthUser{}, fmt.Errorf("get GitHub email: %w", err)
			}
			for _, email := range emails {
				if email != nil && email.GetPrimary() && email.GetVerified() {
					emailAddress = email.GetEmail()
					break
				}
			}
			if emailAddress != "" || page == nil || page.NextPage == 0 {
				break
			}
			options.Page = page.NextPage
		}
		if emailAddress == "" {
			return OAuthUser{}, fmt.Errorf("GitHub account has no verified email")
		}
	}

	name := profile.GetName()
	if name == "" {
		name = profile.GetLogin()
	}
	return OAuthUser{
		ProviderID: fmt.Sprintf("%d", profile.GetID()),
		Email:      emailAddress, Name: name, AvatarURL: profile.GetAvatarURL(),
		AccessToken: token.AccessToken,
	}, nil
}
