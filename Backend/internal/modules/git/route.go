package git

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.GET("/sandboxes/:sandbox_id/git/status", handler.Status)
	router.GET("/sandboxes/:sandbox_id/git/diff", handler.Diff)
	router.POST("/sandboxes/:sandbox_id/git/commit", handler.Commit)
	router.POST("/sandboxes/:sandbox_id/git/revert", handler.Revert)
	router.POST("/sandboxes/:sandbox_id/git/push", handler.Push)
}
