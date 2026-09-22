package workspace

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.POST("/workspaces", handler.CreateWorkspace)
	router.GET("/workspaces", handler.ListWorkspaces)
	router.GET("/workspaces/:workspace_id", handler.GetWorkspace)
	router.PATCH("/workspaces/:workspace_id", handler.UpdateWorkspace)
}
