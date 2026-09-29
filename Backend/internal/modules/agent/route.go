package agent

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.GET("/sandboxes/:sandbox_id/agent/status", handler.Status)
	router.POST("/sandboxes/:sandbox_id/agent/tasks", handler.RunTask)
}
