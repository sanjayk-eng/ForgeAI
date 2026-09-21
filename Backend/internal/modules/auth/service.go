package auth

import (
	"context"
	"database/sql"
	"errors"

	"ai-agent/internal/modules/auth/provider"
	apperrors "ai-agent/internal/shared/errors"
	"ai-agent/internal/shared/logger"
	appbcrypt "ai-agent/pkg/bcrypt"
	appdatabase "ai-agent/pkg/database"
	appjwt "ai-agent/pkg/jwt"

	"github.com/jmoiron/sqlx"
)

type Service struct {
	factory ProviderFactory
	db      *sqlx.DB
	repo    AuthRepository
	jwt     *appjwt.Manager
	log     logger.Logger
}

type ProviderFactory interface {
	Create(providerType string) (provider.ServiceProvider, error)
}

func NewService(factory ProviderFactory, db *sqlx.DB, repo AuthRepository, jwt *appjwt.Manager, log logger.Logger) *Service {
	return &Service{factory: factory, db: db, repo: repo, jwt: jwt, log: log}
}

func (service *Service) Register(ctx context.Context, input RegisterRequest) (AuthResponse, error) {
	if service.db == nil || service.repo == nil || service.jwt == nil {
		return AuthResponse{}, apperrors.NewCodedError(apperrors.ErrCodeInternalServer, "registration is not configured", nil)
	}
	passwordHash, err := appbcrypt.Hash(input.Password)
	if err != nil {
		return AuthResponse{}, apperrors.NewCodedError(apperrors.ErrCodeInternalServer, "password hashing failed", err)
	}
	var userID string
	if err := appdatabase.WithTx(ctx, service.db, func(ctx context.Context, tx *sqlx.Tx) error {
		var err error
		userID, err = service.repo.CreateUser(ctx, tx, input.Email, input.Name, passwordHash)
		return err
	}); err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			return AuthResponse{}, apperrors.NewCodedError(apperrors.ErrCodeConflict, "email is already registered", err)
		}
		return AuthResponse{}, apperrors.NewCodedError(apperrors.ErrCodeInternalServer, "registration failed", err)
	}
	token, err := service.jwt.Generate(userID)
	if err != nil {
		return AuthResponse{}, apperrors.NewCodedError(apperrors.ErrCodeInternalServer, "token generation failed", err)
	}
	return AuthResponse{Token: token}, nil
}

func (service *Service) Login(ctx context.Context, input LoginRequest) (AuthResponse, error) {
	if service.repo == nil || service.jwt == nil {
		return AuthResponse{}, apperrors.NewCodedError(apperrors.ErrCodeInternalServer, "login is not configured", nil)
	}
	credentials, err := service.repo.FindCredentials(ctx, input.Email)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return AuthResponse{}, apperrors.NewCodedError(apperrors.ErrCodeUnauthorized, "invalid email or password", err)
		}
		return AuthResponse{}, apperrors.NewCodedError(apperrors.ErrCodeInternalServer, "login failed", err)
	}
	if err := appbcrypt.Compare(credentials.PasswordHash, input.Password); err != nil {
		return AuthResponse{}, apperrors.NewCodedError(apperrors.ErrCodeUnauthorized, "invalid email or password", ErrInvalidCredentials)
	}
	token, err := service.jwt.Generate(credentials.ID)
	if err != nil {
		return AuthResponse{}, apperrors.NewCodedError(apperrors.ErrCodeInternalServer, "token generation failed", err)
	}
	return AuthResponse{Token: token}, nil
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

	result := OAuthUser{ProviderID: user.ProviderID, Email: user.Email, Name: user.Name, AvatarURL: user.AvatarURL}
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

	token, err := service.jwt.Generate(userID)
	if err != nil {
		return AuthResponse{}, apperrors.NewCodedError(apperrors.ErrCodeInternalServer, "token generation failed", err)
	}
	return AuthResponse{Token: token}, nil
}
