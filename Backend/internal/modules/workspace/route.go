package workspace

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.POST("/workspaces", handler.CreateWorkspace)
	router.GET("/workspaces", handler.ListWorkspaces)
	router.GET("/workspaces/:workspace_id", handler.GetWorkspace)
	router.GET("/workspaces/:workspace_id/members", handler.ListMembers)
	router.POST("/workspaces/:workspace_id/members", handler.AddMember)
	router.PATCH("/workspaces/:workspace_id/members/:user_id/role", handler.UpdateMemberRole)
	router.DELETE("/workspaces/:workspace_id/members/:user_id", handler.RemoveMember)
}
