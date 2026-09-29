package provider

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func jsonResponse(body string, headers http.Header) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     headers,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func oauthScopeHeader(scopes string) http.Header {
	header := make(http.Header)
	header.Set("X-OAuth-Scopes", scopes)
	return header
}

func TestHasOAuthScope(t *testing.T) {
	tests := []struct {
		name        string
		header      string
		required    string
		wantPresent bool
	}{
		{name: "scope present", header: "repo, read:org, user:email", required: "read:org", wantPresent: true},
		{name: "scope with whitespace", header: "repo,  read:org ", required: "read:org", wantPresent: true},
		{name: "scope missing", header: "repo, user:email", required: "read:org"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := hasOAuthScope(test.header, test.required); got != test.wantPresent {
				t.Errorf("hasOAuthScope(%q, %q) = %v, want %v", test.header, test.required, got, test.wantPresent)
			}
		})
	}
}

func TestListRepositoriesUsesAuthenticatedGoGitHubClientAndPaginates(t *testing.T) {
	paths := make([]string, 0, 2)
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if got := request.Header.Get("Authorization"); got != "Bearer github-token" {
			t.Errorf("authorization header = %q, want bearer token", got)
		}
		if got := request.URL.Path; got != "/user/repos" {
			t.Errorf("request path = %q, want /user/repos", got)
		}
		if got := request.URL.Query().Get("type"); got != "all" {
			t.Errorf("repository type = %q, want all", got)
		}
		paths = append(paths, request.URL.Query().Get("page"))
		if len(paths) == 1 {
			return jsonResponse(`[{"id":1,"name":"first","html_url":"https://github.com/acme/first","private":true,"default_branch":"main","owner":{"login":"acme"}}]`, http.Header{
				"Link": {`<https://api.github.com/user/repos?page=2>; rel="next"`},
			}), nil
		}
		return jsonResponse(`[{"id":2,"name":"second","html_url":"https://github.com/acme/second","private":false,"default_branch":"trunk","owner":{"login":"acme"}}]`, make(http.Header)), nil
	})}
	provider := NewGitHubProvider(Config{}, client)

	repositories, err := provider.ListRepositories(context.Background(), "github-token", "")
	if err != nil {
		t.Fatalf("ListRepositories returned error: %v", err)
	}
	if len(repositories) != 2 {
		t.Fatalf("repository count = %d, want 2", len(repositories))
	}
	if repositories[0].Name != "first" || !repositories[0].Private || repositories[1].DefaultBranch != "trunk" {
		t.Fatalf("repository data was not mapped correctly: %#v", repositories)
	}
	if len(paths) != 2 || paths[1] != "2" {
		t.Fatalf("requested pages = %#v, want first and second pages", paths)
	}
}

func TestListRepositoriesUsesOrganizationEndpoint(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.Path != "/orgs/forge-team/repos" {
			t.Errorf("request path = %q, want /orgs/forge-team/repos", request.URL.Path)
		}
		if request.Header.Get("Authorization") != "Bearer github-token" {
			t.Errorf("expected authenticated GitHub API request")
		}
		return jsonResponse(`[{"id":3,"name":"service","html_url":"https://github.com/forge-team/service","private":true,"default_branch":"main","owner":{"login":"forge-team"}}]`, make(http.Header)), nil
	})}
	provider := NewGitHubProvider(Config{}, client)

	repositories, err := provider.ListRepositories(context.Background(), "github-token", " forge-team ")
	if err != nil {
		t.Fatalf("ListRepositories returned error: %v", err)
	}
	if len(repositories) != 1 || repositories[0].Owner != "forge-team" || repositories[0].Name != "service" {
		t.Fatalf("repositories = %#v, want forge-team/service", repositories)
	}
}

func TestInspectRepositoryUsesGoGitHubAndPaginatesBranches(t *testing.T) {
	branchPages := 0
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Header.Get("Authorization") != "Bearer repo-token" {
			t.Errorf("expected authenticated GitHub API request")
		}
		switch request.URL.Path {
		case "/repos/acme/api":
			return jsonResponse(`{"id":42,"name":"api","html_url":"https://github.com/acme/api","private":true,"default_branch":"main","owner":{"login":"acme"}}`, make(http.Header)), nil
		case "/repos/acme/api/branches":
			branchPages++
			if branchPages == 1 {
				return jsonResponse(`[{"name":"main"}]`, http.Header{
					"Link": {`<https://api.github.com/repos/acme/api/branches?page=2>; rel="next"`},
				}), nil
			}
			return jsonResponse(`[{"name":"release"}]`, make(http.Header)), nil
		default:
			t.Errorf("unexpected GitHub API path %q", request.URL.Path)
			return jsonResponse(`{}`, make(http.Header)), nil
		}
	})}
	provider := NewGitHubProvider(Config{}, client)

	repository, err := provider.InspectRepositoryWithToken(context.Background(), "acme", "api", "repo-token")
	if err != nil {
		t.Fatalf("InspectRepositoryWithToken returned error: %v", err)
	}
	if repository.ID != 42 || repository.Owner != "acme" || repository.DefaultBranch != "main" {
		t.Fatalf("repository data was not mapped correctly: %#v", repository)
	}
	if len(repository.Branches) != 2 || repository.Branches[0] != "main" || repository.Branches[1] != "release" {
		t.Fatalf("branches = %#v, want both pages", repository.Branches)
	}
}

func TestListOrganizationsUsesGoGitHubAndChecksOAuthScope(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.Path != "/user/orgs" {
			t.Errorf("request path = %q, want /user/orgs", request.URL.Path)
		}
		if request.Header.Get("Authorization") != "Bearer org-token" {
			t.Errorf("expected authenticated GitHub API request")
		}
		return jsonResponse(`[{"login":"forge-team"}]`, oauthScopeHeader("repo, read:org, user:email")), nil
	})}
	provider := NewGitHubProvider(Config{}, client)

	organizations, err := provider.ListOrganizations(context.Background(), "org-token")
	if err != nil {
		t.Fatalf("ListOrganizations returned error: %v", err)
	}
	if len(organizations) != 1 || organizations[0].Login != "forge-team" {
		t.Fatalf("organizations = %#v, want forge-team", organizations)
	}
}

func TestListOrganizationsRejectsMissingReadOrgScope(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(`[]`, oauthScopeHeader("repo, user:email")), nil
	})}
	provider := NewGitHubProvider(Config{}, client)

	_, err := provider.ListOrganizations(context.Background(), "org-token")
	if err != ErrGitHubOrganizationScopeRequired {
		t.Fatalf("ListOrganizations error = %v, want missing-scope error", err)
	}
}
