package auth

type ProviderType string

const (
	ProviderGoogle ProviderType = "google"
	ProviderGitHub ProviderType = "github"
)

type OAuthUser struct {
	ProviderID string `json:"provider_id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	AvatarURL  string `json:"avatar_url"`
}
