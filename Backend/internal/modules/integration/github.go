package integration

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type githubAppClient struct {
	appID      int64
	privateKey *rsa.PrivateKey
	httpClient *http.Client
}

type githubAPIError struct {
	operation string
	status    int
}

func (err *githubAPIError) Error() string {
	return fmt.Sprintf("GitHub %s returned HTTP %d", err.operation, err.status)
}

func newGitHubAppClient(config GitHubAppConfig) (*githubAppClient, error) {
	keyPEM := strings.ReplaceAll(config.PrivateKey, `\n`, "\n")
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(keyPEM))
	if err != nil {
		return nil, fmt.Errorf("parse GitHub App private key: %w", err)
	}
	return &githubAppClient{appID: config.AppID, privateKey: key, httpClient: &http.Client{Timeout: 15 * time.Second}}, nil
}

func (client *githubAppClient) InstallationToken(ctx context.Context, rawInstallationID string) (string, error) {
	installationID, err := strconv.ParseInt(rawInstallationID, 10, 64)
	if err != nil || installationID < 1 {
		return "", fmt.Errorf("invalid GitHub installation ID")
	}
	token, err := client.appJWT()
	if err != nil {
		return "", err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", installationID), nil)
	if err != nil {
		return "", fmt.Errorf("create GitHub installation token request: %w", err)
	}
	setGitHubHeaders(request, token)
	response, err := client.httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("create GitHub installation token: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		return "", &githubAPIError{operation: "installation token request", status: response.StatusCode}
	}
	var payload struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode GitHub installation token: %w", err)
	}
	if strings.TrimSpace(payload.Token) == "" {
		return "", fmt.Errorf("GitHub returned an empty installation token")
	}
	return payload.Token, nil
}

func (client *githubAppClient) CreateRepository(ctx context.Context, token, accountName, accountType, name string, private bool) (ExportRepository, error) {
	accountName = strings.TrimSpace(accountName)
	name = strings.TrimSpace(name)
	if token == "" || accountName == "" || name == "" {
		return ExportRepository{}, fmt.Errorf("GitHub account, repository name, and installation token are required")
	}
	endpoint := "https://api.github.com/user/repos"
	if accountType == "Organization" {
		endpoint = "https://api.github.com/orgs/" + url.PathEscape(accountName) + "/repos"
	} else if accountType != "User" {
		return ExportRepository{}, fmt.Errorf("unsupported GitHub installation account type")
	}
	body, err := json.Marshal(map[string]any{"name": name, "private": private})
	if err != nil {
		return ExportRepository{}, fmt.Errorf("encode GitHub repository request: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return ExportRepository{}, fmt.Errorf("create GitHub repository request: %w", err)
	}
	setGitHubHeaders(request, token)
	request.Header.Set("Content-Type", "application/json")
	response, err := client.httpClient.Do(request)
	if err != nil {
		return ExportRepository{}, fmt.Errorf("create GitHub repository: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		return ExportRepository{}, &githubAPIError{operation: "repository creation", status: response.StatusCode}
	}
	var payload struct {
		ID            int64  `json:"id"`
		Name          string `json:"name"`
		FullName      string `json:"full_name"`
		HTMLURL       string `json:"html_url"`
		Private       bool   `json:"private"`
		DefaultBranch string `json:"default_branch"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return ExportRepository{}, fmt.Errorf("decode created GitHub repository: %w", err)
	}
	repositoryURL, err := url.Parse(payload.HTMLURL)
	if err != nil || repositoryURL.Scheme != "https" || !strings.EqualFold(repositoryURL.Hostname(), "github.com") ||
		repositoryURL.User != nil || payload.ID < 1 || payload.Name == "" || payload.FullName == "" {
		return ExportRepository{}, fmt.Errorf("GitHub returned invalid repository details")
	}
	return ExportRepository{
		RepositoryID: strconv.FormatInt(payload.ID, 10), RepositoryName: payload.Name,
		FullName: payload.FullName, RepositoryURL: strings.TrimSuffix(payload.HTMLURL, ".git"),
		DefaultBranch: payload.DefaultBranch, Private: payload.Private,
	}, nil
}

func setGitHubHeaders(request *http.Request, token string) {
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
}

func (client *githubAppClient) VerifyInstallation(ctx context.Context, rawInstallationID string) (Installation, error) {
	installationID, err := strconv.ParseInt(rawInstallationID, 10, 64)
	if err != nil || installationID < 1 {
		return Installation{}, fmt.Errorf("invalid GitHub installation ID")
	}
	token, err := client.appJWT()
	if err != nil {
		return Installation{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("https://api.github.com/app/installations/%d", installationID), nil)
	if err != nil {
		return Installation{}, fmt.Errorf("create GitHub installation request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	response, err := client.httpClient.Do(request)
	if err != nil {
		return Installation{}, fmt.Errorf("verify GitHub installation: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return Installation{}, fmt.Errorf("GitHub installation verification returned HTTP %d", response.StatusCode)
	}
	var payload struct {
		ID          int64      `json:"id"`
		AppID       int64      `json:"app_id"`
		SuspendedAt *time.Time `json:"suspended_at"`
		Account     struct {
			ID    int64  `json:"id"`
			Login string `json:"login"`
			Type  string `json:"type"`
		} `json:"account"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return Installation{}, fmt.Errorf("decode GitHub installation: %w", err)
	}
	if payload.ID != installationID || payload.AppID != client.appID || payload.SuspendedAt != nil {
		return Installation{}, fmt.Errorf("GitHub installation does not belong to this app or is suspended")
	}
	accountType := payload.Account.Type
	if payload.Account.ID < 1 || strings.TrimSpace(payload.Account.Login) == "" ||
		(accountType != "User" && accountType != "Organization") {
		return Installation{}, fmt.Errorf("GitHub installation account details are incomplete")
	}
	return Installation{
		ID: payload.ID, AppID: payload.AppID, AccountID: payload.Account.ID,
		AccountName: payload.Account.Login, AccountType: accountType,
	}, nil
}

func (client *githubAppClient) appJWT() (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:    strconv.FormatInt(client.appID, 10),
		IssuedAt:  jwt.NewNumericDate(now.Add(-time.Minute)),
		ExpiresAt: jwt.NewNumericDate(now.Add(9 * time.Minute)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(client.privateKey)
	if err != nil {
		return "", fmt.Errorf("sign GitHub App authentication token: %w", err)
	}
	return signed, nil
}
