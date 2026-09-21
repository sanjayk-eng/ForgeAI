package auth

import (
	"context"

	"ai-agent/internal/modules/auth/provider"
	"ai-agent/internal/shared/logger"
	appdatabase "ai-agent/pkg/database"

	"github.com/jmoiron/sqlx"
)

type Service struct {
	factory *provider.Factory
	db      *sqlx.DB
	repo    AuthRepository
	log     logger.Logger
}

func NewService(factory *provider.Factory, db *sqlx.DB, repo AuthRepository, log logger.Logger) *Service {
	return &Service{factory: factory, db: db, repo: repo, log: log}
}

func (service *Service) ExchangeCode(ctx context.Context, providerType ProviderType, code string) (OAuthUser, error) {
	oauthProvider, err := service.factory.Create(string(providerType))
	if err != nil {
		if service.log != nil {
			service.log.Warn(ctx, "OAuth provider selection failed", "provider", providerType, "error", err)
		}
		return OAuthUser{}, err
	}
	user, err := oauthProvider.ExchangeCode(ctx, code)
	if err != nil && service.log != nil {
		service.log.Error(ctx, "OAuth code exchange failed", "provider", providerType, "error", err)
	}
	if err != nil {
		return OAuthUser{}, err
	}

	result := OAuthUser{ProviderID: user.ProviderID, Email: user.Email, Name: user.Name, AvatarURL: user.AvatarURL}
	if service.db != nil && service.repo != nil {
		if err := appdatabase.WithTx(ctx, service.db, func(ctx context.Context, tx *sqlx.Tx) error {
			userID, err := service.repo.SaveUser(ctx, tx, result)
			if err != nil {
				return err
			}
			return service.repo.SaveOAuthAccount(ctx, tx, providerType, userID, result.ProviderID)
		}); err != nil {
			if service.log != nil {
				service.log.Error(ctx, "OAuth user persistence failed", "provider", providerType, "error", err)
			}
			return OAuthUser{}, err
		}
	}
	return result, nil
}
