package auth

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.GET("/auth/callback", handler.OAuthCallback)
}
