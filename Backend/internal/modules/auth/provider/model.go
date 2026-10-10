package provider

type ServiceUser struct {
	ProviderID  string `json:"provider_id"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	AvatarURL   string `json:"avatar_url"`
	AccessToken string `json:"-"`
}

type oauthTokenResponse struct {
	AccessToken string `json:"access_token"`
	Error       string `json:"error"`
}

type googleProfile struct {
	ID      string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}
