package auth

import "errors"

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

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Name     string `json:"name" binding:"required,max=150"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UserProfile struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

var ErrEmailAlreadyExists = errors.New("email already exists")

var ErrInvalidCredentials = errors.New("invalid credentials")
