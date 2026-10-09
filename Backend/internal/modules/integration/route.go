package integration

import "github.com/gin-gonic/gin"

func RegisterRoutes(protected, public gin.IRouter, handler *Handler) {
	protected.GET("/organizations/:organization_id/integrations", handler.List)
	protected.GET("/organizations/:organization_id/integrations/github", handler.List)
	protected.POST("/organizations/:organization_id/projects/:project_id/github-export", handler.CreateExportRepository)
	protected.GET("/organizations/:organization_id/projects/:project_id/github-export", handler.GetExportRepository)
	protected.GET("/integrations/github/install", handler.Install)
	protected.DELETE("/organizations/:organization_id/integrations/github/:integration_id", handler.Disconnect)
	public.GET("/integrations/github/callback", handler.Callback)
	public.POST("/integrations/github/webhook", handler.Webhook)
}
