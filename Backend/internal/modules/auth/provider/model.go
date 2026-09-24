package provider

type ServiceUser struct {
	ProviderID string `json:"provider_id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	AvatarURL  string `json:"avatar_url"`
}

type oauthTokenResponse struct {
	AccessToken string `json:"access_token"`
	Error       string `json:"error"`
}

type GitHubRepository struct {
	ID            int64    `json:"id"`
	Owner         string   `json:"owner"`
	Name          string   `json:"name"`
	URL           string   `json:"html_url"`
	DefaultBranch string   `json:"default_branch"`
	Branches      []string `json:"branches"`
}

type gitHubProfile struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
}

type gitHubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

type googleProfile struct {
	ID      string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}
