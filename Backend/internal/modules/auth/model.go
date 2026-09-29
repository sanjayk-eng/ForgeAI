package auth

import "errors"

type ProviderType string

const (
	ProviderGoogle ProviderType = "google"
	ProviderGitHub ProviderType = "github"
)

type OAuthUser s
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

var ErrGitHubAccountNotConnected = errors.New("GitHub account is not connected")

var ErrInvalidCredentials = errors.New("invalid credentials")

var ErrEmailNotVerified = errors.New("email is not verified")

var ErrInvalidVerificationToken = errors.New("invalid or expired verification token")
