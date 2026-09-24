package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"ai-agent/internal/modules/auth/provider"
	apperrors "ai-agent/internal/shared/errors"
	"ai-agent/internal/shared/logger"
	appbcrypt "ai-agent/pkg/bcrypt"
	appdatabase "ai-agent/pkg/database"
	appjwt "ai-agent/pkg/jwt"

	"github.com/jmoiron/sqlx"
)

type Service struct {
	factory     ProviderFactory
	db          *sqlx.DB
	repo        AuthRepository
	jwt         *appjwt.Manager
	log         logger.Logger
	email       EmailService
	frontendURL string
}

type EmailService interface {
	SendVerification(ctx context.Context, to, verificationLink string) error
}

type ProviderFactory interface {
	Create(providerType string) (provider.ServiceProvider, error)
}

type ProviderURLFactory interface {
	AuthorizationURL(providerType string) (string, error)
}

func NewService(factory ProviderFactory, db *sqlx.DB, repo AuthRepository, jwt *appjwt.Manager, log logger.Logger, email EmailService, frontendURL string) *Service {
	return &Service{factory: factory, db: db, repo: repo, jwt: jwt, log: log, email: email, frontendURL: frontendURL}
}

func (service *Service) Register(ctx context.Context, input RegisterRequest) (AuthResponse, error) {
	if service.db == nil || service.repo == nil || service.email == nil {
		return AuthResponse{}, apperrors.NewCodedError(apperrors.ErrCodeInternalServer, "registration is not configured", nil)
	}
	input.Email = normalizeEmail(input.Email)
	verificationToken, err := newVerificationToken()
	if err != nil {
		return AuthResponse{}, apperrors.NewCodedError(apperrors.ErrCodeInternalServer, "verification setup failed", err)
	}
	passwordHash, err := appbcrypt.Hash(input.Password)
	if err != nil {
		return AuthResponse{}, apperrors.NewCodedError(apperrors.ErrCodeInternalServer, "password hashing failed", err)
	}
	var userID string
	if err := appdatabase.WithTx(ctx, service.db, func(ctx context.Context, tx *sqlx.Tx) error {
		var txErr error
		userID, txErr = service.repo.CreateUser(ctx, tx, input.Email, input.Name, passwordHash)
		if txErr != nil {
			return txErr
		}
		return service.repo.CreateEmailVerification(ctx, tx, userID, hashVerificationToken(verificationToken), time.Now().UTC().Add(24*time.Hour))
	}); err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			return AuthResponse{}, apperrors.NewCodedError(apperrors.ErrCodeConflict, "email is already registered", err)
		}
		return AuthResponse{}, apperrors.NewCodedError(apperrors.ErrCodeInternalServer, "registration failed", err)
	}
	verificationLink := strings.TrimRight(service.frontendURL, "/") + "/auth/verify?token=" + url.QueryEscape(verificationToken)
	if err := service.email.SendVerification(ctx, input.Email, verificationLink); err != nil {
		if service.log != nil {
			service.log.Error(ctx, "verification email enqueue failed", "error", err)
		}
		return AuthResponse{}, apperrors.NewCodedError(apperrors.ErrCodeInternalServer, "verification email could not be sent", err)
	}
	_ = userID
	return AuthResponse{}, nil
}

func (service *Service) Login(ctx context.Context, input LoginRequest) (AuthResponse, error) {
	if service.repo == nil || service.jwt == nil {
		return AuthResponse{}, apperrors.NewCodedError(apperrors.ErrCodeInternalServer, "login is not configured", nil)
	}
	input.Email = normalizeEmail(input.Email)
	credentials, err := service.repo.FindCredentials(ctx, input.Email)
	if err != nil {
		if errors.Is(err, ErrEmailNotVerified) {
			return AuthResponse{}, apperrors.NewCodedError(apperrors.ErrCodeUnauthorized, "please verify your email before signing in", err)
		}
		if errors.Is(err, ErrInvalidCredentials) {
			return AuthResponse{}, apperrors.NewCodedError(apperrors.ErrCodeUnauthorized, "invalid email or password", err)
		}
		return AuthResponse{}, apperrors.NewCodedError(apperrors.ErrCodeInternalServer, "login failed", err)
	}
	if err := appbcrypt.Compare(credentials.PasswordHash, input.Password); err != nil {
		return AuthResponse{}, apperrors.NewCodedError(apperrors.ErrCodeUnauthorized, "invalid email or password", ErrInvalidCredentials)
	}
	return service.issueTokens(credentials.ID)
}

