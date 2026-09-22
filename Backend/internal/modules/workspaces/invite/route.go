package invite

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.POST("/workspaces/:workspace_id/invites", handler.CreateInvite)
	router.GET("/workspaces/:workspace_id/invites", handler.ListInvites)
	router.POST("/workspaces/:workspace_id/invites/:invite_id/status", handler.UpdateInviteStatus)
}
