package integration

import (
	organizationcore "ai-agent/internal/modules/organization/core"
	organizationmember "ai-agent/internal/modules/organization/member"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type ModuleConfig struct {
	Router        gin.IRouter
	PublicRouter  gin.IRouter
	Database      *sqlx.DB
	Organizations organizationcore.Service
	Members       organizationmember.Service
	GitHubApp     GitHubAppConfig
	FrontendURL   string
}

type Module struct {
	Service Service
	Handler *Handler
}

func LoadModule(config ModuleConfig) (*Module, error) {
	service, err := NewService(NewRepository(config.Database), config.GitHubApp)
	if err != nil {
		return nil, err
	}
	handler := NewHandler(service, config.Organizations, config.Members, config.FrontendURL)
	RegisterRoutes(config.Router, config.PublicRouter, handler)
	return &Module{Service: service, Handler: handler}, nil
}
