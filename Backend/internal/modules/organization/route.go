package organization

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	// Organizations
	router.POST("/organizations", handler.Create)
	router.GET("/organizations", handler.List)
	router.GET("/organizations/:organization_id", handler.Get)
	router.PATCH("/organizations/:organization_id", handler.Update)
	router.PUT("/organizations/:organization_id", handler.Update)
	router.DELETE("/organizations/:organization_id", handler.Delete)

	// Members
	router.GET("/organizations/:organization_id/members", handler.ListMembers)
	router.PATCH("/organizations/:organization_id/members/:user_id/role", handler.UpdateMemberRole)
	router.DELETE("/organizations/:organization_id/members/:user_id", handler.RemoveMember)
}
