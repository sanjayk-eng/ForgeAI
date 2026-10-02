package project

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.POST("/organizations/:organization_id/projects", handler.Create)
	router.GET("/organizations/:organization_id/projects", handler.List)
	router.GET("/organizations/:organization_id/projects/:slug", handler.GetBySlug)
	router.GET("/projects/:project_id", handler.Get)
	router.PATCH("/projects/:project_id", handler.Update)
	router.PUT("/projects/:project_id", handler.Update)
	router.DELETE("/projects/:project_id", handler.Delete)
}
