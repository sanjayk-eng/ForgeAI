package github

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestOAuthProviderExchangesCodeAndUsesVerifiedEmail(t *testing.T) {
	var redirectURI string
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		switch request.URL.Host + request.URL.Path {
		case "github.com/login/oauth/access_token":
			form, err := url.ParseQuery(mustReadBody(t, request))
			if err != nil {
				t.Fatalf("parse token request: %v", err)
			}
			redirectURI = form.Get("redirect_uri")
			return jsonResponse(http.StatusOK, `{"access_token":"github-access-token"}`), nil
		case "api.github.com/user":
			if got := request.Header.Get("Authorization"); got != "Bearer github-access-token" {
				t.Errorf("user request Authorization = %q", got)
			}
			return jsonResponse(http.StatusOK, `{"id":123,"login":"octocat","name":"","avatar_url":"https://avatar.example/octocat"}`), nil
		case "api.github.com/user/emails":
			return jsonResponse(http.StatusOK, `[{"email":"octocat@example.com","primary":true,"verified":true}]`), nil
		default:
			return nil, errors.New("unexpected GitHub request: " + request.URL.String())
		}
	})}
	provider := NewOAuthProvider(OAuthConfig{
		ClientID: "client-id", ClientSecret: "client-secret",
		RedirectURL: "https://forge.example/auth/callback?source=login",
	}, client)

	user, err := provider.ExchangeCode(context.Background(), "oauth-code")
	if err != nil {
		t.Fatalf("ExchangeCode() error = %v", err)
	}
	if user.ProviderID != "123" || user.Email != "octocat@example.com" ||
		user.Name != "octocat" || user.AvatarURL != "https://avatar.example/octocat" ||
		user.AccessToken != "github-access-token" {
		t.Fatalf("ExchangeCode() user = %+v", user)
	}
	redirect, err := url.Parse(redirectURI)
	if err != nil {
		t.Fatalf("parse redirect URI = %q: %v", redirectURI, err)
	}
	if redirect.Query().Get("source") != "login" || redirect.Query().Get("type") != "github" {
		t.Fatalf("redirect URI query = %q", redirect.RawQuery)
	}
}

func mustReadBody(t *testing.T, request *http.Request) string {
	t.Helper()
	body, err := io.ReadAll(request.Body)
	if err != nil {
		t.Fatalf("read request body: %v", err)
	}
	return string(body)
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
