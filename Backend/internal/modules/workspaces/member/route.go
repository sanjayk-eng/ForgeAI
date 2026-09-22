package member

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.GET("/workspaces/:workspace_id/members", handler.ListMembers)
	router.POST("/workspaces/:workspace_id/members", handler.AddMember)
	router.PATCH("/workspaces/:workspace_id/members/:user_id/role", handler.UpdateMemberRole)
	router.DELETE("/workspaces/:workspace_id/members/:user_id", handler.RemoveMember)
}
