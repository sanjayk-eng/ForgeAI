package organization

import (
	"ai-agent/internal/modules/organization/core"
	"ai-agent/internal/modules/organization/member"
	"ai-agent/internal/shared/logger"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type ModuleConfig struct {
	Router   gin.IRouter
	Database *sqlx.DB
	Logger   logger.Logger
}

type Module struct {
	CoreService   core.Service
	MemberService member.Service
	Handler       *Handler
}

func LoadModule(config ModuleConfig) *Module {
	// Repositories
	coreRepo := core.NewRepository(config.Database)
	memberRepo := member.NewRepository(config.Database)

	// Services
	coreService := core.NewService(config.Database, coreRepo)
	memberService := member.NewService(config.Database, memberRepo)

	// Handler
	handler := NewHandler(coreService, memberService)

	// Routes
	RegisterRoutes(config.Router, handler)

	return &Module{
		CoreService:   coreService,
		MemberService: memberService,
		Handler:       handler,
	}
}