func (service *Service) VerifyEmail(ctx context.Context, token string) error {
	if service.db == nil || service.repo == nil {
		return apperrors.NewCodedError(apperrors.ErrCodeInternalServer, "email verification is not configured", nil)
	}
	if strings.TrimSpace(token) == "" {
		return apperrors.NewCodedError(apperrors.ErrCodeBadRequest, "verification token is required", nil)
	}
	err := appdatabase.WithTx(ctx, service.db, func(ctx context.Context, tx *sqlx.Tx) error {
		return service.repo.VerifyEmailToken(ctx, tx, hashVerificationToken(token))
	})
	if errors.Is(err, ErrInvalidVerificationToken) {
		return apperrors.NewCodedError(apperrors.ErrCodeBadRequest, "verification link is invalid or expired", err)
	}
	if err != nil {
		return apperrors.NewCodedError(apperrors.ErrCodeInternalServer, "email verification failed", err)
	}
	return nil
}

func newVerificationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate verification token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

func hashVerificationToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (service *Service) Refresh(ctx context.Context, input RefreshTokenRequest) (AuthResponse, error) {
	if service.jwt == nil {
		return AuthResponse{}, apperrors.NewCodedError(apperrors.ErrCodeInternalServer, "refresh token is not configured", nil)
	}
	claims, err := service.jwt.Parse(input.RefreshToken, appjwt.RefreshToken)
	if err != nil {
		return AuthResponse{}, apperrors.NewCodedError(apperrors.ErrCodeUnauthorized, "invalid refresh token", err)
	}
	return service.issueTokens(claims.Subject)
}

func (service *Service) GetUser(ctx context.Context, userID string) (UserProfile, error) {
	if service.repo == nil {
		return UserProfile{}, apperrors.NewCodedError(apperrors.ErrCodeInternalServer, "user lookup is not configured", nil)
	}
	user, err := service.repo.FindUserByID(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return UserProfile{}, apperrors.NewCodedError(apperrors.ErrCodeNotFound, "user not found", err)
	}
	if err != nil {
		return UserProfile{}, apperrors.NewCodedError(apperrors.ErrCodeInternalServer, "user lookup failed", err)
	}
	return user, nil
}

func (service *Service) issueTokens(userID string) (AuthResponse, error) {
	pair, err := service.jwt.GeneratePair(userID)
	if err != nil {
		return AuthResponse{}, apperrors.NewCodedError(apperrors.ErrCodeInternalServer, "token generation failed", err)
	}
	return AuthResponse{AccessToken: pair.AccessToken, RefreshToken: pair.RefreshToken}, nil
}

func (service *Service) AuthenticateOAuth(ctx context.Context, providerType ProviderType, code string) (AuthResponse, error) {
	if service.factory == nil || service.db == nil || service.repo == nil || service.jwt == nil {
		return AuthResponse{}, apperrors.NewCodedError(apperrors.ErrCodeInternalServer, "OAuth is not configured", nil)
	}
	oauthProvider, err := service.factory.Create(string(providerType))
	if err != nil {
		if service.log != nil {
			service.log.Warn(ctx, "OAuth provider selection failed", "provider", providerType, "error", err)
		}
		return AuthResponse{}, err
	}
	user, err := oauthProvider.ExchangeCode(ctx, code)
	if err != nil && service.log != nil {
		service.log.Error(ctx, "OAuth code exchange failed", "provider", providerType, "error", err)
	}
	if err != nil {
		return AuthResponse{}, err
	}

	result := OAuthUser{ProviderID: user.ProviderID, Email: normalizeEmail(user.Email), Name: user.Name, AvatarURL: user.AvatarURL}
	var userID string
	if err := appdatabase.WithTx(ctx, service.db, func(ctx context.Context, tx *sqlx.Tx) error {
		userID, err = service.repo.FindOAuthUserID(ctx, tx, providerType, result.ProviderID)
		if err == nil {
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		userID, err = service.repo.FindUserIDByEmail(ctx, tx, result.Email)
		if errors.Is(err, sql.ErrNoRows) {
			userID, err = service.repo.CreateOAuthUser(ctx, tx, result.Email, result.Name)
		}
		if err != nil {
			return err
		}
		return service.repo.CreateOAuthAccount(ctx, tx, providerType, userID, result.ProviderID)
	}); err != nil {
		if service.log != nil {
			service.log.Error(ctx, "OAuth authentication persistence failed", "provider", providerType, "error", err)
		}
		return AuthResponse{}, apperrors.NewCodedError(apperrors.ErrCodeInternalServer, "OAuth authentication failed", err)
	}

	return service.issueTokens(userID)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (service *Service) OAuthAuthorizationURL(providerType ProviderType) (string, error) {
	authorizer, ok := service.factory.(ProviderURLFactory)
	if !ok {
		return "", apperrors.NewCodedError(apperrors.ErrCodeInternalServer, "OAuth authorization is not configured", nil)
	}
	redirectURL, err := authorizer.AuthorizationURL(string(providerType))
	if err != nil {
		return "", apperrors.NewCodedError(apperrors.ErrCodeBadRequest, err.Error(), err)
	}
	return redirectURL, nil
}
