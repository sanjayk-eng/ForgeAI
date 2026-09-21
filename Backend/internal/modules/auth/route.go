package auth

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.GET("/auth/callback", handler.OAuthCallback)
	router.POST("/auth/register", handler.Register)
	router.POST("/auth/login", handler.Login)
}
