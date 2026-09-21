package auth

import (
	"context"

	"ai-agent/internal/modules/auth/provider"
	"ai-agent/internal/shared/logger"
)

type Service struct {
	factory *provider.Factory
	log     logger.Logger
}

func NewService(factory *provider.Factory, log logger.Logger) *Service {
	return &Service{factory: factory, log: log}
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
	return OAuthUser{ProviderID: user.ProviderID, Email: user.Email, Name: user.Name, AvatarURL: user.AvatarURL}, err
}
