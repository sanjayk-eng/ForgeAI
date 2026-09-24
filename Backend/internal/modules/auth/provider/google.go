package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type GoogleProvider struct {
	config     Config
	httpClient *http.Client
}

func NewGoogleProvider(config Config, httpClient *http.Client) *GoogleProvider {
	return &GoogleProvider{config: config, httpClient: httpClient}
}

func (provider *GoogleProvider) ExchangeCode(ctx context.Context, code string) (ServiceUser, error) {
	var token oauthTokenResponse
	form := url.Values{
		"client_id":     {provider.config.GoogleClientID},
		"client_secret": {provider.config.GoogleClientSecret},
		"code":          {code},
		"redirect_uri":  {callbackURL(provider.config.GoogleRedirectURL, "google")},
		"grant_type":    {"authorization_code"},
	}
	if err := postForm(ctx, provider.httpClient, "https://oauth2.googleapis.com/token", form, &token); err != nil {
		return ServiceUser{}, fmt.Errorf("exchange Google code: %w", err)
	}
	if token.AccessToken == "" {
		return ServiceUser{}, fmt.Errorf("exchange Google code: %s", token.Error)
	}

	var profile googleProfile
	if err := getJSON(ctx, provider.httpClient, "https://openidconnect.googleapis.com/v1/userinfo", token.AccessToken, &profile); err != nil {
		return ServiceUser{}, fmt.Errorf("get Google user: %w", err)
	}
	return ServiceUser{ProviderID: profile.ID, Email: profile.Email, Name: profile.Name, AvatarURL: profile.Picture}, nil
}
