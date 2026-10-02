package organization

import (
	"ai-agent/internal/modules/organization/core"
	organizationinvite "ai-agent/internal/modules/organization/invite"
	"ai-agent/internal/modules/organization/member"
	"ai-agent/internal/shared/logger"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type ModuleConfig struct {
	Router       gin.IRouter
	PublicRouter gin.IRouter
	Database     *sqlx.DB
	Logger       logger.Logger
	Email        organizationinvite.EmailSender
	FrontendURL  string
}

type Module struct {
	CoreService   core.Service
	MemberService member.Service
	InviteService organizationinvite.Service
	Handler       *Handler
}

func LoadModule(config ModuleConfig) *Module {
	// Repositories
	coreRepo := core.NewRepository(config.Database)
	memberRepo := member.NewRepository(config.Database)

	// Services
	coreService := core.NewService(config.Database, coreRepo)
	memberService := member.NewService(config.Database, memberRepo)
	inviteService := organizationinvite.NewService(
		config.Database,
		organizationinvite.NewRepository(config.Database),
		coreService,
		config.Email,
		config.FrontendURL,
	)

	// Handler
	handler := NewHandler(coreService, memberService)

	// Routes
	RegisterRoutes(config.Router, handler)
	organizationinvite.RegisterRoutes(config.Router, config.PublicRouter, organizationinvite.NewHandler(inviteService))

	return &Module{
		CoreService:   coreService,
		MemberService: memberService,
		InviteService: inviteService,
		Handler:       handler,
	}
}
