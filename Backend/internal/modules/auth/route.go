package auth

import (
	"ai-agent/internal/middleware"
	"ai-agent/internal/shared/logger"
	appjwt "ai-agent/pkg/jwt"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.GET("/auth/callback", handler.OAuthCallback)
	router.POST("/auth/register", handler.Register)
	router.POST("/auth/login", handler.Login)
	router.POST("/auth/refresh", handler.Refresh)
}

func RegisterProtectedRoutes(router gin.IRouter, handler *Handler, jwtManager *appjwt.Manager, log logger.Logger) {
	protected := middleware.ProtectedGroup(router, jwtManager, log)
	protected.GET("/auth/me", handler.Me)
}
