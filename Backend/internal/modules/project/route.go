package project

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.POST("/projects/repository/resolve", handler.ResolveRepository)
	router.GET("/workspaces/:workspace_id/github/repositories", handler.ListGitHubRepositories)
	router.POST("/workspaces/:workspace_id/github/repositories/import", handler.ImportGitHubRepositories)
	router.POST("/workspaces/:workspace_id/projects", handler.Create)
	router.GET("/workspaces/:workspace_id/projects", handler.List)
	router.GET("/workspaces/:workspace_id/projects/slug/:slug", handler.GetBySlug)
	router.GET("/projects/:project_id", handler.Get)
	router.PATCH("/projects/:project_id", handler.Update)
	router.PATCH("/projects/:project_id/repository/branch", handler.UpdateRepositoryBranch)
	router.POST("/projects/:project_id/repository", handler.ConnectRepository)
	router.POST("/projects/:project_id/repository/sync", handler.Sync)
	router.POST("/workspaces/:workspace_id/projects/sync", handler.SyncWorkspaceProjects)
}
